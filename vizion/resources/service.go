package resources

import (
	"fmt"
	"putt/config"
	"putt/libs/k8s"
	"putt/libs/utils"
	"putt/types"
	"strings"
	"sync"
	"time"
)

// ServiceManagerGetter has a method to return a k8s ServiceManager.
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
	GetESPodPvcVolume(esPodName, volumeName string) (pvcVol string, err error)
	GetESNodeIPPvcArrMap() (map[string][]string, error)
	ClusterNodeIPArr() []string
	MasterNodeIPArr() []string
	VsetNodeIPArr(vsetID int) []string
	EtcdNodeIPArr() []string
	ServicedplNodeIPArr() []string
	JddplNodeIPArr() []string
	BdNodeIPArr() []string

	EnableNodeLabels(nodeLabel []string) error
	DisableNodeLabels(nodeLabel []string) error
	DeletePodsByLabel(podLabel string) (err error)
	DeleteFilesInPod(fPath, podName, containerName string) error
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
			user = strV // convert.Base64Encode(v)
		case "CASPwd":
			pwd = strV // convert.Base64Encode(v)
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
			user = strV // convert.Base64Encode(v)
		case "CASPwd":
			pwd = strV // convert.Base64Encode(v)
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

func (s *svManager) ClusterNodeIPArr() []string {
	return s.GetNodeIPArrByLabels([]string{"node-role.kubernetes.io/node"})
}

func (s *svManager) MasterNodeIPArr() []string {
	return s.GetNodeIPArrByLabels([]string{"node-role.kubernetes.io/master"})
}

func (s *svManager) VsetNodeIPArr(vsetID int) []string {
	return s.GetNodeIPArrByLabels([]string{fmt.Sprintf("node-role.kubernetes.io/vset%d", vsetID)})
}

func (s *svManager) EtcdNodeIPArr() []string {
	nodeLabelArr, _ := config.ETCD.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
}

func (s *svManager) ServicedplNodeIPArr() []string {
	nodeLabelArr, _ := config.Servicedpl.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
}

func (s *svManager) JddplNodeIPArr() []string {
	nodeLabelArr, _ := config.Jddpl.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
}

func (s *svManager) DcmapdplNodeIPArr() []string {
	nodeLabelArr, _ := config.Dcmapdpl.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
}

func (s *svManager) McmapdplNodeIPArr() []string {
	nodeLabelArr, _ := config.Mcmapdpl.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
}

func (s *svManager) CmapdplNodeIPArr() []string {
	nodeIPArr := []string{}
	nodeIPArr = append(s.DcmapdplNodeIPArr(), s.CmapdplNodeIPArr()...)
	return nodeIPArr
}

func (s *svManager) BdNodeIPArr() (nodeIPArr []string) {
	nodeLabelArr, _ := config.Dpldagent.GetNodeLabelArr(s.Base)
	return s.GetNodeIPArrByLabels(nodeLabelArr)
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
				logger.Infof("Delete pod %s on %s ...", podName, nodeName)
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

func (s *svManager) DeleteFilesInPod(fPath, podName, containerName string) error {
	rmCmd := fmt.Sprintf("find %s* | grep -v lost+found | xargs rm -rf", fPath)
	lsCmd := fmt.Sprintf("ls -l %s", fPath)

	rmInput := k8s.ExecInput{
		PodName:       podName,
		ContainerName: containerName,
		Command:       rmCmd,
	}
	output, err := s.Exec(rmInput)
	logger.Infof("%v", output)
	if err != nil {
		return err
	}

	lsInput := k8s.ExecInput{
		PodName:       podName,
		ContainerName: containerName,
		Command:       lsCmd,
	}
	output, _ = s.Exec(lsInput)
	logger.Infof("%v", output)
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
