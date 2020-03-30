package s3client

import (
	"crypto/tls"
	"fmt"
	"index/config"
	"index/models"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/defaults"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"learngo/src/github.com/golang/glog"
)

var (
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

// TenantS3Client the s3 client with tenant info
type TenantS3Client struct {
	S3     *s3.S3
	Tenant *models.Tenant
}

// ListFilesV2 list the files in point prefix
func (client *TenantS3Client) ListFilesV2(bucket, prefix string, lastKey *string, maxKeys int64) ([]string, error) {
	s3CollectionLimit <- 1
	defer func() {
		<-s3CollectionLimit
	}()
	var fs []string
	dataChan := make(chan []string)
	errorChan := make(chan error)
	go func(chan []string, chan error) {
		objects, err := client.S3.ListObjectsV2(&s3.ListObjectsV2Input{
			Bucket:     aws.String(bucket),
			Prefix:     aws.String(prefix),
			MaxKeys:    aws.Int64(maxKeys),
			StartAfter: lastKey,
		})
		if err != nil {
			glog.Errorf("Failed to list data from %s/%s, %s\n", bucket, prefix, err.Error())
			errorChan <- err
			return
		}
		glog.V(7).Infof("there are %d objects in path %s", len(objects.Contents), prefix)
		for _, object := range objects.Contents {
			fs = append(fs, *object.Key)
		}
		dataChan <- fs
	}(dataChan, errorChan)
	select {
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("list files timeout")
	case fs := <-dataChan:
		return fs, nil
	case err := <-errorChan:
		return nil, err
	}
}

// ListFiles list the files in point prefix
func (client *TenantS3Client) ListFiles(bucket string, prefix string) ([]string, error) {
	var fs []string
	dataChan := make(chan []string)
	errorChan := make(chan error)
	go func(chan []string, chan error) {
		objects, err := client.S3.ListObjects(&s3.ListObjectsInput{
			Bucket:  aws.String(bucket),
			Prefix:  aws.String(prefix),
			MaxKeys: aws.Int64(config.Config.Master.WatchPageSize),
		})
		if err != nil {
			glog.Errorf("Failed to list data from %s/%s, %s\n", bucket, prefix, err.Error())
			errorChan <- err
			return
		}
		glog.V(7).Infof("there are %d objects in path %s", len(objects.Contents), prefix)
		for _, object := range objects.Contents {
			fs = append(fs, *object.Key)
		}
		dataChan <- fs
	}(dataChan, errorChan)
	select {
	case <-time.After(2 * time.Minute):
		return nil, fmt.Errorf("list files timeout")
	case fs := <-dataChan:
		return fs, nil
	case err := <-errorChan:
		return nil, err
	}
}

// DeleteFile delete the specific file
func (client *TenantS3Client) DeleteFile(file *string) error {
	_, err := client.S3.DeleteObject(&s3.DeleteObjectInput{
		Bucket: &client.Tenant.DefaultBucket,
		Key:    file,
	})
	if err != nil {
		glog.Errorf("%s delete %v  failed %v\n", client.Tenant.Name, *file, err)
		return err
	}
	glog.V(4).Infof("%s deleted file %v", client.Tenant.Name, *file)
	return nil
}

// DeleteFiles delete the specific files
func (client *TenantS3Client) DeleteFiles(files []*string) error {
	var objects []*s3.ObjectIdentifier
	for _, file := range files {
		objects = append(objects, &s3.ObjectIdentifier{Key: file})
	}
	out, err := client.S3.DeleteObjects(&s3.DeleteObjectsInput{
		Bucket: &client.Tenant.DefaultBucket,
		Delete: &s3.Delete{
			Objects: objects,
		},
	})
	if err != nil {
		join := func(points []*string, sep string) string {
			s := ""
			for _, f := range files {
				s += *f + sep
			}
			return s
		}
		glog.Errorf("delete %v failed %v\n", join(files, " "), err)
		return err
	}
	glog.V(4).Infof("deleted files:%v", out)
	return nil
}

// GetObject ...
func GetObject(config *S3Config, s3Bucket string, s3Path string, localFilePath string) error {
	glog.V(4).Infof("Try to download file from s3bucket: %s, path: %s, local: %s\n", s3Bucket, s3Path, localFilePath)
	s3Client := newS3Client(config.Endpoint, config.AccessID, config.AccessSecret)

	file, err := os.Create(localFilePath)
	if err != nil {
		glog.Error("Failed to create file", err)
		return models.ErrCreateLocalFile
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
		glog.Errorln("Failed to download file", err)
		return models.ErrDownloadFile
	}
	glog.V(4).Info("Downloaded file ", file.Name(), " ", numBytes, "bytes")
	return nil
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
		glog.Error(err)
	}
	s3Client := s3.New(newSession, s3Config)

	return s3Client
}

// NewTenantS3Client build s3 client with tenant info
func NewTenantS3Client(tenant *models.Tenant, endpoint string) *TenantS3Client {

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	// Configure to use s3 Server
	s3Config := defaults.Config()
	s3Config.Credentials = credentials.NewStaticCredentials(tenant.Access.Key, tenant.Access.Secret, "")
	s3Config.Endpoint = aws.String(endpoint)
	s3Config.Region = aws.String("us-east-1")
	s3Config.DisableSSL = aws.Bool(true)
	s3Config.S3ForcePathStyle = aws.Bool(true)
	s3Config.HTTPClient = client
	s3Config.LogLevel = aws.LogLevel(aws.LogDebugWithRequestErrors)
	newSession, err := session.NewSession(s3Config)
	if err != nil {
		fmt.Println(err)
	}
	s3Client := s3.New(newSession, s3Config)

	return &TenantS3Client{S3: s3Client, Tenant: tenant}
}
