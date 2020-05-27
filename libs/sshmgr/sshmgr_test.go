package sshmgr

import (
	"log"
	"math/rand"
	"os"
	"pzatest/libs/retry"
	"pzatest/libs/retry/backoff"
	"pzatest/libs/retry/jitter"
	"pzatest/libs/retry/strategy"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestRunCmd(t *testing.T) {
	sshMgr := NewSSHMgr("10.25.119.1", SSHKey{UserName: "root", Password: "password", Port: 22, KeyFile: ""})
	rc, output := sshMgr.RunCmd("pwd; ls")
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
