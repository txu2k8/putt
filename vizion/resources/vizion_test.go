package resources

import (
	"pzatest/libs/sshmgr"
	"pzatest/types"
	"testing"
)

func TestVizion(t *testing.T) {
	host := "10.25.119.77"
	sshKey := sshmgr.SSHKey{
		UserName: "root",
		Password: "password",
		Port:     22,
	}
	b := VizionBase{VizionBaseInput: types.VizionBaseInput{SSHKey: sshKey}}
	exist := b.Node(host).IsDplmodExist()
	logger.Info(exist)
	b.Service().K8sEnableNodeLabel("", "", "")
}
