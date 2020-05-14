package testcase

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"pzatest/libs/prettytable"
	"pzatest/libs/retry"
	"pzatest/libs/retry/backoff"
	"pzatest/libs/retry/strategy"
	"pzatest/libs/s3client"
	"pzatest/libs/utils"
	"pzatest/models"
	"regexp"
	"strings"
	"sync"
	"time"
)

// UploadFile define the local file for upload
type UploadFile struct {
	FileName     string
	FileFullPath string
	FileMd5sum   string
	FileSize     int64
}

// Worker ...
type Worker struct {
	wg          sync.WaitGroup
	done        chan struct{}
	maxParallel int
}

// CreateUploadFiles ...
func CreateUploadFiles(conf models.S3TestInput) []UploadFile {
	logger.Info("> Prepare upload data ...")
	var fileArr []UploadFile
	var randomSize int64
	var mode string = "a+"
	var fileMd5 string

	for _, fileConf := range conf.S3TestFileInputs {
		fileUploadNamePrefix := fileConf.FileNamePrefix
		patternPrefix := fileConf.FileNamePrefix
		emptyIdx := fileConf.FileNum * conf.EmptyPercent / 100
		randomIdx := fileConf.FileNum * conf.RandomPercent / 100
		fileUploadDir := path.Join(fileConf.FileDir, "upload")

		// Get the exist files list
		existFileInfoList, err := ioutil.ReadDir(fileUploadDir)
		if err != nil {
			logger.Panicf("List local files fail: %s", err)
		}

		if conf.RenameFile == true {
			timeStr := time.Now().Format("20060102150405")
			fileUploadNamePrefix += "_" + timeStr
			patternPrefix += "_\\d{14}"
		}

		for i := 0; i < fileConf.FileNum; i++ {
			uploadFile := UploadFile{}
			fileName := fmt.Sprintf("%s_%d.%s", fileUploadNamePrefix, i, fileConf.FileType)
			filePath := path.Join(fileUploadDir, fileName)
			// os.Rename the exist file with diff timeStr
			if conf.RenameFile == true {
				pattern := fmt.Sprintf("%s_%d.%s", patternPrefix, i, fileConf.FileType)
				for _, existFile := range existFileInfoList {
					existFileName := existFile.Name()
					matched, _ := regexp.MatchString(pattern, existFileName)
					if matched {
						existFilePath := path.Join(fileUploadDir, existFileName)
						logger.Infof("os.Rename: %s -> %s", existFilePath, filePath)
						os.Rename(existFilePath, filePath)
						break
					}
				}
			}

			if i < emptyIdx {
				randomSize = 0
				mode = "w+"
			} else {
				randomSize = utils.GetRandomInt64(fileConf.FileSizeMin, fileConf.FileSizeMax)
			}

			fExist, _ := utils.PathExists(filePath)
			if (i < randomIdx) && (fExist == true) {
				fileMd5 = utils.GetFileMd5sumWithPath(filePath)
			} else {
				fileMd5 = utils.CreateFile(filePath, randomSize, 128, mode)
			}

			uploadFile.FileName = fileName
			uploadFile.FileFullPath = filePath
			uploadFile.FileSize = randomSize
			uploadFile.FileMd5sum = fileMd5
			logger.Infof("Local File(md5:%s):%s", fileMd5, filePath)
			fileArr = append(fileArr, uploadFile)
		}
	}
	logger.Panic()
	return fileArr
}

// CreateDownloadDir ...
func CreateDownloadDir(conf models.S3TestInput) string {
	strS3ip := strings.ReplaceAll(conf.S3Ip, ".", "")
	strTime := time.Now().Format("20060102150405")
	dPath := path.Join(conf.LocalDataDir, strS3ip, "download", conf.S3Bucket, strTime)

	_, err := os.Stat(dPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(dPath, os.ModePerm)
		if err != nil {
			logger.Panicf("mkdir failed![%v]", err)
		}
	}

	return dPath
}

