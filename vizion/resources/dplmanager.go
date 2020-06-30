package resources

// Define dplmanager opt on vizion node, by ssh to node and then run dplmanager commands

import (
	"fmt"
	"putt/config"
	"putt/libs/sshmgr"
	"putt/libs/utils"
	"regexp"
	"strings"
)

// DplmanagerGetter has a method to return a Dplmanager.
// A group's client should implement this interface.
type DplmanagerGetter interface {
	DplMgr(host string) Dplmanager
}

// Dplmanager has methods to work on Node resources.
type Dplmanager interface {
	GetJnsStat(vsetID int, anchorUUID, dplUUID string) error
	DplHelo(dplIP string, dplPort int) error
}

// dplMgr implements Dplmanager
type dplMgr struct {
	*sshmgr.SSHMgr
	Image      string // dpl image
	DplmgrPath string //dplmanager path
	EtcdIPPort string // etcd ip:port
}

// newdplMgr returns a dplMgr
func newdplMgr(v *Vizion, host, image string) *dplMgr {
	dplmgrPath := config.DplmanagerLocalPath
	if image != "" {
		dplmgrPath = config.Dplmanager.Path
	}
	etcdIPPort := fmt.Sprintf("%s:%d", v.VaildMasterIP(), 2379)

	return &dplMgr{
		sshmgr.NewSSHMgr(host, v.Base.SSHKey),
		image,
		dplmgrPath,
		etcdIPPort,
	}
}

func (d *dplMgr) RunCmdInDocker(cmdSpec string) (int, string) {
	dockerArgs := "--dns=10.233.0.10 --dns-search=svc.cluster.local"
	dockerCmdSpec := fmt.Sprintf("docker run -i --rm --network host %s -v /dev:/dev -v /etc:/etc --privileged %s bash -c '%s'", dockerArgs, d.Image, cmdSpec)
	return d.RunCmd(dockerCmdSpec)
}

// GetJnsStat ... dplmanager -m jns stat
func (d *dplMgr) GetJnsStat(vsetID int, anchorUUID, dplUUID string) error {
	cmdSpec := fmt.Sprintf("%s -m jns -x vset_id=%d -x anchor_uuid=%s stat", d.DplmgrPath, vsetID, anchorUUID)
	if dplUUID != "" {
		cmdSpec += fmt.Sprintf(" -e dpl_uuid=%s", dplUUID)
	}
	_, output := d.RunCmd(cmdSpec)
	logger.Info(output)
	jnsStrArr := strings.Split(output, "\n")
	pattern := regexp.MustCompile(`'Anchor:\s+([^\s]+)\s+Dpl:\s+([^\s]+)\s+JState:\s+([^\s]+)\s+'`)
	for _, jnsStr := range jnsStrArr {
		if strings.HasPrefix(jnsStr, "Anchor:") {
			continue
		}
		matched := pattern.FindAllStringSubmatch(jnsStr, -1)
		logger.Info(utils.Prettify(matched))

	}

	return nil
}

// DplHelo dplmanager -m dpl helo
func (d *dplMgr) DplHelo(dplIP string, dplPort int) error {
	cmdSpec := fmt.Sprintf("%s -m dpl -a %s -p %d helo", d.DplmgrPath, dplIP, dplPort)
	_, output := d.RunCmdInDocker(cmdSpec)
	if strings.Contains(output, "success connect with dplserver") && strings.Contains(output, "Successfully got HELO response") {
		logger.Info(output)
		return nil
	}
	return fmt.Errorf(output)
}
