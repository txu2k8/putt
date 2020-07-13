package resources

import (
	"platform/libs/sshmgr"
	"platform/libs/utils"
	"platform/types"
	"testing"
)

func TestVizion(t *testing.T) {
	host := "10.25.119.71"
	sshKey := sshmgr.SSHKey{
		UserName: "root",
		Password: "password",
		Port:     22,
	}
	v := Vizion{Base: types.BaseInput{SSHKey: sshKey}}
	n := v.Node(host)
	rsp := n.GetEtcdEndpoints()
	logger.Info(utils.Prettify(rsp))
}