// CheckDownloadMd5sum ...
func CheckDownloadMd5sum(uploadFiles []UploadFile, downloadDir string) error {
	logger.Info(">> Check-Md5sum: download vs upload files ...")
	table, _ := prettytable.NewTable(
		prettytable.Column{Header: "File Name", AlignRight: false, MinWidth: 20},
		prettytable.Column{Header: "Upload", AlignRight: false, MinWidth: 20},
		prettytable.Column{Header: fmt.Sprintf("Download(%s)", downloadDir), AlignRight: false, MinWidth: 30},
	)
	tableErr, _ := prettytable.NewTable(
		prettytable.Column{Header: "File Name", AlignRight: false, MinWidth: 20},
		prettytable.Column{Header: "Upload", AlignRight: false, MinWidth: 20},
		prettytable.Column{Header: fmt.Sprintf("Download(%s)", downloadDir), AlignRight: false, MinWidth: 30},
	)
	table.Separator = " | "
	tableErr.Separator = " | "

	for _, file := range uploadFiles {
		fileName := file.FileName
		fileMd5 := file.FileMd5sum
		downloadFilePath := path.Join(downloadDir, fileName)
		downloadFileMd5 := utils.GetFileMd5sumWithPath(downloadFilePath)

		if fileMd5 == downloadFileMd5 {
			table.AddRow(fileName, fileMd5, downloadFileMd5)
		} else {
			tableErr.AddRow(fileName, fileMd5, downloadFileMd5)
		}
	}

	if len(table.Rows) > 0 {
		logger.Infof("> upload/download files md5 matched:\n%s", table.String())
	}

	if len(tableErr.Rows) > 0 {
		logger.Errorf("> upload/download files md5 mismatch:\n%s", tableErr.String())
		return errors.New("Download md5 mismatch Error")
	}

	return nil
}

// S3UploadFiles ...
func S3UploadFiles(conf models.S3TestInput) ([]UploadFile, error) {
	logger.Info(">> Upload: Vizion S3 upload test start ...")
	conf.ParseS3Input()
	localFiles := CreateUploadFiles(conf)
	session := s3client.NewSession(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)
	for _, file := range localFiles {
		if err := s3client.UploadFileWithProcessRetry(session, conf.S3Bucket, file.FileFullPath); err != nil {
			return []UploadFile{}, err
		}
	}
	logger.Info(">> Upload: Vizion S3 upload test complete ...")
	return localFiles, nil
}

// S3DownloadFiles ...
func S3DownloadFiles(conf models.S3TestInput, downloadFiles []UploadFile) error {
	logger.Info(">> Download: Vizion S3 download test start ...")
	conf.ParseS3Input()
	downloadDir := CreateDownloadDir(conf)
	svc := s3client.NewS3Client(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)

	for _, file := range downloadFiles {
		if err := s3client.DownloadFileWithProcessRetry(svc, conf.S3Bucket, file.FileName, downloadDir); err != nil {
			return err
		}
	}
	logger.Info("Multi-Download: Complete ...")

	if err := CheckDownloadMd5sum(downloadFiles, downloadDir); err != nil {
		return err
	}
	logger.Info(">> Download: Vizion S3 download test complete ...")

	logger.Infof(">> Delete local S3 download files:%s", downloadDir)
	os.RemoveAll(downloadDir)

	return nil
}

// S3ListBuckets ...
func S3ListBuckets(conf models.S3TestInput) error {
	var err error
	logger.Info(">> Multi-ListBuckets: Vizion S3 ListBuckets test start ...")
	conf.ParseS3Input()
	svc := s3client.NewS3Client(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)
	err = s3client.ListBuckets(svc)
	logger.Info(">> Multi-ListBuckets: Vizion S3 ListBuckets test complete ...")
	return err
}

// S3ListBucketObjects ...
func S3ListBucketObjects(conf models.S3TestInput) error {
	var err error
	logger.Info(">> Multi-ListBucketFiles: Vizion S3 ListBucketFiles test start ...")
	conf.ParseS3Input()
	svc := s3client.NewS3Client(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)
	err = s3client.ListBucketFiles(svc, conf.S3Bucket)
	logger.Info(">> Multi-ListBucketFiles: Vizion S3 ListBucketFiles test complete ...")
	return err
}

