package resources

import (
	"fmt"
	"pzatest/libs/k8s"
	"pzatest/libs/utils"
)

// ServiceManagerGetter has a method to return a ServiceManager.
type ServiceManagerGetter interface {
	Service() ServiceManager
}

// ServiceManager ...
type ServiceManager interface {
	GetMasterCassIPs() (ipArr []string)
	GetMasterCassUserPwd() (user, pwd string)
	GetMasterCassPort() (port int)

	GetSubCassIPs(vsetID int) (ipArr []string)
	GetSubCassUserPwd(vsetID int) (user, pwd string)
	GetSubCassPort(vsetID int) (port int)

	K8sEnableNodeLabel(nodeName, nodeLabel, podLabel string) error
	// K8sDisableNodeLabel(nodeName, nodeLabel, podLabel string) error
	// K8sEnableNodeLabelByType(serviceType int) error
	// K8sDisableNodeLabelByType(serviceType int) error
	// K8sStartAll(serviceType int) error
	// K8sShutdownAll(serviceType int) error
	Test(podLabel string) error
}

// svManager implements NodeInterface
type svManager struct {
	k8s.Client
}

// newServiceMgr returns a Nodes
func newServiceMgr(b *VizionBase) *svManager {
	return &svManager{
		b.GetK8sClient(),
	}
}

// GetKubeConfig ...
// func (b *svManager) GetKubeConfig() (cfPath string, err error) {
// 	n := b.Node(b.MasterIPs[0])
// 	cfPath, err = n.GetKubeConfig(false)
// 	if err != nil {
// 		panic(err)
// 	}
// 	b.KubeConfig = cfPath
// 	return cfPath, err
// }

// K8sGetMasterCassIPs ...
func (s *svManager) GetMasterCassIPs() (ipArr []string) {
	ipArr, err := s.GetSvcIPs("cassandra-master-expose")
	if err != nil {
		panic(err)
	}
	return
}

// K8sGetMasterCassUserPwd ...
func (s *svManager) GetMasterCassUserPwd() (user, pwd string) {
	scrt, err := s.GetSecretDetail("cassandra-config")
	if err != nil {
		panic(err)
	}
	for k, v := range scrt.Data {
		strV := string(v[:])
		switch k {
		case "CASUser":
			user = strV // utils.Base64Encode(v)
		case "CASPwd":
			pwd = strV // utils.Base64Encode(v)
		}
	}
	return
}

// K8sGetMasterCassIPs ...
func (s *svManager) GetMasterCassPort() (port int) {
	port, err := s.GetSvcPort("cassandra-master-expose", 9042)
	if err != nil {
		panic(err)
	}
	return
}

// GetSubCassIPs ...
func (s *svManager) GetSubCassIPs(vsetID int) (ipArr []string) {
	ipArr, err := s.GetSvcIPs(fmt.Sprintf("cassandra-vset%d-expose", vsetID))
	if err != nil {
		panic(err)
	}
	return
}

// K8sGetMasterCassUserPwd ...
func (s *svManager) GetSubCassUserPwd(vsetID int) (user, pwd string) {
	scrt, err := s.GetSecretDetail(fmt.Sprintf("cassandra-config-vset%d", vsetID))
	if err != nil {
		panic(err)
	}
	for k, v := range scrt.Data {
		strV := string(v[:])
		switch k {
		case "CASUser":
			user = strV // utils.Base64Encode(v)
		case "CASPwd":
			pwd = strV // utils.Base64Encode(v)
		}
	}
	return
}

// K8sGetMasterCassIPs ...
func (s *svManager) GetSubCassPort(vsetID int) (port int) {
	port, err := s.GetSvcPort(fmt.Sprintf("cassandra-vset%d-expose", vsetID), 9042)
	if err != nil {
		panic(err)
	}
	return
}

// K8sEnableNodeLabel .
func (s *svManager) K8sEnableNodeLabel(nodeName, nodeLabel, podLabel string) error {
	img, _ := s.GetPodImage("cmapmcdpl-1-0", "cmapmcdpl")
	logger.Info(img)
	return nil
}

// Test .
func (s *svManager) Test(podLabel string) error {
	// nameArr, _ := s.GetPodNameListByLabel(podLabel)
	// logger.Info(utils.Prettify(nameArr))
	u, p := s.GetMasterCassUserPwd()
	logger.Info(utils.Prettify(u))
	logger.Info(utils.Prettify(p))
	port := s.GetMasterCassPort()
	logger.Info(utils.Prettify(port))
	return nil
}
