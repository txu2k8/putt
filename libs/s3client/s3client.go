package s3client

import (
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"pzatest/libs/retry"
	"pzatest/libs/retry/strategy"
	"pzatest/libs/testErr"
	"pzatest/libs/utils"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/op/go-logging"
	"github.com/schollz/progressbar/v3"
)

/*
Base Function:
1. 查看S3中包含的bucket
2. bucket中的文件/文件夹
3. bucket的删除
4. bucket的创建
5. bucket的文件上传
6. bucket的文件下载
7. bucket的文件删除
*/

// define const for size unit
const (
	KB = 1024
	MB = 1024 * KB
	GB = 1024 * MB
	TB = 1024 * GB
)

var (
	logger             = logging.MustGetLogger("test")
	cachedS3Clients    = make(map[string]*s3.S3, 10)
	cacheS3ClientsSync sync.RWMutex
	s3CollectionLimit  = make(chan int, 5)
)

// S3Config the configuration of s3 client
type S3Config struct {
	Endpoint     string
	AccessID     string
	AccessSecret string
	Bucket       string
	Prefix       string
}

// A Object provides details of an S3 object
type Object struct {
	Bucket         string
	Key            string
	Encrypted      bool
	EncryptionType string
}

// An ErrObject provides details of the error occurred retrieving
// an object's status.
type ErrObject struct {
	Bucket string
	Key    string
	Error  error
}

// A Bucket provides details about a bucket and its objects
type Bucket struct {
	Owner        string
	Name         string
	CreationDate time.Time
	Region       string
	Objects      []Object
	Error        error
	ErrObjects   []ErrObject
}

type sortalbeBuckets []*Bucket

// EndlessReader is an io.Reader that will always return
// that bytes have been read.
type endlessReader struct{}

// progressWriter tracks the download progress of a file from S3 to a file
// 2020/03/31 13:54:19 File size:1199 downloaded:1199 percentage:100.00%
type progressWriter struct {
	written int64
	writer  io.WriterAt
	size    int64
	fname   string
}

// progressReader tracks the upload progress of a file to S3
// 2020/03/31 13:54:18 total read:1199    progress:100%
type progressReader struct {
	fp    *os.File
	size  int64
	read  int64
	fname string
}

// WriteAt ...
func (pw *progressWriter) WriteAt(p []byte, off int64) (int, error) {
	atomic.AddInt64(&pw.written, int64(len(p)))
	progress := int(float32(pw.written*100) / float32(pw.size))
	// log.Printf("File size:%d downloaded:%d progress:%.2f%%\n", pw.size, pw.written, progress)
	bar := progressbar.New(100)
	bar.Describe(fmt.Sprintf("File:%s, Size:%d, Downloaded:%d, Progress -", pw.fname, pw.size, pw.written))
	bar.Set(progress)
	if progress >= 100 {
		fmt.Println()
	}

	return pw.writer.WriteAt(p, off)
}

// Read ...
func (r *progressReader) Read(p []byte) (int, error) {
	return r.fp.Read(p)
}

// ReadAt ...
func (r *progressReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.fp.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	// Got the length have read( or means has uploaded)
	atomic.AddInt64(&r.read, int64(n))

	// I have no idea why the read length need to be div 2,
	// maybe the request read once when Sign and actually send call ReadAt again
	// It works for me
	// log.Printf("file:%s read:%d  progress:%d%%\n", r.fname, r.read/2, int(float32(r.read*100/2)/float32(r.size)))
	progress := int(float32(r.read*100/2) / float32(r.size))
	bar := progressbar.New(100)
	bar.Describe(fmt.Sprintf("File:%s, Size:%d, Read:%d, Progress -", r.fname, r.size, r.read/2))
	bar.Set(progress)
	if progress >= 100 {
		fmt.Println()
	}

	return n, err
}

// Seek ...
func (r *progressReader) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}

// Read will report that it has read len(p) bytes in p.
// The content in the []byte will be unmodified.
// This will never return an error.
func (e endlessReader) Read(p []byte) (int, error) {
	return len(p), nil
}

func parseFilename(keyString string) (filename string) {
	ss := strings.Split(keyString, "/")
	s := ss[len(ss)-1]
	return s
}

func byteCountDecimal(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

func getFileSize(svc *s3.S3, bucket string, prefix string) (filesize int64, error error) {
	params := &s3.HeadObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(prefix),
	}

	resp, err := svc.HeadObject(params)
	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}

