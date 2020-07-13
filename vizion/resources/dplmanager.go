package resources

// Define dplmanager opt on vizion node, by ssh to node and then run dplmanager commands

import (
	"fmt"
	"platform/config"
	"platform/libs/sshmgr"
	"regexp"
	"strconv"
	"strings"
)

// DplmanagerGetter has a method to return a Dplmanager.
// A group's client should implement this interface.
type DplmanagerGetter interface {
	DplMgr(host string) Dplmanager
}

// Dplmanager has methods to work on Node resources.
type Dplmanager interface {
	// dpl helo
	DplHelo(dplIP string, dplPort int) error

	// Jns stat
	GetJnsStat(vsetID int, anchorUUID, jnlUUID string) (jnsMapArr []map[string]string, err error)
	IsJnsStatExpected(vsetID int, anchorUUID, jnlUUID, expected string) error
	IsJnsStatPrimary(vsetID int, anchorUUID, jnlUUID string) error
	IsAnyJnsStatPrimary(vsetID int, anchorUUID, jnlUUID string) error
	IsJnsStatSecondary(vsetID int, anchorUUID, jnlUUID string) error
	IsJnsStatDisconnected(vsetID int, anchorUUID, jnlUUID string) error

	// Channel
	PrintDplChannels(dplIP string, dplPort int) error
	GetDplChannels(dplIP string, dplPort int) (chMapArr []map[string]string, err error)
}

// dplMgr implements Dplmanager
type dplMgr struct {
	*sshmgr.SSHMgr
	Image      string // dpl image
	DplmgrPath string //dplmanager path
	EtcdIPPort string // etcd ip:port
	EtcdArgs   string // docker -e ETCD_CERT/ETCD_CA/ETCD_KEY
	DNSArgs    string // docker --dns
}

// newdplMgr returns a dplMgr
func newdplMgr(v *Vizion, host, image string) *dplMgr {
	dplmgrPath := config.DplmanagerLocalPath
	if image != "" {
		dplmgrPath = config.Dplmanager.Path
	}

	// etcdIPPort := fmt.Sprintf("%s:%d", v.VaildMasterIP(), 2379)
	var etcdIPPorts []string
	for _, mIP := range v.Base.MasterIPs {
		etcdIPPorts = append(etcdIPPorts, fmt.Sprintf("%s:%d", mIP, 2379))
	}
	etcdIPPort := strings.Join(etcdIPPorts, ",")
	logger.Debugf("etcdIPPort: %s", etcdIPPort)
	dnsArgs := "--dns=10.233.0.10 --dns-search=svc.cluster.local "
	etcdArgs := "-e ETCD_CERT='/etc/kubernetes/pki/etcd/peer.crt' -e ETCD_CA='/etc/kubernetes/pki/etcd/ca.crt' -e ETCD_KEY='/etc/kubernetes/pki/etcd/peer.key' "
	return &dplMgr{
		sshmgr.NewSSHMgr(host, v.Base.SSHKey),
		image,
		dplmgrPath,
		etcdIPPort,
		etcdArgs,
		dnsArgs,
	}
}

func (d *dplMgr) RunCmdInDocker(cmdSpec string, dockerArgs ...string) (int, string) {
	if len(dockerArgs) == 0 {
		dockerArgs = append(dockerArgs, d.DNSArgs)
	}
	dockerCmdSpec := fmt.Sprintf("docker run -i --rm --network host %s -v /dev:/dev -v /etc:/etc --privileged %s bash -c '%s'", strings.Join(dockerArgs, " "), d.Image, cmdSpec)
	return d.RunCmd(dockerCmdSpec)
}

// =============== DPL Helo: dplmanager -m dpl helo ===============
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

// =============== JState: dplmanager -m jns stat ===============
// GetJnsStat ... dplmanager -m jns stat
func (d *dplMgr) GetJnsStat(vsetID int, anchorUUID, jnlUUID string) (jnsMapArr []map[string]string, err error) {
	cmdSpec := fmt.Sprintf("%s -m jns -x vset_id=%d -x anchor_uuid=%s stat", d.DplmgrPath, vsetID, anchorUUID)
	if jnlUUID != "" {
		cmdSpec += fmt.Sprintf(" -e jnl_uuid=%s", jnlUUID)
	}
	rc, output := d.RunCmdInDocker(cmdSpec, d.DNSArgs, d.EtcdArgs)
	logger.Info(output)
	if rc != 0 {
		err = fmt.Errorf("Get jns status failed")
		return
	}
	jnsStrArr := strings.Split(output, "\n")
	pattern := regexp.MustCompile(`Anchor:\s+([^\s]+)\s+Jnl:\s+([^\s]+)\s+JState:\s+([^\s]+)\s+`)
	for _, jnsStr := range jnsStrArr {
		if !strings.HasPrefix(jnsStr, "Anchor:") {
			continue
		}
		jnsMap := map[string]string{}
		matched := pattern.FindAllStringSubmatch(jnsStr, -1)
		// logger.Info(utils.Prettify(matched))
		if len(matched) > 0 {
			jnsMap["Anchor"] = matched[0][1]
			jnsMap["Jnl"] = matched[0][2]
			jnsMap["JState"] = matched[0][3]
			jnsMapArr = append(jnsMapArr, jnsMap)
		}
	}

	return
}

