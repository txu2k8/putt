package resources

import (
	"fmt"
	"gtest/libs/s3client"
	"gtest/libs/utils"
	"path"
)

// S3TestFileConfig define the s3 test file config
type S3TestFileConfig struct {
	FileType       string // txt or dd
	FileNum        int    // file number
	FileSizeMin    int64  // the min size of file
	FileSizeMax    int64  // the max size of the file
	FileNamePrefix string // the file name prefix
	FileDir        string // the file dir path
}

// S3TestConfig define the S3 test config
type S3TestConfig struct {
	S3Ip              string // endpoint: https://<S3Ip>:<S3Port>, eg: https://10.25.119.86:443
	S3AccessID        string
	S3SecretKey       string
	S3Port            int                // port (default: 443)
	S3Bucket          string             // s3 bucket for test
	LocalDataDir      string             // The local data Dir
	S3TestFileConfigs []S3TestFileConfig // S3 files config list
	RandomPercent     int                // percent of files with random data
	EmptyPercent      int                // percent of files with empty data
	RenameFile        bool               // rename files name each time if true
	DeleteFile        bool               // delete files from s3 bucket after test if true
	Clients           int                // S3 Client number for test at the same time
}

// UploadFile define the local file for upload
type UploadFile struct {
	FileName     string
	FileFullPath string
	FileMd5sum   string
	FileSize     int64
}

func (conf *S3TestConfig) setup() {

}

// CreateUploadFiles ...
func CreateUploadFiles(confs []S3TestFileConfig) []UploadFile {
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
func S3UploadFiles(conf S3TestConfig) {
	logger.Info(">> Upload: Vizion S3 upload test ...")
	logger.Info(conf)
	localFiles := CreateUploadFiles(conf.S3TestFileConfigs)
	endpoint := fmt.Sprintf("https://%s:%d", conf.S3Ip, conf.S3Port)
	session := s3client.NewSession(endpoint, conf.S3AccessID, conf.S3SecretKey)
	for _, file := range localFiles {
		s3client.UploadFileWithProcess(session, conf.S3Bucket, file.FileFullPath)
	}
}