// NewSession ...
func NewSession(endpoint string, accessID string, accessSecret string) *session.Session {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Configure to use Minio Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessID, accessSecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-west-2"),
		DisableSSL:       aws.Bool(false),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       client,
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		logger.Fatal(err)
	}
	return newSession
}

// NewS3Client ...
func NewS3Client(endpoint string, accessID string, accessSecret string) *s3.S3 {
	// The purpose of the two judgments is to avoid locking each time.
	if s3Client, hit := cachedS3Clients[accessID]; hit {
		return s3Client
	}
	cacheS3ClientsSync.RLock()
	defer cacheS3ClientsSync.RUnlock()
	if s3Client, hit := cachedS3Clients[accessID]; hit {
		return s3Client
	}

	newSession := NewSession(endpoint, accessID, accessSecret)
	s3Client := s3.New(newSession, newSession.Config)

	return s3Client
}

// GetObject :DownloadFile ...
func GetObject(config *S3Config, s3Bucket string, s3Path string, localFilePath string) error {
	logger.Infof("Try to download file from s3bucket: %s, path: %s, local: %s\n", s3Bucket, s3Path, localFilePath)
	s3Client := NewS3Client(config.Endpoint, config.AccessID, config.AccessSecret)

	file, err := os.Create(localFilePath)
	if err != nil {
		logger.Error("Failed to create file", err)
		return testErr.ErrCreateLocalFile
	}
	defer file.Close()

	Bucket := aws.String(s3Bucket)
	Path := aws.String(s3Path)

	downloader := s3manager.NewDownloader(session.Must(session.NewSession(&s3Client.Config)))

	numBytes, err := downloader.Download(
		file,
		&s3.GetObjectInput{
			Bucket: Bucket,
			Key:    Path,
		})
	if err != nil {
		logger.Errorf("Failed to download file", err)
		return testErr.ErrDownloadFile
	}
	logger.Info("Downloaded file ", file.Name(), " ", numBytes, "bytes")
	return nil
}

// UploadFile ...
func UploadFile(sess *session.Session, s3Bucket string, localFilePath string) error {
	file, err := os.Open(localFilePath)
	if err != nil {
		logger.Errorf("ERROR:", err)
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("ERROR:", err)
		return err
	}

	logger.Infof("Starting upload(size:%s):%s", byteCountDecimal(fileInfo.Size()), localFilePath)
	timeStart := time.Now()
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * MB
		u.LeavePartsOnError = true
	})

	_, sBase := path.Split(localFilePath)
	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(sBase),
		Body:   file,
	})

	if err != nil {
		logger.Errorf("ERROR:", err)
		return err
	}
	timeEnd := time.Now()
	timeDelta := timeEnd.Sub(timeStart)
	logger.Infof("Upload PASS: %s (Elapsed:%s)", output.Location, timeDelta)
	return nil
}

// UploadFileRetry ...
func UploadFileRetry(sess *session.Session, s3Bucket string, localFilePath string) error {
	action := func(attempt uint) error {
		return UploadFile(sess, s3Bucket, localFilePath)
	}
	err := retry.Retry(
		action,
		strategy.Limit(25),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Second)),
	)
	return err
}

// UploadFileWithProcess ...
func UploadFileWithProcess(sess *session.Session, s3Bucket string, localFilePath string) error {
	file, err := os.Open(localFilePath)
	if err != nil {
		logger.Errorf("ERROR:", err)
		return err
	}

	fileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("ERROR:", err)
		return err
	}

	reader := &progressReader{
		fp:    file,
		size:  fileInfo.Size(),
		fname: fileInfo.Name(),
	}

	logger.Infof("Starting upload(size:%s):%s", byteCountDecimal(reader.size), localFilePath)
	timeStart := time.Now()
	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * MB
		u.LeavePartsOnError = true
	})

	_, sBase := path.Split(localFilePath)
	output, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(sBase),
		Body:   reader,
	})

	if err != nil {
		logger.Errorf("ERROR:", err)
		return err
	}
	timeEnd := time.Now()
	timeDelta := timeEnd.Sub(timeStart)
	logger.Infof("Upload PASS: %s (Elapsed:%s)", output.Location, timeDelta)
	return nil
}

// UploadFileWithProcessRetry ...
func UploadFileWithProcessRetry(sess *session.Session, s3Bucket string, localFilePath string) error {
	action := func(attempt uint) error {
		return UploadFileWithProcess(sess, s3Bucket, localFilePath)
	}
	err := retry.Retry(
		action,
		strategy.Limit(5),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Microsecond)),
	)
	return err
}

