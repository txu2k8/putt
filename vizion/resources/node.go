package resources

// Define OP on vizion node, by ssh to node and then run commands

import (
	"fmt"
	"os"
	"path"
	"putt/config"
	"putt/libs/retry"
	"putt/libs/retry/strategy"
	"putt/libs/sshmgr"
	"putt/libs/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chenhg5/collection"
)

// NodeGetter has a method to return a NodeInterface.
// A group's client should implement this interface.
type NodeGetter interface {
	Node(host string) NodeInterface
}

// NodeInterface has methods to work on Node resources.
type NodeInterface interface {
	sshmgr.SSHManager
	// kube/etcd
	GetKubeConfig(localPath string) error
	GetKubeVipIP(fqdn string) (vIP string)
	GetEtcdCertsPathArr() []string
	GetEtcdCerts(localPath string) (localCertsPathArr []string, err error)
	PutEtcdCerts(localCertsPathArr []string) error
	GetEtcdMembers() []string
	GetEtcdEndpoints() []string
	PrintDplBalance() error
	// logs/path
	GetCrashDirs() (crashArr []string)
	GetCoreFiles(dirFilter []string) (coreArr []string)
	GetLogDirs(dirFilter []string) (logDirArr []string)
	CleanLog(dirFilter []string) error
	DeleteFiles(topPath string) error
	ChangeDplmanagerShellImage(image, dplmgrPath string) error
	// bd dplmod/device.zpool
	IsDplmodExist() bool
	RmModDpl() error
	IsDplDeviceExist(devName string) bool
	WaitDplDeviceRemoved(devName string) error
	GetJdeviceStgUnitNumber(jdevice string) (stgUnitNum int64)
	GetZpoolStatus() (zs *ZpoolStatus)
	IsZpoolStatusOK() error
}

// nodes implements NodeInterface
type node struct {
	*sshmgr.SSHMgr
}

// newNode returns a Nodes
func newNode(v *Vizion, host string) *node {
	return &node{sshmgr.NewSSHMgr(host, v.Base.SSHKey)}
}

