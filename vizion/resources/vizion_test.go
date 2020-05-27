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
	v := Vizion{Base: types.VizionBaseInput{SSHKey: sshKey}}
	exist := v.Node(host).IsDplmodExist()
	logger.Info(exist)
	v.Service().K8sEnableNodeLabel("", "", "")
}
