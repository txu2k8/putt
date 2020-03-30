package s3client

import (
	"crypto/tls"
	"fmt"
	"gtest/libs/testErr"
	"io"
	"io/ioutil"
	"net/http"
	"os"
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
		Region:           aws.String("us-west-2"),
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

// DownloadFile ...
func DownloadFile(svc *s3.S3, s3Bucket string, s3Path string, localFilePath string) bool {
	filename := parseFilename(s3Path)
	size, err := getFileSize(svc, s3Bucket, s3Path)
	if err != nil {
		panic(err)
	}

	logger.Info("Starting download, size:", byteCountDecimal(size))
	temp, err := ioutil.TempFile(localFilePath, "s3-download-tmp-")
	if err != nil {
		panic(err)
	}
	tempfileName := temp.Name()

	writer := &progressWriter{writer: temp, size: size, written: 0}
	params := &s3.GetObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(s3Path),
	}

	downloader := s3manager.NewDownloader(session.Must(session.NewSession(&svc.Config)))
	if _, err := downloader.Download(writer, params); err != nil {
		logger.Errorf("Download failed! Deleting tempfile: %s", tempfileName)
		os.Remove(tempfileName)
		panic(err)
	}

	if err := temp.Close(); err != nil {
		panic(err)
	}

	if err := os.Rename(temp.Name(), filename); err != nil {
		panic(err)
	}
	logger.Infof("File downloaded! Avaliable at:", filename)
	return true
}
