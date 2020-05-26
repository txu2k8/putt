package resources

import (
	"pzatest/libs/db"
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
	HealthCheckerGetter
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

// Check returns HealthChecker
func (b *VizionBase) Check() HealthChecker {
	return newHealthChecker(b)
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
func (b *VizionBase) GetCassConfig() map[string]db.CassConfig {
	masterCassIPs := b.Service().GetMasterCassIPs()
	masterUser, masterPwd := b.Service().GetMasterCassUserPwd()
	masterPort := b.Service().GetMasterCassPort()
	cf := map[string]db.CassConfig{}
	cf["0"] = db.CassConfig{
		Hosts:    masterCassIPs,
		Username: masterUser,
		Password: masterPwd,
		Port:     masterPort,
		Keyspace: "vizion",
	}
	for _, vsetID := range b.VizionBaseInput.VsetIDs {
		vsetCassIPs := b.Service().GetSubCassIPs(vsetID)
		vsetUser, vsetPwd := b.Service().GetSubCassUserPwd(vsetID)
		vsetPort := b.Service().GetSubCassPort(vsetID)
		cf[strconv.Itoa(vsetID)] = db.CassConfig{
			Hosts:    vsetCassIPs,
			Username: vsetUser,
			Password: vsetPwd,
			Port:     vsetPort,
			Keyspace: "vizion",
		}
	}
	// logger.Infof("CassConfig:%s\n", utils.Prettify(cf))
	return cf
}
