package main

import (
	"fmt"
	_ "gtest/config"
	s3client "gtest/libs/s3client"
	_ "gtest/testinit"
	"os"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

func testLogging() {
	logger.Info("------info")
	logger.Notice("------notice")
	logger.Warning("------warning")
	logger.Error("------err")
	logger.Critical("------crit")
}

func testS3Upload() {
	bucket := "vset1_s3bucket_17_34"
	localFilePath := "./test.log"
	file, err := os.OpenFile(localFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Panicf("ERROR:", err)
	}
	file.WriteString("======Test S3 UploadFile Download File========\n")
	session := s3client.NewSession("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	s3client.UploadFileWithProcess(session, bucket, localFilePath)
}

func testS3Download() {
	bucket := "vset1_s3bucket_17_34"
	s3Path := "test.log"
	svc := s3client.NewS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	s3client.DownloadFileWithProcess(svc, bucket, s3Path, "./")
}

func testS3ListObject() {
	bucket := "vset1_s3bucket_17_34"
	accounts := []string{"vset1_s3user"}
	svc := s3client.NewS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	fmt.Println(s3client.ListBuckets(svc))
	s3client.ListObjectsConcurrently(svc, bucket, accounts)
}

func main() {
	// testLogging()
	// testS3Upload()
	// testS3Download()
	testS3ListObject()
}
