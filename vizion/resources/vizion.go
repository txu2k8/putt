package resources

import (
	"pzatest/libs/k8s"
	"pzatest/types"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// BaseInterface ...
type BaseInterface interface {
	NodeGetter
	DplmanagerGetter
	ServiceManagerGetter
}

// VizionBase is used to interact with features provided by the  group.
type VizionBase struct {
	types.VizionBaseInput
}

// Node returns NodeInterface
func (b *VizionBase) Node(host string) NodeInterface {
	return newNode(b, host)
}

// DplMgr returns Dplmanager
func (b *VizionBase) DplMgr(host string) Dplmanager {
	return newdplMgr(b, host)
}

// Service returns ServiceManager
func (b *VizionBase) Service() ServiceManager {
	return newServiceMgr(b)
}

// GetK8sClient returns a k8s.Client that is used to communicate
// with K8S API server by this client implementation.
func (b *VizionBase) GetK8sClient() k8s.Client {
	c, err := k8s.NewClientWithRetry(b.VizionBaseInput.KubeConfig)
	if err != nil {
		panic(err)
	}
	c.NameSpace = b.VizionBaseInput.K8sNameSpace
	return c
}

// GetMasterCassClient returns a cassandra Client that is used to communicate
// with master cassandra server by this client implementation.
// TODO
func (b *VizionBase) GetMasterCassClient() k8s.Client {
	c, err := k8s.NewClientWithRetry(b.VizionBaseInput.KubeConfig)
	if err != nil {
		panic(err)
	}
	c.NameSpace = b.VizionBaseInput.K8sNameSpace
	return c
}