// DownloadFile ...
func DownloadFile(svc *s3.S3, s3Bucket string, s3Path string, localDir string) error {
	fullPath := *svc.Config.Endpoint + "/" + s3Bucket + "/" + s3Path
	filename := parseFilename(s3Path)
	localFilePath := path.Join(localDir, filename)
	logger.Infof("Starting download file:%s", fullPath)
	// tempfile, err := ioutil.TempFile(localDir, "download_*_"+filename)
	tempfile, err := os.OpenFile(localFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer tempfile.Close()
	params := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Path),
	}

	tempfileName := tempfile.Name()
	downloader := s3manager.NewDownloader(session.Must(session.NewSession(&svc.Config)))
	if n, err := downloader.Download(tempfile, params); n != 0 && err != nil {
		logger.Errorf("Download failed! Deleting tempfile: %s", tempfileName)
		os.Remove(tempfileName)
		return err
	}

	logger.Info("Download PASS:", fullPath)
	return nil
}

// DownloadFileRetry ...
func DownloadFileRetry(svc *s3.S3, s3Bucket string, s3Path string, localDir string) error {
	action := func(attempt uint) error {
		return DownloadFile(svc, s3Bucket, s3Path, localDir)
	}
	err := retry.Retry(
		action,
		strategy.Limit(25),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Second)),
	)
	return err
}

// DownloadFileWithProcess ...
func DownloadFileWithProcess(svc *s3.S3, s3Bucket string, s3Path string, localDir string) error {
	fullPath := *svc.Config.Endpoint + "/" + s3Bucket + "/" + s3Path
	filename := parseFilename(s3Path)
	localFilePath := path.Join(localDir, filename)
	size, err := getFileSize(svc, s3Bucket, s3Path)
	if err != nil {
		return err
	}

	logger.Infof("Starting download(size:%s):%s", byteCountDecimal(size), fullPath)
	// tempfile, err := ioutil.TempFile(localDir, "download_*_"+filename)
	tempfile, err := os.OpenFile(localFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0666)
	if err != nil {
		return err
	}
	defer tempfile.Close()
	writer := &progressWriter{writer: tempfile, size: size, written: 0, fname: filename}
	params := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Path),
	}

	tempfileName := tempfile.Name()
	downloader := s3manager.NewDownloader(session.Must(session.NewSession(&svc.Config)))
	if n, err := downloader.Download(writer, params); n != 0 && err != nil {
		logger.Errorf("Download failed! Deleting tempfile: %s", tempfileName)
		os.Remove(tempfileName)
		return err
	}

	logger.Info("Download PASS:", fullPath)
	return nil
}

// DownloadFileWithProcessRetry ...
func DownloadFileWithProcessRetry(svc *s3.S3, s3Bucket string, s3Path string, localDir string) error {
	action := func(attempt uint) error {
		return DownloadFileWithProcess(svc, s3Bucket, s3Path, localDir)
	}
	err := retry.Retry(
		action,
		strategy.Limit(25),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Second)),
	)
	return err
}

// ListBuckets ...
func ListBuckets(svc *s3.S3) error {
	buckets, err := getBuckets(svc)
	if err != nil {
		logger.Errorf("ListBuckets failed, %v", err)
		return err
	}
	logger.Infof("Bucket list:\n%s", utils.Prettify(buckets))
	return err
}

// ListBucketsRetry ...
func ListBucketsRetry(svc *s3.S3) error {
	action := func(attempt uint) error {
		return ListBuckets(svc)
	}
	err := retry.Retry(
		action,
		strategy.Limit(25),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Second)),
	)
	return err
}

// ListBucketFiles ...
func ListBucketFiles(svc *s3.S3, s3Bucket string) error {
	objs, errObjs, err := getBucketObjects(svc, s3Bucket)
	if err != nil {
		logger.Errorf("ListBucketFiles failed, %v", err)
		return err
	}
	fileArry := []string{}
	errFileArry := []string{}
	for _, obj := range objs {
		fileArry = append(fileArry, obj.Key)
	}
	for _, errObj := range errObjs {
		errFileArry = append(errFileArry, errObj.Key)
	}
	logger.Infof("Bucket Files:\n%s", utils.Prettify(fileArry))
	if len(errFileArry) > 0 {
		logger.Warningf("Bucket Err Files:\n%s", utils.Prettify(errFileArry))
	}
	return err
}

// ListBucketFilesRetry ...
func ListBucketFilesRetry(svc *s3.S3, s3Bucket string) error {
	action := func(attempt uint) error {
		return ListBucketFiles(svc, s3Bucket)
	}
	err := retry.Retry(
		action,
		strategy.Limit(25),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Second)),
	)
	return err
}

