package main

import (
	"gtest/cmd"
	_ "gtest/config"
	"gtest/libs/retry"
	"gtest/libs/retry/backoff"
	"gtest/libs/retry/jitter"
	"gtest/libs/retry/strategy"
	_ "gtest/testinit"
	"log"
	"math/rand"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/op/go-logging"
	. "github.com/smartystreets/goconvey/convey"
)

var logger = logging.MustGetLogger("test")

func testLogging(filePath string) error {
	logger.Infof("Open logFile: %s...", filePath)
	_, err := os.Open(filePath)
	if err == nil {
		logger.Info("------info")
		logger.Notice("------notice")
		logger.Warning("------warning")
		logger.Error("------err")
		logger.Critical("------crit")
	}
	return err
}

func testRetry() bool {
	const logFilePath = "./test.log1"

	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))
	err := retry.Retry(func(attempt uint) error {
		return testLogging(logFilePath)
	},
		strategy.Limit(3),
		strategy.Wait(2*time.Second),
		strategy.BackoffWithJitter(
			backoff.BinaryExponential(10*time.Millisecond),
			jitter.Deviation(random, 0.5),
		),
	)

	if err != nil {
		log.Fatalf("Unable to open file %q with error %q", logFilePath, err)
	}
	return true
}

// TestRetry ...
func TestRetry(t *testing.T) {
	Convey("Test Retry", t, func() {
		So(testRetry(), ShouldEqual, true)
	})
}

func main() {
	// testLogging()
	// testS3Upload()
	// testS3Download()
	// testS3ListObject()
	// utils.SleepProgressBar(2)
	// testRetry()
	// fmt.Printf("  %-3s %-12s  CaseDescription", "NO.", "CaseName")
	logger.Infof("Args: gtest %s", strings.Join(os.Args[1:], " "))
	cmd.Execute()
}