// S3DeleteBucketFiles ...
func S3DeleteBucketFiles(conf models.S3TestInput, uploadFiles []UploadFile) error {
	logger.Info(">> Delete: Vizion S3 delete test start ...")
	conf.ParseS3Input()
	svc := s3client.NewS3Client(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)

	for _, file := range uploadFiles {
		action := func(attempt uint) error {
			return s3client.DeleteBucketFileRetry(svc, conf.S3Bucket, file.FileName)
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

	if err := S3ListBuckets(conf); err != nil {
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

// ================ Multi ================

// MultiS3UploadFiles ...
func MultiS3UploadFiles(conf models.S3TestInput) ([]UploadFile, error) {
	var err error
	logger.Info(">> Multi-Upload: Vizion S3 upload test start ...")
	conf.ParseS3Input()
	localFiles := CreateUploadFiles(conf)
	session := s3client.NewSession(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)

	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)
	for _, file := range localFiles {
		fileFullPath := file.FileFullPath
		select {
		case ch <- struct{}{}:
			w.wg.Add(1)
			go func() {
				err = s3client.UploadFileRetry(session, conf.S3Bucket, fileFullPath)
				if err != nil {
					w.wg.Done()
					w.done <- struct{}{}
				}
				<-ch
				w.wg.Done()
			}()
		case <-w.done:
			break
		}
	}
	w.wg.Wait()
	logger.Info(">> Multi-Upload: Vizion S3 upload test complete ...")
	return localFiles, err
}

// MultiS3DownloadFiles ...
func MultiS3DownloadFiles(conf models.S3TestInput, downloadFiles []UploadFile) error {
	var err error
	logger.Info(">> Multi-Download: Vizion S3 download test start ...")
	conf.ParseS3Input()
	downloadDir := CreateDownloadDir(conf)
	svc := s3client.NewS3Client(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)

	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)
	for _, file := range downloadFiles {
		fileName := file.FileName
		select {
		case ch <- struct{}{}:
			w.wg.Add(1)
			go func() {
				err = s3client.DownloadFileRetry(svc, conf.S3Bucket, fileName, downloadDir)
				if err != nil {
					w.wg.Done()
					w.done <- struct{}{}
				}
				<-ch
				w.wg.Done()
			}()
		case <-w.done:
			break
		}
	}
	w.wg.Wait()
	if err != nil {
		return err
	}
	logger.Info("Multi-Download: Complete ...")

	if err = CheckDownloadMd5sum(downloadFiles, downloadDir); err != nil {
		return err
	}

	logger.Info(">> Multi-Download: Vizion S3 download test complete ...")
	logger.Infof(">> Delete local S3 download files:%s", downloadDir)
	os.RemoveAll(downloadDir)

	return err
}

// MultiS3DeleteBucketFiles ...
func MultiS3DeleteBucketFiles(conf models.S3TestInput, uploadFiles []UploadFile) error {
	var err error
	logger.Info(">> Multi-Delete: Vizion S3 delete test start ...")
	conf.ParseS3Input()
	svc := s3client.NewS3Client(conf.Endpoint, conf.S3AccessID, conf.S3SecretKey)

	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)
	for _, file := range uploadFiles {
		fileName := file.FileName
		select {
		case ch <- struct{}{}:
			w.wg.Add(1)
			go func() {
				err = s3client.DeleteBucketFileRetry(svc, conf.S3Bucket, fileName)
				if err != nil {
					w.done <- struct{}{}
				}
				<-ch
				w.wg.Done()
			}()
		case <-w.done:
			break
		}
	}
	w.wg.Wait()
	logger.Info(">> Multi-Delete: Vizion S3 delete test complete ...")
	return err
}

// MultiS3UploadDownloadListDeleteFiles ...
func MultiS3UploadDownloadListDeleteFiles(conf models.S3TestInput) error {
	uploadFiles, err := MultiS3UploadFiles(conf)
	if err != nil {
		return err
	}

	if err := S3ListBuckets(conf); err != nil {
		return err
	}

	if err := S3ListBucketObjects(conf); err != nil {
		return err
	}

	if err := MultiS3DownloadFiles(conf, uploadFiles); err != nil {
		return err
	}

	if err := MultiS3DeleteBucketFiles(conf, uploadFiles); err != nil {
		return err
	}

	return nil
}

// MultiUserS3UploadDownloadListDeleteFiles ... TODO
func MultiUserS3UploadDownloadListDeleteFiles(conf models.S3TestInput) error {
	var err error
	logger.Info(">> Multi-Users: Vizion S3 UploadDownloadListDeleteFiles test start ...")

	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)
	for i := 0; i < conf.Clients; i++ {
		time.Sleep(2 * time.Second)
		select {
		case ch <- struct{}{}:
			w.wg.Add(1)
			go func() {
				err = S3UploadDownloadListDeleteFiles(conf)
				if err != nil {
					w.wg.Done()
					w.done <- struct{}{}
				}
				<-ch
				w.wg.Done()
			}()
		case <-w.done:
			break
		}
	}
	w.wg.Wait()
	logger.Info(">> Multi-Users: Vizion S3 UploadDownloadListDeleteFiles test complete ...")
	return err
}
