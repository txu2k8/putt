package resources

import (
	"putt/libs/sshmgr"
	"putt/types"
	"testing"
)

func TestVizion(t *testing.T) {
	host := "10.25.119.77"
	sshKey := sshmgr.SSHKey{
		UserName: "root",
		Password: "password",
		Port:     22,
	}
	v := Vizion{Base: types.BaseInput{SSHKey: sshKey}}
	exist := v.Node(host).IsDplmodExist()
	logger.Info(exist)
}