// =============== Kube: ./kube/config ===============
// GetKubeConfig ...
func (n *node) GetKubeConfig(localPath string) error {
	remoteCf := "/root/.kube/config"

	localDir := strings.Split(localPath, path.Base(localPath))[0]
	_, err := os.Stat(localDir)
	if os.IsNotExist(err) {
		err := os.MkdirAll(localDir, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	n.ConnectSftpClient()
	err = n.ScpGet(localPath, remoteCf)

	return err
}

// GetKubeVipIP .
// fqdn: "kubernetes.vizion.local"
func (n *node) GetKubeVipIP(fqdn string) (vIP string) {
	re := regexp.MustCompile(`(([01]{0,1}\d{0,1}\d|2[0-4]\d|25[0-5])\.){3}([01]{0,1}\d{0,1}\d|2[0-4]\d|25[0-5])`)
	cmdSpec := fmt.Sprintf("ping %s -c 1", fqdn)
	_, output := n.RunCmd(cmdSpec)
	matched := re.FindAllStringSubmatch(output, -1)
	// logger.Info(utils.Prettify(matched))
	for _, match := range matched {
		vIP = match[0]
	}
	return
}

// =============== ETCD: cert / etcdctlv3 ===============
func (n *node) GetEtcdCertsPathArr() []string {
	var certs = []string{}
	cmdSpec := "find " + path.Join(config.EtcdCertPath, "*")
	_, output := n.RunCmd(cmdSpec)
	if strings.Contains(output, "No such file or directory") {
		cmdMkdir := "mkdir -p " + config.EtcdCertPath
		n.RunCmd(cmdMkdir)
	}
	certs = append(certs, strings.Split(strings.TrimSuffix(output, "\n"), "\n")...)
	return certs
}

func (n *node) GetEtcdCerts(localPath string) (localCertsPathArr []string, err error) {
	localCertTopPath := path.Join(localPath, "etcd", n.Cfg.Host)
	_, err = os.Stat(localCertTopPath)
	if os.IsNotExist(err) {
		err = os.MkdirAll(localCertTopPath, os.ModePerm)
		if err != nil {
			return
		}
	}

	certsPathArr := n.GetEtcdCertsPathArr()
	for _, cert := range certsPathArr {
		certPath := strings.TrimSuffix(cert, "\n")
		certName := path.Base(certPath)
		localCertPath := path.Join(localCertTopPath, certName)
		localCertsPathArr = append(localCertsPathArr, localCertPath)

		_, err = os.Stat(localCertPath)
		if os.IsNotExist(err) {
			if n.Cfg.UserName != "root" {
				tmpPath := "/tmp/" + certName
				n.RunCmd(fmt.Sprintf("cp %s %s", certPath, tmpPath))
				n.RunCmd(fmt.Sprintf("chmod 666 %s", tmpPath))
				n.RunCmd(fmt.Sprintf("chmod -R %s:%s %s", n.Cfg.UserName, n.Cfg.UserName, tmpPath))
				certPath = tmpPath
			}
			n.ConnectSftpClient()
			err = n.ScpGet(localCertPath, certPath)
			if err != nil {
				return
			}
		}
	}
	return
}

func (n *node) PutEtcdCerts(localCertsPathArr []string) error {
	var err error
	n.ConnectSftpClient()
	for _, cert := range localCertsPathArr {
		certName := path.Base(cert)
		n.MkdirIfNotExist(config.EtcdCertPath)
		remotePath := config.EtcdCertPath + certName
		if n.Cfg.UserName == "root" {
			err = n.ScpPut(cert, remotePath)
			if err != nil {
				return err
			}
		} else {
			tmpPath := "/tmp/" + certName
			err = n.ScpPut(cert, tmpPath)
			if err != nil {
				return err
			}
			n.RunCmd(fmt.Sprintf("cp %s %s", tmpPath, remotePath))
		}
	}
	return nil
}

func (n *node) GetEtcdMembers() []string {
	cmdSpec := "etcdctlv3 member list"
	_, output := n.RunCmd(cmdSpec)
	logger.Infof("\n%s", output)
	members := strings.Split(strings.TrimSuffix(output, "\n"), "\n")
	return members
}

func (n *node) GetEtcdEndpoints() []string {
	endPoints := []string{}
	members := n.GetEtcdMembers()
	pattern := regexp.MustCompile(`https://(\S+:2379)`)
	for _, member := range members {
		matched := pattern.FindAllStringSubmatch(member, -1)
		if len(matched) > 0 {
			endPoints = append(endPoints, matched[0][1])
		}
	}
	return endPoints
}

func (n *node) PrintDplBalance() error {
	cmdSpec := "etcdctlv3 get --prefix /vizion/dpl/balance/"
	_, output := n.RunCmd(cmdSpec)
	logger.Infof("\n%s", output)
	return nil
}

// =============== Path/Log: crash/core dump/logs ===============
func (n *node) MkdirIfNotExist(remotePath string) error {
	cmdSpec := "ls " + remotePath
	_, output := n.RunCmd(cmdSpec)
	if strings.Contains(output, "No such file or directory") {
		cmdMkdir := "mkdir -p " + remotePath
		n.RunCmd(cmdMkdir)
	}
	return nil
}

func (n *node) GetCrashDirs() (crashArr []string) {
	crashPathArr := []string{
		"/var/crash/",
	}

	for _, crashPath := range crashPathArr {
		cmdSpec := fmt.Sprintf("ls %s", crashPath)
		_, output := n.RunCmd(cmdSpec)
		if strings.Contains(output, "No such file or directory") {
			continue
		}
		crashList := strings.Split(strings.Trim(output, "\n"), "\n")
		index := 0
		for index < len(crashList) {
			if crashList[index] == "" {
				crashList = append(crashList[:index], crashList[index+1:]...)
				continue
			}
			index++
		}
		crashArr = append(crashArr, crashList...)
	}
	return
}

func (n *node) GetCoreFiles(dirFilter []string) (coreArr []string) {
	logger.Info("Find core files ...")
	logDirs := n.GetLogDirs(dirFilter)
	for _, logDir := range logDirs {
		cmdSpec := fmt.Sprintf("find %s -name 'core.*'", logDir)
		_, output := n.RunCmd(cmdSpec)
		if output != "" {
			coreArr = append(coreArr, strings.Split(strings.TrimSuffix(output, "\n"), "\n")...)
		}
	}
	return
}

func (n *node) GetLogDirs(dirFilter []string) (logDirArr []string) {
	logPathArr := []string{
		"/var/log/",
	}

	for _, logPath := range logPathArr {
		cmdSpec := fmt.Sprintf("find %s -maxdepth 1 -type d", logPath)
		_, output := n.RunCmd(cmdSpec)
		if strings.Contains(output, "No such file or directory") {
			continue
		}

		logDirs := strings.Split(strings.Trim(output, "\n"), "\n")
		for _, logDir := range logDirs {
			dirBase := path.Base(logDir)
			if collection.Collect(dirFilter).Contains(dirBase) {
				logDirArr = append(logDirArr, logDir)
			}
		}
	}
	return
}

func (n *node) CleanLog(dirFilter []string) error {
	logDirs := n.GetLogDirs(dirFilter)
	for _, logDir := range logDirs {
		cmdSpec := fmt.Sprintf("rm -rf %s/*", logDir)
		n.RunCmd(cmdSpec)
	}
	return nil
}

func (n *node) DeleteFiles(topPath string) error {
	if !strings.HasSuffix(topPath, "/") {
		topPath += "/"
	}

	cmdSpec1 := fmt.Sprintf("ls -1 %s | awk '{print i$0}' i='%s' | grep -v lost+found | xargs rm -rf", topPath, topPath)
	_, output := n.RunCmd(cmdSpec1)
	logger.Debug(output)

	cmdSpec2 := fmt.Sprintf("ls -l %s", topPath)
	_, output = n.RunCmd(cmdSpec2)
	logger.Info(output)
	return nil
}

func (n *node) ChangeDplmanagerShellImage(image, dplmgrPath string) error {
	tmpImage := strings.Replace(image, "/", `\/`, -1)
	n.RunCmd(fmt.Sprintf("chmod 777 %s", dplmgrPath))

	cmdSpec := fmt.Sprintf("sed -ri 's/registry.vizion.\\S+/%s/g' %s", tmpImage, dplmgrPath)
	_, output := n.RunCmd(cmdSpec)
	logger.Info(output)

	n.RunCmd(fmt.Sprintf("chmod 544 %s", dplmgrPath))

	_, output = n.RunCmd(fmt.Sprintf("cat %s", dplmgrPath))
	logger.Debug(output)
	if !strings.Contains(output, image) {
		return fmt.Errorf("Change %s image failed", dplmgrPath)
	}
	return nil
}

// =============== BD: dplmod/device/zpool ===============
// IsDplmodExist ...
func (n *node) IsDplmodExist() bool {
	cmdSpec := "lsmod | grep dpl"
	_, output := n.RunCmd(cmdSpec)
	logger.Info(output)
	if output != "" && strings.Contains(output, "dpl") {
		return true
	}
	return false
}

// RmModDpl "rmmod dpl" on node
func (n *node) RmModDpl() error {
	if n.IsDplmodExist() == false {
		return nil
	}

	cmdSpec := "rmmod dpl"
	rc, output := n.RunCmd(cmdSpec)
	logger.Infof("%d,%s", rc, output)
	if n.IsDplmodExist() == false {
		logger.Infof("PASS：rmmod dpl ---> %s", n.Cfg.Host)
		return nil
	}
	logger.Errorf("FAIL：rmmod dpl ---> %s", n.Cfg.Host)
	return fmt.Errorf("FAIL: rmmod dpl!(A bug or need reset node)")
}

// IsDplDeviceExist
func (n *node) IsDplDeviceExist(devName string) bool {
	if devName == "" {
		devName = "/dev/dpl*"
	}
	cmdSpec := fmt.Sprintf("ls -lh %s | grep -v /dev/dpl0", devName)
	rc, output := n.RunCmd(cmdSpec)
	logger.Infof("%d,%s", rc, output)

	if strings.Contains(output, "No such file or directory") || !strings.Contains(output, "/dev/dpl") {
		logger.Infof("%s not exist on node %s", devName, n.Cfg.Host)
		return false
	}
	return true
}

func (n *node) WaitDplDeviceRemoved(devName string) error {
	action := func(attempt uint) error {
		if n.IsDplDeviceExist(devName) == true {
			return fmt.Errorf("%s still exist", devName)
		}
		return nil
	}
	err := retry.Retry(
		action,
		strategy.Limit(30),
		strategy.Wait(30*time.Second),
	)
	return err
}

func (n *node) GetJdeviceStgUnitNumber(jdevice string) (stgUnitNum int64) {
	cmdSpec := fmt.Sprintf("fdisk -l %s | grep Disk | awk -F ',' '{print $2}' | awk '{print $1}'", jdevice)
	_, output := n.RunCmd(cmdSpec)
	sBytes, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimRight(output, "\n")), 10, 64)
	stgUnitNum = sBytes/1024/1024/1024/2 - 1
	return
}

