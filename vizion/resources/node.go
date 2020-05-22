package resources

// Define OP on vizion node, by ssh to node and then run commands

import (
	"os"
	"path"
	"pzatest/libs/sshmgr"
	"strings"
)

// NodeGetter has a method to return a NodeInterface.
// A group's client should implement this interface.
type NodeGetter interface {
	Node(host string) NodeInterface
}

// NodeInterface has methods to work on Node resources.
type NodeInterface interface {
	GetKubeConfig(bool) (string, error)
	IsDplmodExist() bool
}

// nodes implements NodeInterface
type node struct {
	ssh sshmgr.SSHInput
}

// newNode returns a Nodes
func newNode(b *VizionBase, host string) *node {
	return &node{
		ssh: sshmgr.SSHInput{Host: host, SSHKey: b.SSHKey},
	}
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
		err = n.ssh.SCPGet(cfPath, remoteCf)
	}

	return cfPath, err
}

// IsDplmodExist ...
func (n *node) IsDplmodExist() bool {
	cmdSpec := "lsmod | grep dpl"
	_, output := n.ssh.RunCmd(cmdSpec)
	logger.Info(output)
	if output != "" && strings.Contains(output, "dpl") {
		return true
	}
	return false
}
