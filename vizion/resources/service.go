package resources

import (
	"pzatest/libs/k8s"
)

// ServiceManagerGetter has a method to return a ServiceManager.
type ServiceManagerGetter interface {
	Service() ServiceManager
}

// ServiceManager ...
type ServiceManager interface {
	K8sEnableNodeLabel(nodeName, nodeLabel, podLabel string) error
	// K8sDisableNodeLabel(nodeName, nodeLabel, podLabel string) error
	// K8sEnableNodeLabelByType(serviceType int) error
	// K8sDisableNodeLabelByType(serviceType int) error
	// K8sStartAll(serviceType int) error
	// K8sShutdownAll(serviceType int) error
}

// svManager implements NodeInterface
type svManager struct {
	k8sclient k8s.Client
}

// newServiceMgr returns a Nodes
func newServiceMgr(b *VizionBase) *svManager {
	return &svManager{
		k8sclient: b.GetK8sClient(),
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

func (s *svManager) K8sEnableNodeLabel(nodeName, nodeLabel, podLabel string) error {
	img, _ := s.k8sclient.GetPodImage("cmapmcdpl-1-0", "cmapmcdpl")
	logger.Info(img)
	return nil
}