// DeleteBucketFile ...
func DeleteBucketFile(svc *s3.S3, s3Bucket string, s3Path string) error {
	fullPath := *svc.Config.Endpoint + "/" + s3Bucket + "/" + s3Path
	deleteInput := s3.DeleteObjectInput{
		Bucket: &s3Bucket,
		Key:    aws.String(s3Path),
	}
	_, err := svc.DeleteObject(&deleteInput)
	if err != nil {
		logger.Errorf("Unable to delete object %s, %v", fullPath, err)
		return err
	}

	err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Path),
	})

	if err != nil {
		logger.Errorf("Deleted object still exist:%s, %v", fullPath, err)
		return err
	}

	logger.Info("Deleted PASS:", fullPath)
	return nil
}

// DeleteBucketFileRetry ...
func DeleteBucketFileRetry(svc *s3.S3, s3Bucket string, s3Path string) error {
	action := func(attempt uint) error {
		return DeleteBucketFile(svc, s3Bucket, s3Path)
	}
	err := retry.Retry(
		action,
		strategy.Limit(25),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(30*time.Second)),
	)
	return err
}

// CreateBucket returns a bucket created for the tests.
func CreateBucket(svc *s3.S3, bucketName string) error {
	logger.Info("Setup: Creating test bucket,", bucketName)
	_, err := svc.CreateBucket(&s3.CreateBucketInput{Bucket: &bucketName})
	if err != nil {
		return fmt.Errorf("failed to create bucket %s, %v", bucketName, err)
	}

	fmt.Println("Setup: Waiting for bucket to exist,", bucketName)
	err = svc.WaitUntilBucketExists(&s3.HeadBucketInput{Bucket: &bucketName})
	if err != nil {
		return fmt.Errorf("failed waiting for bucket %s to be created, %v", bucketName, err)
	}

	return nil
}

// DeleteBucket ...
func DeleteBucket(svc *s3.S3, bucket string) error {
	bucketName := &bucket

	objs, err := svc.ListObjects(&s3.ListObjectsInput{Bucket: bucketName})
	if err != nil {
		return fmt.Errorf("failed to list bucket %q objects, %v", *bucketName, err)
	}

	for _, o := range objs.Contents {
		svc.DeleteObject(&s3.DeleteObjectInput{Bucket: bucketName, Key: o.Key})
	}

	uploads, err := svc.ListMultipartUploads(&s3.ListMultipartUploadsInput{Bucket: bucketName})
	if err != nil {
		return fmt.Errorf("failed to list bucket %q multipart objects, %v", *bucketName, err)
	}

	for _, u := range uploads.Uploads {
		svc.AbortMultipartUpload(&s3.AbortMultipartUploadInput{
			Bucket:   bucketName,
			Key:      u.Key,
			UploadId: u.UploadId,
		})
	}

	_, err = svc.DeleteBucket(&s3.DeleteBucketInput{Bucket: bucketName})
	if err != nil {
		return fmt.Errorf("failed to delete bucket %q, %v", *bucketName, err)
	}

	return nil
}

// ============= ListBucketObjectsConcurrently =============
func sortBuckets(buckets []*Bucket) {
	s := sortalbeBuckets(buckets)
	sort.Sort(s)
}

func (s sortalbeBuckets) Len() int      { return len(s) }
func (s sortalbeBuckets) Swap(a, b int) { s[a], s[b] = s[b], s[a] }
func (s sortalbeBuckets) Less(a, b int) bool {
	if s[a].Owner == s[b].Owner && s[a].Name < s[b].Name {
		return true
	}

	if s[a].Owner < s[b].Owner {
		return true
	}

	return false
}

func (b *Bucket) encryptedObjects() []Object {
	encObjs := []Object{}
	for _, obj := range b.Objects {
		if obj.Encrypted {
			encObjs = append(encObjs, obj)
		}
	}
	return encObjs
}

// getBuckets ...
func getBuckets(svc *s3.S3) ([]*Bucket, error) {
	res, err := svc.ListBuckets(&s3.ListBucketsInput{})
	if err != nil {
		return nil, err
	}

	buckets := make([]*Bucket, len(res.Buckets))
	for i, b := range res.Buckets {
		buckets[i] = &Bucket{
			Name:         *b.Name,
			CreationDate: *b.CreationDate,
			Region:       "us-west-2",
		}
		// locRes, err := svc.GetBucketLocation(&s3.GetBucketLocationInput{
		// 	Bucket: b.Name,
		// })
		// if err != nil {
		// 	buckets[i].Error = err
		// 	continue
		// }

		// if locRes.LocationConstraint == nil {
		// 	buckets[i].Region = "us-west-2"
		// } else {
		// 	buckets[i].Region = *locRes.LocationConstraint
		// }
	}

	return buckets, nil
}

