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
	"regexp"
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
	GetKubeConfig(localPath string) error
	GetKubeVipIP(fqdn string) (vIP string)
	GetEtcdCertsPathArr() []string
	GetEtcdCerts(localPath string) (localCertsPathArr []string, err error)
	PutEtcdCerts(localCertsPathArr []string) error

	GetCrashDirs() (crashArr []string)
	GetLogDirs(dirFilter []string) (logDirArr []string)
	CleanLog(dirFilter []string) error
	DeleteFiles(topPath string) error
	ChangeDplmanagerShellImage(image, dplmgrPath string) error
	IsDplmodExist() bool
	RmModDpl() error
	IsDplDeviceExist(devName string) bool
	WaitDplDeviceRemoved(devName string) error
}

// nodes implements NodeInterface
type node struct {
	*sshmgr.SSHMgr
}

// newNode returns a Nodes
func newNode(v *Vizion, host string) *node {
	return &node{sshmgr.NewSSHMgr(host, v.Base.SSHKey)}
}

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
		cmdSpec := fmt.Sprintf("find %s/* -maxdepth 1 -type d", crashPath)
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
