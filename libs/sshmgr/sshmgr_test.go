package sshmgr

import (
	"gtest/libs/retry"
	"gtest/libs/retry/backoff"
	"gtest/libs/retry/jitter"
	"gtest/libs/retry/strategy"
	"log"
	"math/rand"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunCmdWithOutput(t *testing.T) {
	sshInput := SSHInput{"10.25.119.1", "root", "password", 22, ""}
	session, _ := sshInput.NewSessionWithRetry()
	defer session.Close()
	rc, output := RunCmdWithOutput(session, "pwd; ls")
	logger.Infof("%d, %s\n", rc, output)
}

func TestRunCmd(t *testing.T) {
	sshInput := SSHInput{"10.25.119.1", "root", "password", 22, ""}
	rc, output := sshInput.RunCmd("pwd; ls")
	logger.Infof("%d, %s\n", rc, output)
}

func testRetry() bool {
	const logFilePath = "./myapp.log"

	seed := time.Now().UnixNano()
	random := rand.New(rand.NewSource(seed))
	err := retry.Retry(func(attempt uint) error {
		_, err := os.Open(logFilePath)
		return err
	},
		strategy.Limit(5),
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
func TestRetry(t *testing.T) {
	Convey("Test Retry", t, func() {
		So(testRetry(), ShouldEqual, true)
	})
}
