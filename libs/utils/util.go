package utils

import (
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"os"

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

// SizeToName returns a human-readable string for the given size bytes
func SizeToName(size int) string {
	units := []string{"B", "KB", "MB", "GB"}
	i := 0
	for size >= 1024 {
		size /= 1024
		i++
	}

	if i > len(units)-1 {
		i = len(units) - 1
	}

	return fmt.Sprintf("%d%s", size, units[i])
}
