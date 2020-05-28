package resources

import (
	"fmt"
	"pzatest/libs/k8s"
	"pzatest/libs/utils"
	"strings"
	"sync"
	"time"
)

// ServiceManagerGetter has a method to return a ServiceManager.
type ServiceManagerGetter interface {
	Service() ServiceManager
}

// ServiceManager ...
type ServiceManager interface {
	k8s.ClientSet
	GetMasterCassIPs() (ipArr []string)
	GetMasterCassUserPwd() (user, pwd string)
	GetMasterCassPort() (port int)

	GetSubCassIPs(vsetID int) (ipArr []string)
	GetSubCassUserPwd(vsetID int) (user, pwd string)
	GetSubCassPort(vsetID int) (port int)

	GetAllNodeIPs() (ipArr []string)

	EnableNodeLabelByLabels(nodeLabel []string) error
	DisableNodeLabelByLabels(nodeLabel []string) error
	DeletePodsByLabel(podLabel string) (err error)
	// K8sDisableNodeLabel(nodeName, nodeLabel, podLabel string) error
	// K8sEnableNodeLabelByType(serviceType int) error
	// K8sDisableNodeLabelByType(serviceType int) error
	// K8sStartAll(serviceType int) error
	// K8sShutdownAll(serviceType int) error
	Test(podLabel string) error
}

// Worker ...
type Worker struct {
	wg          sync.WaitGroup
	done        chan struct{}
	maxParallel int
}

// svManager implements NodeInterface
type svManager struct {
	k8s.Client
}

// newServiceMgr returns a Nodes
func newServiceMgr(v *Vizion) *svManager {
	return &svManager{
		v.GetK8sClient(),
	}
}

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

func (s *svManager) GetAllNodeIPs() (ipArr []string) {
	nodeArr := s.GetNodeInfoArr()
	for _, node := range nodeArr {
		ipArr = append(ipArr, node["IP"])
	}
	return
}

// EnableNodeLabel .
func (s *svManager) EnableNodeLabelByLabels(nodeLabelArr []string) error {
	return nil
}

func (s *svManager) DisableNodeLabelByLabels(nodeLabelArr []string) (err error) {
	var nodeLabelNameMap map[string][]string
	for _, nodeLabel := range nodeLabelArr {
		nodeLabelNameMap[nodeLabel] = s.GetNodeNameArrByLabel(nodeLabel)
	}

	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)

	for nLable, nodeNameArr := range nodeLabelNameMap {
		for _, nodeName := range nodeNameArr {
			time.Sleep(2 * time.Second)
			select {
			case ch <- struct{}{}:
				w.wg.Add(1)
				go func() {
					nLables := strings.Split(nLable, ",")
					err = s.DisableNodeLabel(nodeName, nLables[len(nLables)-1])
					if err != nil {
						w.wg.Done()
						w.done <- struct{}{}
					}
					<-ch
					w.wg.Done()
				}()
			case <-w.done:
				break
			}
		}
	}
	w.wg.Wait()
	return
}

func (s *svManager) DeletePodsByLabel(podLabel string) (err error) {
	pods, err := s.GetPodListByLabel(podLabel)
	if err != nil {
		return
	}

	w := Worker{maxParallel: 100}
	ch := make(chan struct{}, w.maxParallel)

	for _, pod := range pods.Items {
		podName := pod.ObjectMeta.Name
		nodeName := pod.Spec.NodeName
		time.Sleep(2 * time.Second)
		select {
		case ch <- struct{}{}:
			w.wg.Add(1)
			go func() {
				logger.Info("Kubectl delete pod %s ...(%s)", podName, nodeName)
				err = s.DeletePod(podName)
				if err != nil {
					w.wg.Done()
					w.done <- struct{}{}
				}
				<-ch
				w.wg.Done()
			}()
		case <-w.done:
			break
		}
	}
	w.wg.Wait()
	return
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