// IsJnsStatExpected status: PRIMARY | SECONDARY | DISCONNECTED
func (d *dplMgr) IsJnsStatExpected(vsetID int, anchorUUID, jnlUUID, expected string) error {
	jnsStatArr, _ := d.GetJnsStat(vsetID, anchorUUID, jnlUUID)
	if len(jnsStatArr) == 0 {
		if expected == "DISCONNECTED" {
			return nil
		} else if expected == "SECONDARY" {
			return fmt.Errorf("Get None JNS info")
		}
		logger.Warning("Get None JNS info, ignore ...")
		return nil
	}
	logger.Debugf("JNS Stats:%v", jnsStatArr)
	for _, jnsStat := range jnsStatArr {
		if jnsStat["JState"] != expected {
			return fmt.Errorf("JStat not expected(act/exp): %s/%s", jnsStat, expected)
		}
	}
	return nil
}

func (d *dplMgr) IsJnsStatPrimary(vsetID int, anchorUUID, jnlUUID string) error {
	return d.IsJnsStatExpected(vsetID, anchorUUID, jnlUUID, "PRIMARY")
}

func (d *dplMgr) IsAnyJnsStatPrimary(vsetID int, anchorUUID, jnlUUID string) error {
	jnsStatArr, _ := d.GetJnsStat(vsetID, anchorUUID, jnlUUID)
	if len(jnsStatArr) == 0 {
		logger.Warning("Get None JNS info, ignore ...")
		return nil
	}
	logger.Debugf("JNS Stats:%v", jnsStatArr)
	for _, jnsStat := range jnsStatArr {
		if jnsStat["JState"] == "PRIMARY" {
			return nil
		}
		continue
	}
	return fmt.Errorf("None of JStat PRIMARY")
}

func (d *dplMgr) IsJnsStatSecondary(vsetID int, anchorUUID, jnlUUID string) error {
	return d.IsJnsStatExpected(vsetID, anchorUUID, jnlUUID, "SECONDARY")
}

func (d *dplMgr) IsJnsStatDisconnected(vsetID int, anchorUUID, jnlUUID string) error {
	return d.IsJnsStatExpected(vsetID, anchorUUID, jnlUUID, "DISCONNECTED")
}

// =============== Channel: dplmanager -m ch list ===============
func (d *dplMgr) PrintDplChannels(dplIP string, dplPort int) error {
	cmdSpec := fmt.Sprintf("%s -m ch -a %s -p %d list", d.DplmgrPath, dplIP, dplPort)
	_, output := d.RunCmdInDocker(cmdSpec)
	logger.Info(output)
	return nil
}

func (d *dplMgr) GetDplChannels(dplIP string, dplPort int) (chMapArr []map[string]string, err error) {
	cmdSpec := fmt.Sprintf("%s -m ch -a %s -p %d list", d.DplmgrPath, dplIP, dplPort)
	rc, output := d.RunCmdInDocker(cmdSpec)
	logger.Info(output)
	if rc != 0 {
		err = fmt.Errorf("Get jns status failed")
		return
	}
	chStrArr := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	pattern := regexp.MustCompile(`(.*)\s{2,}(((25[0-5]|2[0-4]\d|[01]?\d\d?)\.){3}(25[0-5]|2[0-4]\d|[01]?\d\d?))\s{2,}(\w+)\s{2,}(\w+)`)
	for _, chStr := range chStrArr {
		chMap := map[string]string{}
		matched := pattern.FindAllStringSubmatch(chStr, -1)
		if len(matched) > 0 {
			chMap["Channel_UUID"] = strings.TrimSpace(matched[0][1])
			chMap["Client_ip"] = strings.TrimSpace(matched[0][2])
			chMap["Channel_type"] = strings.TrimSpace(matched[0][6])
			chMap["PO_state"] = strings.TrimSpace(matched[0][7])
			chMap["dpl_ip"] = dplIP
			chMap["dpl_port"] = strconv.Itoa(dplPort)
			chMapArr = append(chMapArr, chMap)
		}
	}
	return
}
