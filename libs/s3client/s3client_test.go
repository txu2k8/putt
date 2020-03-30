package s3client

import (
	"testing"
)

func TestS3DownloadFile(t *testing.T) {
	bucket := "vset1_s3bucket_17_34"
	key := "a.txt"
	s3Client := newS3Client("http://10.25.119.86:443", "4CKG9PM8MG86LEOD2EPN", "s3OkruYUFuZ6xjskUjJuWU7dSxVcy6455o8xMEeJ")

	DownloadFile(s3Client, bucket, key, "./")
}
