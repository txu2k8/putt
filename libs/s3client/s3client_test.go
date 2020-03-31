package s3client

import (
	"os"
	"testing"
)

func TestS3UploadFileWithProcess(t *testing.T) {
	bucket := "vset1_s3bucket_17_34"
	localFilePath := "./a.txt"
	file, err := os.OpenFile(localFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Panicf("ERROR:", err)
	}
	file.WriteString("======Test S3 UploadFile Download File========")
	session := newSession("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	UploadFileWithProcess(session, bucket, localFilePath)
}

func TestS3DownloadFileWithProcess(t *testing.T) {
	bucket := "vset1_s3bucket_17_34"
	s3Path := "a.txt"
	svc := newS3Client("https://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	DownloadFileWithProcess(svc, bucket, s3Path, "./")
}
