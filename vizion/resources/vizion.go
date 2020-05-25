package resources

import (
	"pzatest/libs/k8s"
	"pzatest/types"
	"strconv"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// BaseInterface ...
type BaseInterface interface {
	NodeGetter
	DplmanagerGetter
	ServiceManagerGetter
	CassClusterGetter
}

// VizionBase is used to interact with features provided by the  group.
type VizionBase struct {
	types.VizionBaseInput
	KubeConfig string // kubeconfig file path
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

// Cass returns CassCluster
func (b *VizionBase) Cass() CassCluster {
	return newSessCluster(b)
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

// GetCassConfig returns cassandra cluster configs
func (b *VizionBase) GetCassConfig() (cf map[string]CassConfig) {
	masterCassIPs := b.Service().GetMasterCassIPs()
	masterUser, masterPwd := b.Service().GetMasterCassUserPwd()
	masterPort := b.Service().GetMasterCassPort()
	cf["0"] = CassConfig{
		Index:    0,
		IPs:      masterCassIPs,
		User:     masterUser,
		Password: masterPwd,
		Port:     masterPort,
		Keyspace: "vizion",
	}
	for _, vsetID := range b.VizionBaseInput.VsetIDs {
		vsetCassIPs := b.Service().GetSubCassIPs(vsetID)
		vsetUser, vsetPwd := b.Service().GetSubCassUserPwd(vsetID)
		vsetPort := b.Service().GetSubCassPort(vsetID)
		cf[strconv.Itoa(vsetID)] = CassConfig{
			Index:    vsetID,
			IPs:      vsetCassIPs,
			User:     vsetUser,
			Password: vsetPwd,
			Port:     vsetPort,
			Keyspace: "vizion",
		}
	}
	return
}
