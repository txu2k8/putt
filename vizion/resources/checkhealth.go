package resources

import (
	"fmt"
	"pzatest/libs/k8s"
)

// HealthCheckerGetter has a method to return a HealthChecker.
type HealthCheckerGetter interface {
	HealthCheck() HealthChecker
}

// HealthChecker ...
type HealthChecker interface {
	IsPingOK(ip string) error
}

// checker implements HealthChecker Interface
type checker struct {
	k8sclient k8s.Client
}

// newHealthChecker returns a Nodes
func newHealthChecker(b *VizionBase) *checker {
	return &checker{
		k8sclient: b.GetK8sClient(),
	}
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
