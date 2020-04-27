package main

import (
	"gtest/cmd"
	_ "gtest/config"
	"gtest/libs/retry"
	"gtest/libs/retry/backoff"
	"gtest/libs/retry/jitter"
	"gtest/libs/retry/strategy"
	s3client "gtest/libs/s3client"
	_ "gtest/testinit"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey"
)

var logger = logging.MustGetLogger("test")

func runCase(t *testing.T, testCase func(*testing.T)) {
	Before()
	defer After()

	testCase(t)
}

func SetUp() {

}

func TearDown() {

}

// TestRunSuite ...
func TestRunSuite(t *testing.T) {
	SetUp()
	defer TearDown()
	Convey("初始化", t, nil)

	runCase(t, TestRetry)
}

func testLogging(filePath string) error {
	logger.Infof("Open logFile: %s...", filePath)
	_, err := os.Open(filePath)
	if err == nil {
		logger.Info("------info")
		logger.Notice("------notice")
		logger.Warning("------warning")
		logger.Error("------err")
		logger.Critical("------crit")
	}
	return err
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

	bs, _ := s3client.ListBuckets(svc)
	for _, b := range bs {
		logger.Info(b.Name)
		logger.Warning(b.Region)
	}
	s3client.ListBucketObjectsConcurrently(svc, bucket, accounts)
}

func testRetry() bool {
	const logFilePath = "./test.log1"

	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))
	err := retry.Retry(func(attempt uint) error {
		return testLogging(logFilePath)
	},
		strategy.Limit(3),
		strategy.Wait(2*time.Second),
		strategy.BackoffWithJitter(
			backoff.BinaryExponential(10*time.Millisecond),
			jitter.Deviation(random, 0.5),
		),
	)

	if err != nil {
		log.Fatalf("Unable to open file %q with error %q", logFilePath, err)
	}
	return true
}

func TestRetry(t *testing.T) {
	Convey("Test Retry", t, func() {
		So(testRetry(), ShouldEqual, true)
	})
}

func main() {
	// testLogging()
	// testS3Upload()
	// testS3Download()
	// testS3ListObject()
	// utils.SleepProgressBar(2)
	// testRetry()
	cmd.Execute()
}
