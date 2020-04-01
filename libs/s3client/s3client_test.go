package s3client

import (
	"fmt"
	_ "gtest/testinit"
	"os"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
)

func TestS3UploadFileWithProcess(t *testing.T) {
	bucket := "vset1_s3bucket_17_34"
	localFilePath := "./test.log"
	file, err := os.OpenFile(localFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Panicf("ERROR:", err)
	}
	file.WriteString("======Test S3 UploadFile Download File========\n")
	session := NewSession("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	UploadFileWithProcess(session, bucket, localFilePath)
}

func TestS3DownloadFileWithProcess(t *testing.T) {
	bucket := "vset1_s3bucket_17_34"
	s3Path := "test.log"
	svc := NewS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	DownloadFileWithProcess(svc, bucket, s3Path, "./")
}

func TestS3ListBucketObjectsConcurrently(t *testing.T) {
	bucket := "vset1_s3bucket_17_34"
	accounts := []string{"vset1_s3user"}
	svc := NewS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	bs, _ := ListBuckets(svc)
	for _, b := range bs {
		logger.Info(b.Name + "," + b.Region)
	}
	ListBucketObjectsConcurrently(svc, bucket, accounts)
}

func TestS3DeleteBucket(t *testing.T) {
	bucketPrefix := "vset1_s3bucket_17_34"
	svc := NewS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	bs, _ := ListBuckets(svc)
	for _, b := range bs {
		logger.Info(b.Name + "," + b.Region)
		bucket := aws.StringValue(&b.Name)
		if !strings.HasPrefix(bucket, bucketPrefix) {
			continue
		}

		logger.Infof("Delete bucket %q? [y/N]: ", bucket)
		var v string
		if _, err := fmt.Scanln(&v); err != nil || !(v == "Y" || v == "y") {
			logger.Info("Skipping")
			continue
		}

		logger.Info("Deleting")
		if err := DeleteBucket(svc, bucket); err != nil {
			logger.Panicf("failed to delete bucket %q, %v", bucket, err)
		}
	}
}

func TestS3CreateBucket(t *testing.T) {
	bucketName := "test_bucket"
	svc := NewS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")
	CreateBucket(svc, bucketName)
	bs, _ := ListBuckets(svc)
	for _, b := range bs {
		logger.Info(b.Name + "," + b.Region)
	}
}
