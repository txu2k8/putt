package resources

import (
	"fmt"
	"gtest/libs/retry"
	"gtest/libs/retry/backoff"
	"gtest/libs/retry/strategy"
	"gtest/libs/s3client"
	"gtest/libs/utils"
	"gtest/models"
	"os"
	"path"
	"strings"
	"time"
)

// UploadFile define the local file for upload
type UploadFile struct {
	FileName     string
	FileFullPath string
	FileMd5sum   string
	FileSize     int64
}

// CreateUploadFiles ...
func CreateUploadFiles(conf models.S3TestInput) []UploadFile {
	logger.Info("> Prepare upload data ...")
	var fileArr []UploadFile
	var randomSize int64
	var fileMd5 string
	timeStr := time.Now().Format("20060102150405")

	for _, fileConf := range conf.S3TestFileInputs {
		if conf.RenameFile == true {
			fileConf.FileNamePrefix = fileConf.FileNamePrefix + "_" + timeStr
		}
		emptyIdx := fileConf.FileNum * conf.EmptyPercent / 100
		randomIdx := fileConf.FileNum * conf.RandomPercent / 100
		for i := 0; i < fileConf.FileNum; i++ {
			uploadFile := UploadFile{}
			fileName := fmt.Sprintf("%s_%d.%s", fileConf.FileNamePrefix, i, fileConf.FileType)
			filePath := path.Join(fileConf.FileDir, fileName)

			if i < emptyIdx {
				randomSize = 0
			} else {
				randomSize = utils.GetRandomInt64(fileConf.FileSizeMin, fileConf.FileSizeMax)
			}

			fExist, _ := utils.PathExists(filePath)
			if (i < randomIdx) && (fExist == true) {
				fileMd5 = utils.GetFileMd5sumWithPath(filePath)
			} else {
				fileMd5 = utils.CreateFile(filePath, randomSize, 128)
			}

			uploadFile.FileName = fileName
			uploadFile.FileFullPath = filePath
			uploadFile.FileSize = randomSize
			uploadFile.FileMd5sum = fileMd5
			fileArr = append(fileArr, uploadFile)
		}
	}
	return fileArr
}

// S3UploadFiles ...
func S3UploadFiles(conf models.S3TestInput) ([]UploadFile, error) {
	logger.Info(">> Upload: Vizion S3 upload test start ...")
	conf.ParseS3Input()
	logger.Info(conf)
	localFiles := CreateUploadFiles(conf)
	endpoint := fmt.Sprintf("https://%s:%d", conf.S3Ip, conf.S3Port)
	session := s3client.NewSession(endpoint, conf.S3AccessID, conf.S3SecretKey)
	for _, file := range localFiles {
		action := func(attempt uint) error {
			return s3client.UploadFileWithProcess(session, conf.S3Bucket, file.FileFullPath)
		}
		err := retry.Retry(
			action,
			strategy.Limit(5),
			strategy.Backoff(backoff.Fibonacci(10*time.Millisecond)),
		)
		if err != nil {
			return []UploadFile{}, err
		}
	}
	logger.Info(">> Upload: Vizion S3 upload test complete ...")
	return localFiles, nil
}

// CreateDownloadDir ...
func CreateDownloadDir(conf models.S3TestInput) string {
	strS3ip := strings.ReplaceAll(conf.S3Ip, ".", "")
	strTime := time.Now().Format("20060102150405")
	dPath := path.Join(conf.LocalDataDir, fmt.Sprintf("download_%s", strS3ip), conf.S3Bucket, strTime)

	_, err := os.Stat(dPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dPath, os.ModePerm)
		if err != nil {
			logger.Panicf("mkdir failed![%v]", err)
		}
	}

	return dPath
}

// S3DownloadFiles ...
func S3DownloadFiles(conf models.S3TestInput, downloadFiles []UploadFile) error {
	logger.Info(">> Download: Vizion S3 download test start ...")
	conf.ParseS3Input()
	logger.Info(conf)
	downloadDir := CreateDownloadDir(conf)
	endpoint := fmt.Sprintf("https://%s:%d", conf.S3Ip, conf.S3Port)
	svc := s3client.NewS3Client(endpoint, conf.S3AccessID, conf.S3SecretKey)

	for _, file := range downloadFiles {
		action := func(attempt uint) error {
			return s3client.DownloadFileWithProcess(svc, conf.S3Bucket, file.FileName, downloadDir)
		}
		err := retry.Retry(
			action,
			strategy.Limit(5),
			strategy.Backoff(backoff.Fibonacci(10*time.Millisecond)),
		)
		if err != nil {
			return err
		}
	}
	logger.Info(">> Download: Vizion S3 download test complete ...")
	return nil
}

// S3ListBucketObjects ...
func S3ListBucketObjects(conf models.S3TestInput) error {
	return nil
}

// S3DeleteBucketFiles ...
func S3DeleteBucketFiles(conf models.S3TestInput, uploadFiles []UploadFile) error {
	logger.Info(">> Delete: Vizion S3 delete test start ...")
	conf.ParseS3Input()
	logger.Info(conf)
	endpoint := fmt.Sprintf("https://%s:%d", conf.S3Ip, conf.S3Port)
	svc := s3client.NewS3Client(endpoint, conf.S3AccessID, conf.S3SecretKey)

	for _, file := range uploadFiles {
		action := func(attempt uint) error {
			return s3client.DeleteBucketFile(svc, conf.S3Bucket, file.FileName)
		}
		err := retry.Retry(
			action,
			strategy.Limit(5),
			strategy.Backoff(backoff.Fibonacci(10*time.Millisecond)),
		)
		if err != nil {
			return err
		}
	}
	logger.Info(">> Delete: Vizion S3 delete test complete ...")
	return nil
}

// S3UploadDownloadListDeleteFiles ...
func S3UploadDownloadListDeleteFiles(conf models.S3TestInput) error {
	uploadFiles, err := S3UploadFiles(conf)
	if err != nil {
		return err
	}

	if err := S3ListBucketObjects(conf); err != nil {
		return err
	}

	if err := S3DownloadFiles(conf, uploadFiles); err != nil {
		return err
	}

	if err := S3DeleteBucketFiles(conf, uploadFiles); err != nil {
		return err
	}

	return nil
}
