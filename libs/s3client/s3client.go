package s3client

import (
	"crypto/tls"
	"fmt"
	"gtest/libs/testErr"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/op/go-logging"
)

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

// progressWriter tracks the download progress of a file from S3 to a file
// as the writeAt method is called, the byte size is added to the written total,
// and then a log is printed of the written percentage from the total size
// it looks like this on the command line:
//  2019/02/22 12:59:15 File size:35943530 downloaded:16360 percentage:0%
//  2019/02/22 12:59:15 File size:35943530 downloaded:16988 percentage:0%
//  2019/02/22 12:59:15 File size:35943530 downloaded:33348 percentage:0%
type progressWriter struct {
	written int64
	writer  io.WriterAt
	size    int64
}

// CustomReader ...
type CustomReader struct {
	fp   *os.File
	size int64
	read int64
}

func (pw *progressWriter) WriteAt(p []byte, off int64) (int, error) {
	atomic.AddInt64(&pw.written, int64(len(p)))

	percentageDownloaded := float32(pw.written*100) / float32(pw.size)

	fmt.Printf("File size:%d downloaded:%d percentage:%.2f%%\r", pw.size, pw.written, percentageDownloaded)

	return pw.writer.WriteAt(p, off)
}

func byteCountDecimal(b int64) string {
	const unit = 1000
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

	fmt.Println(params)
	resp, err := svc.HeadObject(params)
	if err != nil {
		return 0, err
	}

	return *resp.ContentLength, nil
}

func parseFilename(keyString string) (filename string) {
	ss := strings.Split(keyString, "/")
	s := ss[len(ss)-1]
	return s
}

func (r *CustomReader) Read(p []byte) (int, error) {
	return r.fp.Read(p)
}

// ReadAt ...
func (r *CustomReader) ReadAt(p []byte, off int64) (int, error) {
	n, err := r.fp.ReadAt(p, off)
	if err != nil {
		return n, err
	}

	// Got the length have read( or means has uploaded), and you can construct your message
	atomic.AddInt64(&r.read, int64(n))

	// I have no idea why the read length need to be div 2,
	// maybe the request read once when Sign and actually send call ReadAt again
	// It works for me
	log.Printf("total read:%d    progress:%d%%\n", r.read/2, int(float32(r.read*100/2)/float32(r.size)))

	return n, err
}

// Seek ...
func (r *CustomReader) Seek(offset int64, whence int) (int64, error) {
	return r.fp.Seek(offset, whence)
}

func newSession(endpoint string, accessID string, accessSecret string) *session.Session {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Configure to use Minio Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessID, accessSecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       client,
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		logger.Fatal(err)
	}
	return newSession
}

func newS3Client(endpoint string, accessID string, accessSecret string) *s3.S3 {
	// The purpose of the two judgments is to avoid locking each time.
	if s3Client, hit := cachedS3Clients[accessID]; hit {
		return s3Client
	}
	cacheS3ClientsSync.RLock()
	defer cacheS3ClientsSync.RUnlock()
	if s3Client, hit := cachedS3Clients[accessID]; hit {
		return s3Client
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Configure to use Minio Server
	s3Config := &aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessID, accessSecret, ""),
		Endpoint:         aws.String(endpoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true),
		HTTPClient:       client,
	}
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		logger.Error(err)
	}
	s3Client := s3.New(newSession, s3Config)

	return s3Client
}

// GetObject ...
func GetObject(config *S3Config, s3Bucket string, s3Path string, localFilePath string) error {
	logger.Infof("Try to download file from s3bucket: %s, path: %s, local: %s\n", s3Bucket, s3Path, localFilePath)
	s3Client := newS3Client(config.Endpoint, config.AccessID, config.AccessSecret)

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

// DownloadFileWithProcess ...
func DownloadFileWithProcess(svc *s3.S3, s3Bucket string, s3Path string, locairlDir string) bool {
	filename := parseFilename(s3Path)
	size, err := getFileSize(svc, s3Bucket, s3Path)
	if err != nil {
		panic(err)
	}

	logger.Info("Starting download, size:", byteCountDecimal(size))
	temp, err := ioutil.TempFile(locairlDir, "download_*_"+filename)
	if err != nil {
		panic(err)
	}
	defer temp.Close()
	writer := &progressWriter{writer: temp, size: size, written: 0}
	params := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Path),
	}

	tempfileName := temp.Name()
	downloader := s3manager.NewDownloader(session.Must(session.NewSession(&svc.Config)))
	if _, err := downloader.Download(writer, params); err != nil {
		logger.Errorf("Download failed! Deleting tempfile: %s", tempfileName)
		os.Remove(tempfileName)
		panic(err)
	}

	logger.Info("Download PASS: " + s3Bucket + "/" + s3Path)
	return true
}

// UploadFileWithProcess ...
func UploadFileWithProcess(sess *session.Session, s3Bucket string, localFilePath string) bool {
	file, err := os.Open(localFilePath)
	if err != nil {
		logger.Errorf("ERROR:", err)
		return false
	}

	fileInfo, err := file.Stat()
	if err != nil {
		logger.Errorf("ERROR:", err)
		return false
	}

	reader := &CustomReader{
		fp:   file,
		size: fileInfo.Size(),
	}

	uploader := s3manager.NewUploader(sess, func(u *s3manager.Uploader) {
		u.PartSize = 5 * 1024 * 1024
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
		return false
	}

	logger.Info("Upload PASS: " + output.Location)
	return true
}
