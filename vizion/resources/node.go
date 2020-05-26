package resources

// Define OP on vizion node, by ssh to node and then run commands

import (
	"fmt"
	"os"
	"path"
	"pzatest/libs/sshmgr"
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
	GetKubeConfig(bool) (string, error)
	GetCrashDirs() (crashArr []string)
	GetLogDirs(dirFilter []string) (logDirArr []string)
	CleanLog(dirFilter []string) error
	IsDplmodExist() bool
}

// nodes implements NodeInterface
type node struct {
	sshmgr.SSHMgr
}

// newNode returns a Nodes
func newNode(b *VizionBase, host string) *node {
	sshCfg := sshmgr.NewSSHConfig(host, b.SSHKey)
	session, err := sshCfg.CreateSession()
	if err != nil {
		panic(err)
	}
	return &node{sshmgr.SSHMgr{session, sshCfg}}
}

// GetKubeConfig ...
func (n *node) GetKubeConfig(overwrite bool) (cfPath string, err error) {
	remoteCf := "/root/.kube/config"
	localDir := "/tmp"
	_, err = os.Stat(cfPath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(cfPath, os.ModePerm)
		if err != nil {
			logger.Panicf("mkdir failed![%v]", err)
		}
	}

	cfPath = path.Join(localDir, "config")
	if overwrite {
		err = n.SCPGet(cfPath, remoteCf)
	}

	return cfPath, err
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
