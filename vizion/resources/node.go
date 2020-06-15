package resources

// Define OP on vizion node, by ssh to node and then run commands

import (
	"fmt"
	"os"
	"path"
	"pzatest/libs/retry"
	"pzatest/libs/retry/strategy"
	"pzatest/libs/sshmgr"
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
