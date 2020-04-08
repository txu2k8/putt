package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"
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

// SizeCountByte Parse a size string to int64 byte
func SizeCountByte(s string) int64 {
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