// getBucketObjects : return objs
func getBucketObjects(svc *s3.S3, bucket string) ([]Object, []ErrObject, error) {
	logger.Infof("getBucketObjects:%s", bucket)
	listRes, err := svc.ListObjects(&s3.ListObjectsInput{
		Bucket: &bucket,
	})
	if err != nil {
		return nil, nil, err
	}

	objs := make([]Object, 0, len(listRes.Contents))
	errObjs := []ErrObject{}
	for _, listObj := range listRes.Contents {
		objData, err := svc.HeadObject(&s3.HeadObjectInput{
			Bucket: &bucket,
			Key:    listObj.Key,
		})

		if err != nil {
			errObjs = append(errObjs, ErrObject{Bucket: bucket, Key: *listObj.Key, Error: err})
			continue
		}

		obj := Object{Bucket: bucket, Key: *listObj.Key}
		logger.Debug(obj.Bucket + ":" + obj.Key)
		if objData.ServerSideEncryption != nil {
			obj.Encrypted = true
			obj.EncryptionType = *objData.ServerSideEncryption
		}

		objs = append(objs, obj)
	}

	return objs, errObjs, nil
}

// get bucket details: return bucket:objs
func bucketDetails(svc *s3.S3, bucket *Bucket) {
	objs, errObjs, err := getBucketObjects(svc, bucket.Name)
	if err != nil {
		bucket.Error = err
	} else {
		bucket.Objects = objs
		bucket.ErrObjects = errObjs
	}
}

// getAccountBucketsDetails: return Account -> buckets:objects
func getAccountBucketsDetails(sess *session.Session, bucketCh chan<- *Bucket, owner string) error {
	svc := s3.New(sess)
	buckets, err := getBuckets(svc)
	if err != nil {
		return fmt.Errorf("failed to list buckets, %v", err)
	}
	for _, bucket := range buckets {
		bucket.Owner = owner
		if bucket.Error != nil {
			continue
		}

		bckSvc := s3.New(sess, &aws.Config{
			Region:      aws.String(bucket.Region),
			Credentials: svc.Config.Credentials,
		})
		bucketDetails(bckSvc, bucket)
		bucketCh <- bucket
	}

	return nil
}

// ListBucketObjectsConcurrently ...
func ListBucketObjectsConcurrently(svc *s3.S3, bucket string, accounts []string) {
	// Spin off a worker for each account to retrieve that account's
	bucketCh := make(chan *Bucket, 5)
	var wg sync.WaitGroup
	for _, acc := range accounts {
		wg.Add(1)
		go func(acc string) {
			defer wg.Done()

			sess, err := session.NewSessionWithOptions(session.Options{
				Config:  svc.Config,
				Profile: acc,
			})
			if err != nil {
				logger.Errorf("failed to create session for account, %s, %v\n", acc, err)
				return
			}
			if err = getAccountBucketsDetails(sess, bucketCh, acc); err != nil {
				logger.Errorf("failed to get account %s's bucket info, %v\n", acc, err)
				return
			}
		}(acc)
	}
	// Spin off a goroutine which will wait until all account buckets have
	// been collected and added to the bucketCh. Close the bucketCh so the
	// for range below will exit once all bucket info is printed.
	go func() {
		wg.Wait()
		close(bucketCh)
	}()

	// Receive from the bucket channel printing the information for each bucket
	//  to the console when the bucketCh channel is drained.
	buckets := []*Bucket{}
	for b := range bucketCh {
		buckets = append(buckets, b)
	}

	sortBuckets(buckets)
	for _, b := range buckets {
		if b.Error != nil {
			fmt.Printf("Bucket %s, owned by: %s, failed: %v\n", b.Name, b.Owner, b.Error)
			continue
		}

		encObjs := b.encryptedObjects()
		logger.Infof("Bucket: %s, owned by: %s, total objects: %d, failed objects: %d, encrypted objects: %d\n",
			b.Name, b.Owner, len(b.Objects), len(b.ErrObjects), len(encObjs))
		if len(encObjs) > 0 {
			for _, encObj := range encObjs {
				logger.Infof("\t%s %s:%s/%s\n", encObj.EncryptionType, b.Region, b.Name, encObj.Key)
			}
		}
	}
}
