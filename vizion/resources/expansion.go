package resources

import (
	"bufio"
	"os"
	"path"
	"pzatest/config"
	"pzatest/libs/k8s"
	"pzatest/libs/utils"
	"strings"

	"github.com/chenhg5/collection"
)

// ============ Get /root/.kube/config ============

// ReplaceKubeServer .
func ReplaceKubeServer(cfPath, server string) {
	defaultServer := "kubernetes.vizion.local"
	logger.Infof("Replace kube-config server: %s -> %s", defaultServer, server)
	file, err := os.OpenFile(cfPath, os.O_RDWR, 0666)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	fc := bufio.NewScanner(file)
	var content string
	for fc.Scan() {
		lineString := fc.Text()
		if strings.Contains(lineString, defaultServer) {
			lineString = strings.Replace(lineString, defaultServer, server, 1)
		}
		content += lineString + "\n"
	}

	err = file.Truncate(0)
	if nil != err {
		panic(err)
	}

	file.Seek(0, 0)
	_, err = file.WriteString(content)
	if nil != err {
		panic(err)
	}
}

// GetKubeConfig ...
func (v *Vizion) GetKubeConfig() {
	fqdn := "kubernetes.vizion.local"
	kubePath := "/tmp/kube"
	cfPath := path.Join(kubePath, v.Base.MasterIPs[0]+".config")

	_, err := os.Stat(kubePath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(kubePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	_, err = os.Stat(cfPath)
	if os.IsNotExist(err) {
		n := v.Node(v.Base.MasterIPs[0])
		err := n.GetKubeConfig(cfPath)
		if err != nil {
			panic(err)
		}

		localIP := utils.GetLocalIP()
		if !collection.Collect(v.Base.MasterIPs).Contains(localIP) {
			server := n.GetKubeVipIP(fqdn)
			ReplaceKubeServer(cfPath, server)
		}
	}
	v.Base.KubeConfig = cfPath
}

// ============ Stop/Start Services ============

// StopService .
func (v *Vizion) StopService(sv config.Service) error {
	logger.Infof(">> Stop service %s:%d ...", sv.TypeName, sv.Type)
	podLabel := sv.GetPodLabel(v.Base)
	nodeLabelArr, _ := sv.GetNodeLabelArr(v.Base)
	svMgr := v.Service()

	switch sv.K8sKind {
	case config.K8sStatefulsets:
		if sv.Type == config.ES.Type { // disable label
			svMgr.DisableNodeLabelByLabels(nodeLabelArr)
			svMgr.DeletePodsByLabel(podLabel)
			svMgr.WaitForAllPodDown(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 30)
		}
		// set replicas
		k8sNameArr, _ := svMgr.GetStatefulSetsNameArrByLabel(podLabel)
		for _, k8sName := range k8sNameArr {
			svMgr.SetStatefulSetsReplicas(k8sName, 0)
			svMgr.WaitForPodDown(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 30)
		}
	case config.K8sDeployment: // set replicas
		k8sNameArr, _ := svMgr.GetDeploymentsNameArrByLabel(podLabel)
		for _, k8sName := range k8sNameArr {
			svMgr.SetDeploymentsReplicas(k8sName, 0)
			svMgr.WaitForPodDown(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 30)
		}
	default: // disable label
		svMgr.DisableNodeLabelByLabels(nodeLabelArr)
		svMgr.DeletePodsByLabel(podLabel)
		svMgr.WaitForAllPodDown(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 30)
	}
	return nil
}

// StartService .
func (v *Vizion) StartService(sv config.Service) error {
	logger.Infof(">> Start service %s:%d ...", sv.TypeName, sv.Type)
	podLabel := sv.GetPodLabel(v.Base)
	nodeLabelArr, nodeLabelKVArr := sv.GetNodeLabelArr(v.Base)

	svMgr := v.Service()

	var replicas int
	switch sv.Type {
	case config.Dplmanager.Type, config.Dplexporter.Type, config.Cdcgcbd.Type, config.Cdcgcs3.Type:
		replicas = sv.Replicas
	default:
		var nodeNameArr []string
		for _, nLabelKv := range nodeLabelKVArr {
			nodeNameArr = append(nodeNameArr, svMgr.GetNodeNameArrByLabel(nLabelKv)...)
		}
		validReplicas := len(nodeNameArr)
		replicas = utils.MaxInt(sv.Replicas, validReplicas)
	}

	switch sv.K8sKind {
	case config.K8sStatefulsets:
		if sv.Type == config.ES.Type { // disable label
			svMgr.EnableNodeLabelByLabels(nodeLabelArr)
			svMgr.WaitForAllPodReady(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 30)
		}
		// set replicas
		if replicas == 0 {
			break
		}
		k8sNameArr, _ := svMgr.GetStatefulSetsNameArrByLabel(podLabel)
		for _, k8sName := range k8sNameArr {
			svMgr.SetStatefulSetsReplicas(k8sName, replicas)
			svMgr.WaitForPodReady(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 30)
		}
	case config.K8sDeployment: // set replicas
		if replicas == 0 {
			break
		}
		k8sNameArr, _ := svMgr.GetDeploymentsNameArrByLabel(podLabel)
		for _, k8sName := range k8sNameArr {
			svMgr.SetDeploymentsReplicas(k8sName, replicas)
			svMgr.WaitForPodReady(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 30)
		}
	default: // disable label
		svMgr.EnableNodeLabelByLabels(nodeLabelArr)
		svMgr.WaitForAllPodReady(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 30)
	}
	return nil
}

// ============ Clean up ============

// CleanLog .
func (v *Vizion) CleanLog(svArr []config.Service) error {
	logger.Info("> Clean Services Logs ...")
	logPathArr := []string{}
	for _, sv := range svArr {
		logArr := sv.GetLogDirArr(v.Base)
		// logger.Info(utils.Prettify(logArr))
		logPathArr = append(logPathArr, logArr...)
	}
	for _, nodeIP := range v.Service().GetAllNodeIPs() {
		node := v.Node(nodeIP)
		node.CleanLog(logPathArr)
	}

	return nil
}

// CleanJournal .
func (v *Vizion) CleanJournal() error {
	return nil
}

// CleanStorageCache .
func (v *Vizion) CleanStorageCache() error {
	return nil
}

// CleanSubCassTables .
func (v *Vizion) CleanSubCassTables(tableNameArr []string) error {
	for _, vsetID := range v.Base.VsetIDs {
		subCass := v.Cass().SetIndex(string(vsetID))
		for _, tableName := range tableNameArr {
			err := subCass.TruncateTable(tableName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// UpdateMasterCassTables .
func (v *Vizion) UpdateMasterCassTables() error {
	// masterCass := v.Cass().SetIndex("0")
	logger.Info("> Updata VPM ...")
	logger.Info("> Updata DPL ...")
	logger.Info("> Updata ANCHSERVER ...")
	logger.Info("> Clean JFS table ...")
	logger.Info("> Insert index_map table ...")
	return nil
}

// SetBdVolumeKV .
func (v *Vizion) SetBdVolumeKV(kvArr []string) error {
	for _, vsetID := range v.Base.VsetIDs {
		subCass := v.Cass().SetIndex(string(vsetID))
		subCass.TruncateTable("")
		for _, kv := range kvArr {
			logger.Infof("> Set vizion.volume: %s ...", kv)
		}
	}
	return nil
}

// UpdateSubCassTables .
func (v *Vizion) UpdateSubCassTables() error {
	kvArr := []string{
		"format=False",
		"status=2",
		"block_device_service=null",
	}
	v.SetBdVolumeKV(kvArr)
	return nil
}

// CleanEtcd .
func (v *Vizion) CleanEtcd(prefixArr []string) error {
	return nil
}

// CleanCdcgc .
func (v *Vizion) CleanCdcgc() error {
	return nil
}
