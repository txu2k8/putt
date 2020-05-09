package resources

import (
	"fmt"
	"gtest/libs/retry"
	"gtest/libs/retry/backoff"
	"gtest/libs/retry/strategy"
	"gtest/libs/s3client"
	"gtest/libs/utils"
	"gtest/models"
	"path"
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
func S3UploadFiles(conf models.S3TestInput) error {
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
			return err
		}
	}
	logger.Info(">> Upload: Vizion S3 upload test complete ...")
	return nil
}

// S3DownloadFiles ...
func S3DownloadFiles(conf models.S3TestInput) error {
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
			return err
		}
	}
	logger.Info(">> Upload: Vizion S3 upload test complete ...")
	return nil
}
