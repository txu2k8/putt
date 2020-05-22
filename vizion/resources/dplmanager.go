package resources

// Define dplmanager opt on vizion node, by ssh to node and then run dplmanager commands

import (
	"pzatest/libs/sshmgr"
	"strings"
)

// DplmanagerGetter has a method to return a Dplmanager.
// A group's client should implement this interface.
type DplmanagerGetter interface {
	DplMgr(host string) Dplmanager
}

// Dplmanager has methods to work on Node resources.
type Dplmanager interface {
	GetJnsStat() bool
}

// dplMgr implements Dplmanager
type dplMgr struct {
	ssh sshmgr.SSHInput
}

// newdplMgr returns a dplMgr
func newdplMgr(b *VizionBase, host string) *dplMgr {
	return &dplMgr{
		ssh: sshmgr.SSHInput{Host: host, SSHKey: b.SSHKey},
	}
}

// GetJnsStat ... TODO
func (d *dplMgr) GetJnsStat() bool {
	cmdSpec := "lsmod | grep dpl"
	_, output := d.ssh.RunCmd(cmdSpec)
	logger.Info(output)
	if output != "" && strings.Contains(output, "dpl") {
		return true
	}
	return false
}
