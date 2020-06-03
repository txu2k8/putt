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
	*Vizion
}

// newHealthChecker returns a Nodes
func newHealthChecker(v *Vizion) *checker {
	return &checker{v}
}

func (c *checker) IsNodeCrashed() error {
	crashed := false
	crashArrMap := map[string][]string{}
	for _, nodeIP := range c.Vizion.Service().GetAllNodeIPs() {
		node := c.Vizion.Node(nodeIP)
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
