package resources

import (
	"platform/libs/sshmgr"
	"testing"
)

func TestIsDplmodExist(t *testing.T) {
	host := "10.25.119.77"
	sshKey := sshmgr.SSHKey{
		UserName: "root",
		Password: "password",
		Port:     22,
	}
	n := node{sshmgr.NewSSHMgr(host, sshKey)}
	logger.Info(n.IsDplmodExist())
	logger.Info(n.IsDplmodExist())
}
