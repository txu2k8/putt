package node

// Define OP on vizion node, by ssh to node and then run commands

import (
	"pzatest/libs/sshmgr"
	"strings"

	"github.com/op/go-logging"
	"golang.org/x/crypto/ssh"
)

var logger = logging.MustGetLogger("test")

// Node ...
type Node struct {
	sshmgr.SSHInput
	Session *ssh.Session // ssh to node session
}

/*
// ZoolStatus ...
func (n *Node) ZoolStatus() []map[string]map[string]string {
	zpoolConfig := []map[string]map[string]string{}
	cmdSpec := "zpool status"
	_, output := n.RunCmd(cmdSpec)
	logger.Infof(output)
	if strings.Contains(output, "no pools available") {
		return zpoolConfig
	}
	ignoreKey := []string{"config", "NAME", "scan"}
	reg := regexp.MustCompile(`\s+(\S+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)`)
	for _, item := range strings.Split(output, "\n") {
		item = strings.Split(item, "\n")[0]
		logger.Info(item)
		if item == "" {
			continue
		}
		skip := false
		for _, k := range ignoreKey {
			if strings.Contains(item, k) {
				skip = true
			}
		}
		if skip == true {
			continue
		}

		if strings.Contains(item, ":") {
			kv := strings.Split(item, ":")
			k := strings.Split(kv[0], "\n")[0]
			v := strings.Split(kv[1], "\n")[0]
			zpoolStatus[k] = v
		} else {
			matched := reg.FindStringSubmatch(item)
			config := make(map[string]string)
			config["NAME"] = matched[1]
			config["STATE"] = matched[1]
			config["READ"] = matched[1]
			config["WRITE"] = matched[1]
			config["CKSUM"] = matched[1]
			zpoolConfig = append(zpoolConfig, config)
		}
	}
	return zpoolStatus
}
*/

// IsDplmodExist ...
func (n *Node) IsDplmodExist() bool {
	cmdSpec := "lsmod | grep dpl"
	_, output := n.RunCmd(cmdSpec)
	logger.Info(output)
	if output != "" && strings.Contains(output, "dpl") {
		return true
	}
	return false
}