type zpoolStatusConfig struct {
	NAME  string
	STATE string
	READ  int
	WRITE int
	CKSUM int
}

// ZpoolStatus .
type ZpoolStatus struct {
	Pool   string
	State  string
	Status string
	Action string
	Scan   string
	Errors string
	Config []zpoolStatusConfig
}

func (n *node) GetZpoolStatus() (zs *ZpoolStatus) {
	cmdSpec := "zpool status"
	_, output := n.RunCmd(cmdSpec)
	logger.Info(utils.Prettify(output))
	if strings.Contains(output, "no pools available") {
		return
	}
	pattern := regexp.MustCompile(`\s+(\S+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	for _, item := range strings.Split(output, "\n") {
		item = strings.TrimSuffix(item, "\n")
		if item == "" || strings.Contains(item, "config") ||
			strings.Contains(item, "NAME") ||
			strings.Contains(item, "scan") {
			continue
		}

		if strings.Contains(item, ":") {
			kv := strings.Split(item, ":")
			switch kv[0] {
			case "pool":
				zs.Pool = kv[1]
			case "state":
				zs.State = kv[1]
			case "status":
				zs.Status = kv[1]
			case "action":
				zs.Action = kv[1]
			case "errors":
				zs.Errors = kv[1]
			}
		} else {
			matched := pattern.FindAllStringSubmatch(item, -1)
			if len(matched) > 0 {
				matchCfg := matched[0]
				read, _ := strconv.Atoi(matchCfg[3])
				write, _ := strconv.Atoi(matchCfg[4])
				cksum, _ := strconv.Atoi(matchCfg[5])
				cfg := zpoolStatusConfig{
					NAME:  matchCfg[1],
					STATE: matchCfg[2],
					READ:  read,
					WRITE: write,
					CKSUM: cksum,
				}
				zs.Config = append(zs.Config, cfg)
			}
		}
	}
	return
}

func (n *node) IsZpoolStatusOK() error {
	zs := n.GetZpoolStatus()
	if zs == nil {
		return nil
	}
	if !strings.Contains(zs.Errors, "No known data errors") {
		return fmt.Errorf("zpool errors: %s", zs.Errors)
	}
	if zs.State != "ONLINE" {
		return fmt.Errorf("zpool state: %s", zs.State)
	}

	for _, cfg := range zs.Config {
		if cfg.STATE != "ONLINE" {
			return fmt.Errorf("zpool config: %v", cfg)
		}
		if cfg.READ != 0 {
			return fmt.Errorf("zpool config: %v", cfg)
		}
		if cfg.WRITE != 0 {
			return fmt.Errorf("zpool config: %v", cfg)
		}
		if cfg.CKSUM != 0 {
			return fmt.Errorf("zpool config: %v", cfg)
		}
	}
	return nil
}
