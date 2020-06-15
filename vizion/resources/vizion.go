package resources

import (
	"fmt"
	"pzatest/libs/db"
	"pzatest/libs/k8s"
	"pzatest/libs/utils"
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

// Vizion is used to interact with features provided by the  group.
type Vizion struct {
	Base         types.VizionBaseInput    // command/args input
	CassConfig   map[string]db.CassConfig // cassandra configs map
	KubeConfig   string                   // kubeconfig file path
	K8sNameSpace string                   // k8s namespace
}

// Node returns NodeInterface
func (v *Vizion) Node(host string) NodeInterface {
	return newNode(v, host)
}

// VaildMasterIP returns Vaild MasterIP --> Ping OK
func (v *Vizion) VaildMasterIP() string {
	for _, masterIP := range v.Base.MasterIPs {
		err := utils.IsPingOK(masterIP)
		if err != nil {
			continue
		}
		return masterIP
	}
	panic(fmt.Sprintf("All MasterIPs Ping failed: %v", v.Base.MasterIPs))
}

// MasterNode returns NodeInterface with host=masterIP
func (v *Vizion) MasterNode() NodeInterface {
	masterIP := v.VaildMasterIP()
	return newNode(v, masterIP)
}

// DplMgr returns Dplmanager
func (v *Vizion) DplMgr(host string) Dplmanager {
	return newdplMgr(v, host)
}

// Service returns ServiceManager
func (v *Vizion) Service() ServiceManager {
	return newServiceMgr(v)
}

// Cass returns CassCluster
func (v *Vizion) Cass() CassCluster {
	return newSessCluster(v)
}

// GetK8sClient returns a k8s.Client that is used to communicate
// with K8S API server by this client implementation.
func (v *Vizion) GetK8sClient() k8s.Client {
	v.GetKubeConfig()
	c, err := k8s.NewClientWithRetry(v.KubeConfig)
	if err != nil {
		panic(err)
	}
	c.NameSpace = v.Base.K8sNameSpace
	return c
}

// GetCassConfig returns cassandra cluster configs
func (v *Vizion) GetCassConfig() map[string]db.CassConfig {
	// Get once
	if v.CassConfig != nil {
		return v.CassConfig
	}
	masterCassIPs := v.Service().GetMasterCassIPs()
	masterUser, masterPwd := v.Service().GetMasterCassUserPwd()
	masterPort := v.Service().GetMasterCassPort()
	// master --> vset1
	// masterCassIPs := v.Service().GetSubCassIPs(1)
	// masterUser, masterPwd := v.Service().GetSubCassUserPwd(1)
	// masterPort := v.Service().GetSubCassPort(1)

	cf := map[string]db.CassConfig{}
	cf["0"] = db.CassConfig{
		Hosts:    masterCassIPs,
		Username: masterUser,
		Password: masterPwd,
		Port:     masterPort,
		Keyspace: "vizion",
	}
	for _, vsetID := range v.Base.VsetIDs {
		vsetCassIPs := v.Service().GetSubCassIPs(vsetID)
		vsetUser, vsetPwd := v.Service().GetSubCassUserPwd(vsetID)
		vsetPort := v.Service().GetSubCassPort(vsetID)
		cf[strconv.Itoa(vsetID)] = db.CassConfig{
			Hosts:    vsetCassIPs,
			Username: vsetUser,
			Password: vsetPwd,
			Port:     vsetPort,
			Keyspace: "vizion",
		}
	}
	// logger.Infof("CassConfig:%s\n", utils.Prettify(cf))
	v.CassConfig = cf
	return cf
}
