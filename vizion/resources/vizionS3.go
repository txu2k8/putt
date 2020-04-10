package resources

import (
	"fmt"
	"gtest/libs/s3client"
	"gtest/libs/utils"
	"gtest/models"
	"path"
)

// UploadFile define the local file for upload
type UploadFile struct {
	FileName     string
	FileFullPath string
	FileMd5sum   string
	FileSize     int64
}

// CreateUploadFiles ...
func CreateUploadFiles(confs []models.S3TestFileInput) []UploadFile {
	logger.Info("> Prepare upload data ...")
	var fileList []UploadFile
	for _, conf := range confs {
		for i := 0; i < conf.FileNum; i++ {
			uploadFile := UploadFile{}
			fileName := fmt.Sprintf("%s_%d.%s", conf.FileNamePrefix, i, conf.FileType)
			filePath := path.Join(conf.FileDir, fileName)
			randomSize := utils.GetRangeRand(conf.FileSizeMin, conf.FileSizeMax)
			fileMd5 := utils.CreateFile(filePath, randomSize, 128)
			uploadFile.FileName = fileName
			uploadFile.FileFullPath = filePath
			uploadFile.FileSize = randomSize
			uploadFile.FileMd5sum = fileMd5
			fileList = append(fileList, uploadFile)
		}
	}
	return fileList
}

// S3UploadFiles ...
func S3UploadFiles(conf models.S3TestInput) {
	logger.Info(">> Upload: Vizion S3 upload test ...")
	logger.Info(conf)
	localFiles := CreateUploadFiles(conf.S3TestFileInputs)
	endpoint := fmt.Sprintf("https://%s:%d", conf.S3Ip, conf.S3Port)
	session := s3client.NewSession(endpoint, conf.S3AccessID, conf.S3SecretKey)
	for _, file := range localFiles {
		s3client.UploadFileWithProcess(session, conf.S3Bucket, file.FileFullPath)
	}
}
