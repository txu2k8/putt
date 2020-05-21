package tlog

import (
	"fmt"
	"os"
	"path"
	"testing"
	"time"

	"github.com/op/go-logging"
)

func initLogging() {
	dir, _ := os.Getwd()
	fileLogName := "test"
	fileLogPath := path.Join(dir, "log")
	timeStr := time.Now().Format("20060102150405")
	fileLogName = fmt.Sprintf("%s-%s.log", fileLogName, timeStr)
	fileLogPath = path.Join(fileLogPath, fileLogName)

	conf := NewOptions(
		OptionSetFileLogPath(fileLogPath),
	)
	conf.InitLogging()
}

// Testtlog ...
func TestTlog(t *testing.T) {
	initLogging()
	logger := logging.MustGetLogger("test")
	logger.Info("Test tlog")
}
