package resources

import (
	"fmt"
	"pzatest/config"
	"pzatest/libs/k8s"
	"pzatest/libs/utils"
	"pzatest/types"
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
	GetNodeNameArrByLabels(nodeLabelArr []string) (nodeNameArr []string)
	GetNodeIPArrByLabels(nodeLabelArr []string) (nodeIPArr []string)
	GetBdNodeIPArr() (nodeIPArr []string)
	GetESPodPvcVolume(esPodName, volumeName string) (pvcVol string, err error)
	GetESNodeIPPvcArrMap() (map[string][]string, error)

	EnableNodeLabels(nodeLabel []string) error
	DisableNodeLabels(nodeLabel []string) error
	DeletePodsByLabel(podLabel string) (err error)
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
	Base types.VizionBaseInput
}

// newServiceMgr returns a Nodes
func newServiceMgr(v *Vizion) *svManager {
	return &svManager{
		v.GetK8sClient(), v.Base,
	}
}

// ----------- Get Cassandra configs from k8s -----------
// GetMasterCassIPs ...
func (s *svManager) GetMasterCassIPs() (ipArr []string) {
	ipArr, err := s.GetSvcIPs("cassandra-master-expose")
	if err != nil {
		panic(err)
	}
	return
}

// GetMasterCassUserPwd ...
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

// GetMasterCassPort ...
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

// GetSubCassUserPwd ...
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

// GetSubCassPort ...
func (s *svManager) GetSubCassPort(vsetID int) (port int) {
	port, err := s.GetSvcPort(fmt.Sprintf("cassandra-vset%d-expose", vsetID), 9042)
	if err != nil {
		panic(err)
	}
	return
}

//  ----------- Get node IPs/Names -----------
func (s *svManager) GetAllNodeIPs() (ipArr []string) {
	nodeArr := s.GetNodeInfoArr()
	for _, node := range nodeArr {
		ipArr = append(ipArr, node["IP"])
	}
	return
}

func (s *svManager) GetNodeNameArrByLabels(nodeLabelArr []string) (nodeNameArr []string) {
	for _, nLable := range nodeLabelArr {
		nodeNameArr = append(nodeNameArr, s.GetNodeNameArrByLabel(nLable)...)
	}
	return
}

func (s *svManager) GetNodeIPArrByLabels(nodeLabelArr []string) (nodeIPArr []string) {
	for _, nLable := range nodeLabelArr {
		nodeIPArr = append(nodeIPArr, s.GetNodeIPArrByLabel(nLable)...)
	}
	return
}

func (s *svManager) GetBdNodeIPArr() (nodeIPArr []string) {
	nodeLabelArr, _ := config.Dpldagent.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
}

// GetESNodeIPPvcArrMap ...
func (s *svManager) GetESPodPvcVolume(esPodName, volumeName string) (pvcVol string, err error) {
	if volumeName == "" {
		volumeName = "es-data-store"
	}

	esPod, err := s.GetPodDetail(esPodName)
	if err != nil {
		return pvcVol, err
	}

	for _, vol := range esPod.Spec.Volumes {
		volPvc := vol.PersistentVolumeClaim
		if vol.Name == volumeName && volPvc != nil {
			pvcVol = volPvc.ClaimName
			break
		}
	}
	return
}

// GetESNodeIPPvcArrMap ...
func (s *svManager) GetESNodeIPPvcArrMap() (map[string][]string, error) {
	nodeIPPvcArr := map[string][]string{} // ip:pvcArr
	podLabel := config.ES.GetPodLabel(s.Base)
	esPods, err := s.GetPodListByLabel(podLabel)
	if err != nil {
		return nodeIPPvcArr, err
	}
	for _, pod := range esPods.Items {
		nodeName := pod.Spec.NodeName
		if nodeName == "" {
			continue
		}
		nodeIP := s.GetNodePriorIPByName(nodeName)
		pvcName, err := s.GetESPodPvcVolume(pod.ObjectMeta.Name, "es-data-store")
		if err != nil {
			return nodeIPPvcArr, err
		}
		nodeIPPvcArr[nodeIP] = []string{pvcName}
	}
	return nodeIPPvcArr, nil
}

// ----------- Enable/Disable Node Labels -----------
func (s *svManager) EnableNodeLabels(nodeLabelArr []string) (err error) {
	nodeLabelNameMap := map[string][]string{}
	for _, nodeLabel := range nodeLabelArr {
		nodeLabelNameMap[nodeLabel] = s.GetNodeNameArrByLabel(nodeLabel)
	}
	// logger.Info(nodeLabelNameMap)
	w := Worker{maxParallel: 10}
	ch := make(chan struct{}, w.maxParallel)

	for nLable, nodeNameArr := range nodeLabelNameMap {
		for _, nodeName := range nodeNameArr {
			// time.Sleep(2 * time.Second)
			select {
			case ch <- struct{}{}:
				w.wg.Add(1)
				go func() {
					nLables := strings.Split(nLable, ",")
					err = s.EnableNodeLabel(nodeName, nLables[len(nLables)-1])
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
			w.wg.Wait()
		}
	}
	return
}

func (s *svManager) DisableNodeLabels(nodeLabelArr []string) (err error) {
	nodeLabelNameMap := map[string][]string{}
	for _, nodeLabel := range nodeLabelArr {
		nodeLabelNameMap[nodeLabel] = s.GetNodeNameArrByLabel(nodeLabel)
	}
	// logger.Info(nodeLabelNameMap)
	w := Worker{maxParallel: 10}
	ch := make(chan struct{}, w.maxParallel)

	for nLable, nodeNameArr := range nodeLabelNameMap {
		for _, nodeName := range nodeNameArr {
			// time.Sleep(2 * time.Second)
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
			w.wg.Wait()
		}
	}
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
				logger.Infof("Kubectl delete pod %s ...(%s)", podName, nodeName)
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
