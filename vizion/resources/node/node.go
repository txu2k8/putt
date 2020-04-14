package node

// Define OP on vizion node, by ssh to node and then run commands

import (
	"gtest/libs/sshmgr"
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

// SSHConnect connect to node sshd
func (n *Node) SSHConnect() {
	var err error
	n.Session, err = n.NewSessionWithRetry()
	if err != nil {
		logger.Fatal(err)
	}
}

// ZoolStatus ...
func (n *Node) ZoolStatus() map[string]string {
	zpoolStatus := make(map[string]string)
	cmdSpec := "zpool status"
	_, output := n.RunCmd(cmdSpec)
	logger.Infof(output)
	if strings.Contains(output, "no pools available") {
		return zpoolStatus
	}
	pattern := "\s+(\S+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)"
	for _, item := range strings.Split(output, "\n") {
		item = strings.Split(item, "\n")[0]
	}

	// if 'no pools available' in output:
	//     return zpool_status_dict

	// zpool_status_dict['config'] = []
	// pattern = r'\s+(\S+)\s+(\S+)\s+(\d+)\s+(\d+)\s+(\d+)'
	// for item in output.strip().split('\n'):
	//     item = item.strip('\n')
	//     if not item or any(k in item for k in ['config', 'NAME', 'scan']):
	//         continue
	//     if ':' in item:
	//         k, v = item.split(':')
	//         zpool_status_dict[k.strip()] = v.strip()
	//     else:
	//         cfg = re.search(pattern, item)
	//         config_dict = {}
	//         config_dict['NAME'] = cfg.group(1)
	//         config_dict['STATE'] = cfg.group(2)
	//         config_dict['READ'] = int(cfg.group(3))
	//         config_dict['WRITE'] = int(cfg.group(4))
	//         config_dict['CKSUM'] = int(cfg.group(5))
	//         zpool_status_dict['config'].append(config_dict)
	// logger.debug(zpool_status_dict)
	// return zpool_status_dict
}
