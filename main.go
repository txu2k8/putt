package main

import (
	_ "gtest/config"
	s3client "gtest/libs/s3client"
	_ "gtest/testinit"
	"os"

	"github.com/op/go-logging"
	"gopkg.in/alecthomas/kingpin.v2"
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

	bs, _ := s3client.ListBuckets(svc)
	for _, b := range bs {
		logger.Info(b.Name)
		logger.Warning(b.Region)
	}
	s3client.ListBucketObjectsConcurrently(svc, bucket, accounts)
}

func baseApp() *kingpin.Application {
	app := kingpin.New("gtest", "Vizion Test Project")
	app.Flag("debug", "Enable debug mode.").Bool()
	// app.Arg("iteration", "total iteration").Int64()
	return app
}

func vizionStressCommand(app *kingpin.Application) *kingpin.CmdClause {
	stress := app.Command("stress", "Vizion Stress Test")
	stress.Arg("sys_user", "host login username").Default("root").String()
	stress.Arg("sys_pwd", "host login password").Default("password").String()
	stress.Arg("key_file", "host login key files").Default("").String()
	return stress
}

func main() {
	// testLogging()
	// testS3Upload()
	// testS3Download()
	// testS3ListObject()

	app := baseApp()
	stress := vizionStressCommand(app)
	// var s3args testcases.S3TestArgs
	// s3Cmd := testcases.S3Command(stress, &s3args)

	switch kingpin.MustParse(app.Parse(os.Args[1:])) {
	// stress args
	case stress.FullCommand():
		println(stress)
		// case s3Cmd.FullCommand():
		// 	println(s3Cmd)
	}
}
