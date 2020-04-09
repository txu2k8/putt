package utils

import (
	"crypto/md5"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"path"
	"regexp"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// UniqueID returns a unique UUID-like identifier.
func UniqueID() string {
	uuid := make([]byte, 16)
	io.ReadFull(rand.Reader, uuid)
	return fmt.Sprintf("%s", uuid)
}

// GetRangeRand return rand int64 i range [-m, n]
func GetRangeRand(min, max int64) int64 {
	if min > max {
		logger.Panic("The min is greater than max!")
	}

	if min < 0 {
		f64Min := math.Abs(float64(min))
		i64Min := int64(f64Min)
		result, _ := rand.Int(rand.Reader, big.NewInt(max+1+i64Min))
		return result.Int64() - i64Min
	}
	result, _ := rand.Int(rand.Reader, big.NewInt(max-min+1))
	return min + result.Int64()
}

// GetRandomString return a random string
func GetRandomString(strSize int64) string {
	b := make([]byte, strSize)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	str := fmt.Sprintf("%X", b)
	return str
}

// GetFileMd5sum ...
func GetFileMd5sum(f *os.File) string {
	logger.Debugf("Get file MD5: %s", f.Name())
	md5 := md5.New()
	io.Copy(md5, f)
	fileMd5 := hex.EncodeToString(md5.Sum(nil))
	logger.Debug(fileMd5, f.Name())
	return fileMd5
}

// PathExists use "os.Stat" judge if the file or dir exist
// nil: exist
// os.IsNotExist(err) == true: not exist
// other error type: not sure
func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

// CreateFile create original file, each line with line_number, and specified line size
//
func CreateFile(filePath string, fileSize int64, lineSize int64) string {
	logger.Infof(">> Create file: %s", filePath)
	fileDir := path.Dir(filePath)
	err := os.MkdirAll(fileDir, os.ModePerm)
	if err != nil {
		logger.Panic(err)
	}

	lineCount := fileSize / lineSize
	unalignedSize := fileSize % lineSize

	file, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		logger.Panic(err)
	}
	defer file.Close()

	var lineNum int64
	for lineNum = 0; lineNum < lineCount; lineNum++ {
		randString := GetRandomString(lineSize-int64(2)-int64(len(string(lineNum)))) + "\n"
		randLineString := fmt.Sprintf("%s:%s", strconv.FormatInt(lineNum, 10), randString)
		file.WriteString(randLineString)
	}
	if unalignedSize > 0 {
		file.WriteString(GetRandomString(unalignedSize))
	}
	fileMd5 := GetFileMd5sum(file)
	return fileMd5
}

// CreateFileOfSize will return an *os.File that is of size bytes
func CreateFileOfSize(dir string, fileNamePrefix string, size int64) (*os.File, error) {
	file, err := ioutil.TempFile(dir, fileNamePrefix)
	if err != nil {
		return nil, err
	}

	err = file.Truncate(size)
	if err != nil {
		file.Close()
		os.Remove(file.Name())
		return nil, err
	}

	return file, nil
}

// ByteCountDecimal returns a human-readable string for the given size bytes
// precision: decimal, 12 * 1024 * 1024 * 1000 --> 11.7GB
func ByteCountDecimal(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "kMGTPE"[exp])
}

// SizeCountToByte Parse a size string to int64 byte
func SizeCountToByte(s string) int64 {
	const unit = 1024
	const u = "kMGTPE"
	div := float64(unit)

	reg := regexp.MustCompile(`(^[+-]?(0|([1-9]\d*))(\.\d+)?)\s?(\S+)`)
	matched := reg.FindStringSubmatch(s)
	exp := strings.Index(u, strings.ToUpper(matched[len(matched)-1][:1]))
	for x := 0; x < exp; x++ {
		div *= unit
	}
	n, _ := strconv.ParseFloat(matched[1], 64)
	b := int64(n * div)
	return b
}
