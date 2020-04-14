package sshmgr

import (
	"testing"
)

func aTestSessionRun(t *testing.T) {
	sshInput := SSHInput{"10.25.119.1", "root", "password", 22, ""}
	session, _ := sshInput.NewSessionWithRetry()
	defer session.Close()
	rc, output := SessionRun(session, "pwd; ls")
	logger.Infof("%d, %s\n", rc, output)
}

func TestRunCmd(t *testing.T) {
	sshInput := SSHInput{"10.25.119.1", "root", "password", 22, ""}
	rc, output := sshInput.RunCmd("pwd; ls")
	logger.Infof("%d, %s\n", rc, output)
}
