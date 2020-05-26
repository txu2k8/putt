package resources

import (
	"fmt"
	"pzatest/libs/utils"
)

// HealthCheckerGetter has a method to return a HealthChecker.
type HealthCheckerGetter interface {
	Check() HealthChecker
}

// HealthChecker ...
type HealthChecker interface {
	IsNodeCrashed() error
}

// checker implements HealthChecker Interface
type checker struct {
	*VizionBase
}

// newHealthChecker returns a Nodes
func newHealthChecker(b *VizionBase) *checker {
	return &checker{b}
}

// IsPingOK ...
func IsPingOK(ip string) error {
	var cmd string
	sysstr := ""
	switch sysstr {
	case "Windows":
		cmd = fmt.Sprintf("ping %s", ip)
	case "Linux":
		cmd = fmt.Sprintf("ping -c1 %s", ip)
	default:
		cmd = fmt.Sprintf("ping %s", ip)
	}
	logger.Info(cmd)
	return nil
}

func (c *checker) IsNodeCrashed() error {
	crashed := false
	crashArrMap := map[string][]string{}
	for _, nodeIP := range c.VizionBase.Service().GetAllNodeIPs() {
		node := c.VizionBase.Node(nodeIP)
		crashArr := node.GetCrashDirs()
		if len(crashArr) > 0 {
			crashed = true
			crashArrMap[nodeIP] = crashArr
			logger.Errorf("%s has crash files\n%s", nodeIP, utils.Prettify(crashArr))
		}
	}
	if crashed == true {
		return fmt.Errorf("Some node has crash files\n%s", utils.Prettify(crashArrMap))
	}
	return nil
}
