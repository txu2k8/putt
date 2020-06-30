package resources

import (
	"fmt"
	"putt/config"
	"putt/libs/db"
	"putt/libs/k8s"
	"putt/libs/runner/schedule"
	"putt/libs/utils"
	"putt/types"
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
	Schedule     schedule.Schedule        // Schedule
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
	podLabel := config.Servicedpl.GetPodLabel(v.Base)
	svMgr := v.Service()
	k8sNameArr, _ := svMgr.GetStatefulSetsNameArrByLabel(podLabel)
	dplImage, _ := svMgr.GetStatefulSetsImage(k8sNameArr[0], config.Servicedpl.Container)
	return newdplMgr(v, host, dplImage)
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

	cf := map[string]db.CassConfig{}
	vk8s := v.Service()
	/* // masterCass -> SubCass-1
	masterCassIPs := vk8s.GetMasterCassIPs()
	masterUser, masterPwd := vk8s.GetMasterCassUserPwd()
	masterPort := vk8s.GetMasterCassPort()
	cf["0"] = db.CassConfig{
		Hosts:    masterCassIPs,
		Username: masterUser,
		Password: masterPwd,
		Port:     masterPort,
		Keyspace: "vizion",
	}
	*/
	for _, vsetID := range v.Base.VsetIDs {
		vsetCassIPs := vk8s.GetSubCassIPs(vsetID)
		vsetUser, vsetPwd := vk8s.GetSubCassUserPwd(vsetID)
		vsetPort := vk8s.GetSubCassPort(vsetID)
		cf[strconv.Itoa(vsetID)] = db.CassConfig{
			Hosts:    vsetCassIPs,
			Username: vsetUser,
			Password: vsetPwd,
			Port:     vsetPort,
			Keyspace: "vizion",
		}
		if vsetID == 1 {
			cf["0"] = cf["1"]
		}
	}
	// logger.Infof("CassConfig:%s\n", utils.Prettify(cf))
	v.CassConfig = cf
	return cf
}
