package resources

// Define OP on vizion node, by ssh to node and then run commands

import (
	"fmt"
	"os"
	"path"
	"pzatest/libs/sshmgr"
	"regexp"
	"strings"

	"github.com/chenhg5/collection"
)

// NodeGetter has a method to return a NodeInterface.
// A group's client should implement this interface.
type NodeGetter interface {
	Node(host string) NodeInterface
}

// NodeInterface has methods to work on Node resources.
type NodeInterface interface {
	GetKubeConfig(localPath string) error
	GetKubeVipIP(fqdn string) (vIP string)
	GetCrashDirs() (crashArr []string)
	GetLogDirs(dirFilter []string) (logDirArr []string)
	CleanLog(dirFilter []string) error
	IsDplmodExist() bool
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
