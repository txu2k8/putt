package sshmgr

import (
	"testing"
)

func aTestRunCmdWithOutput(t *testing.T) {
	sshInput := SSHInput{"10.25.119.1", "root", "password", 22, ""}
	session, _ := sshInput.NewSessionWithRetry()
	defer session.Close()
	rc, output := RunCmdWithOutput(session, "pwd; ls")
	logger.Infof("%d, %s\n", rc, output)
}

func aTestRunCmd(t *testing.T) {
	sshInput := SSHInput{"10.25.119.1", "root", "password", 22, ""}
	rc, output := sshInput.RunCmd("pwd; ls")
	logger.Infof("%d, %s\n", rc, output)
}


func TestRetry(t * testing T) {
	const logFilePath = "./myapp.log"
	var logFile *os.File

	err := retry.Retry(func(attempt uint) error {
		var err error
		logFile, err = os.Open(logFilePath)
		return err
	})

	if nil != err {
		log.Fatalf("Unable to open file %q with error %q", logFilePath, err)
	}
}