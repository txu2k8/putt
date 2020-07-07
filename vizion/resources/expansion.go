package resources

import (
	"bufio"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"putt/config"
	"putt/libs/k8s"
	"putt/libs/runner/schedule"
	"putt/libs/utils"
	"putt/types"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	// Get once
	if v.KubeConfig != "" {
		return
	}
	// Use input KubeConfig path
	if v.Base.KubeConfig != "" {
		v.KubeConfig = v.Base.KubeConfig
		return
	}
	fqdn := "kubernetes.vizion.local"
	dir, _ := os.Getwd()
	kubePath := path.Join(dir, "tmp", "kube")
	_, err := os.Stat(kubePath)
	if os.IsNotExist(err) {
		err := os.MkdirAll(kubePath, os.ModePerm)
		if err != nil {
			panic(err)
		}
	}

	// use exist .kube/config if exist with MasterIPs
	for _, masterIP := range v.Base.MasterIPs {
		tmpCfPath := path.Join(kubePath, masterIP+".config")
		_, err = os.Stat(tmpCfPath)
		if err == nil || os.IsExist(err) {
			v.KubeConfig = tmpCfPath
			return
		}
	}

	masterIP := v.VaildMasterIP()
	cfPath := path.Join(kubePath, masterIP+".config")
	_, err = os.Stat(cfPath)
	if os.IsNotExist(err) {
		n := v.Node(masterIP)
		err := n.GetKubeConfig(cfPath)
		if err != nil {
			panic(err)
		}

		localIP := utils.GetLocalIP()
		if !collection.Collect(v.Base.MasterIPs).Contains(localIP) &&
			v.Base.UserName != "centos" {
			server := n.GetKubeVipIP(fqdn)
			ReplaceKubeServer(cfPath, server)
		}
	}
	v.KubeConfig = cfPath
}

// ============ Get etcd cert files ============

// GetEtcdConfig ...
func (v *Vizion) GetEtcdConfig() (crtArr []string) {
	var err error
	// Get once
	if v.EtcdCrt != "" {
		crtArr, err = filepath.Glob(filepath.Join(v.EtcdCrt, "*"))
		if err != nil {
			panic(err)
		}
		return
	}

	dir, _ := os.Getwd()
	tmpPath := path.Join(dir, "tmp")

	// use exist .etcd/config if exist with MasterIPs
	for _, masterIP := range v.Base.MasterIPs {
		tmpCrtPath := path.Join(tmpPath, "etcd", masterIP)
		_, err = os.Stat(tmpCrtPath)
		if err == nil || os.IsExist(err) {
			v.EtcdCrt = tmpCrtPath
			crtArr, err = filepath.Glob(filepath.Join(tmpCrtPath, "*"))
			if err != nil {
				panic(err)
			}
			return
		}
	}

	masterIP := v.VaildMasterIP()
	crtPath := path.Join(tmpPath, "etcd", masterIP)
	_, err = os.Stat(crtPath)
	if os.IsNotExist(err) {
		n := v.Node(masterIP)
		crtArr, err = n.GetEtcdCerts(tmpPath)
		if err != nil {
			panic(err)
		}
	}
	v.EtcdCrt = crtPath
	return
}

// ============ Stop/Start/Apply Services ============

// StopServices .
func (v *Vizion) StopServices(svArr []config.Service) error {
	svMgr := v.Service()
	for _, sv := range svArr {
		logger.Infof(">> Stop service %s:%d ...", sv.TypeName, sv.Type)
		podLabel := sv.GetPodLabel(v.Base)
		nodeLabelArr, _ := sv.GetNodeLabelArr(v.Base)
		nodeIPPvcArr := map[string][]string{}
		if sv.Type == config.ES.Type { // Get es nodeip <-> pvcArr
			nodeIPPvcArrMap, err := svMgr.GetESNodeIPPvcArrMap()
			if err != nil {
				return err
			}
			nodeIPPvcArr = nodeIPPvcArrMap
		}

		// stop
		switch sv.K8sKind {
		case config.K8sStatefulsets:
			if sv.Type == config.ES.Type { // disable label
				svMgr.DisableNodeLabels(nodeLabelArr)
				svMgr.DeletePodsByLabel(podLabel)
				svMgr.WaitForAllPodDown(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 60)
			} else { // set replicas
				k8sNameArr, _ := svMgr.GetStatefulSetsNameArrByLabel(podLabel)
				for _, k8sName := range k8sNameArr {
					svMgr.SetStatefulSetsReplicas(k8sName, 0)
					svMgr.WaitForPodDown(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 60)
				}
			}
		case config.K8sDeployment: // set replicas
			k8sNameArr, _ := svMgr.GetDeploymentsNameArrByLabel(podLabel)
			for _, k8sName := range k8sNameArr {
				svMgr.SetDeploymentsReplicas(k8sName, 0)
				svMgr.WaitForPodDown(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 60)
			}
		default: // disable label
			svMgr.DisableNodeLabels(nodeLabelArr)
			svMgr.DeletePodsByLabel(podLabel)
			svMgr.WaitForAllPodDown(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 60)
		}

		// check after stop
		if sv.Type == config.ES.Type {
			// expected all volume(pvc) status==2
			logger.Info("> Wait BD Volume Status=2 ...")
			v.WaitBdVolumeStatusExpected(2, "", "", []string{})
			logger.Info("> Wait BlockDevice Removed ...")
			for nodeIP, pvcArr := range nodeIPPvcArr {
				v.WaitBlockDeviceRemoved("", nodeIP, pvcArr)
			}
		}
		if sv.Type == config.Dpldagent.Type {
			// rmmod dpl after bd pod stop
			logger.Info("> rmmod dpl ...")
			v.RmmodDplOnBD()
		}
	}
	return nil
}

// StartServices .
func (v *Vizion) StartServices(svArr []config.Service, cleanJdevice, cleanSC bool) error {
	var err error
	svMgr := v.Service()
	for _, sv := range svArr {
		logger.Infof(">> Start service %s:%d ...", sv.TypeName, sv.Type)
		podLabel := sv.GetPodLabel(v.Base)
		nodeLabelArr, nodeLabelKVArr := sv.GetNodeLabelArr(v.Base)

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

		// Check before check
		switch sv.Type {
		case config.Vizions3.Type, config.Dpldagent.Type:
			// Check is all anchor ok
			v.WaitForAllJnsPrimary()
		}

		// start
		switch sv.K8sKind {
		case config.K8sStatefulsets:
			if sv.Type == config.ES.Type { // Enable label
				svMgr.EnableNodeLabels(nodeLabelArr)
				svMgr.WaitForAllPodReady(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 60)
			} else { // set replicas
				if replicas == 0 {
					break
				}
				k8sNameArr, _ := svMgr.GetStatefulSetsNameArrByLabel(podLabel)
				for _, k8sName := range k8sNameArr {
					svMgr.SetStatefulSetsReplicas(k8sName, replicas)
					// svMgr.WaitForPodReady(k8s.IsPodReadyInput{PodNamePrefix: k8sName}, 60)
				}
				svMgr.WaitForAllPodReady(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 60)
			}
		case config.K8sDeployment: // set replicas
			if replicas == 0 {
				break
			}
			k8sNameArr, _ := svMgr.GetDeploymentsNameArrByLabel(podLabel)
			for _, k8sName := range k8sNameArr {
				svMgr.SetDeploymentsReplicas(k8sName, replicas)
			}
			svMgr.WaitForAllPodReady(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 60)
		default: // disable label
			svMgr.EnableNodeLabels(nodeLabelArr)
			svMgr.WaitForAllPodReady(k8s.IsAllPodReadyInput{PodLabel: podLabel}, 60)
		}

		// Check after start
		switch sv.Type {
		case config.Jddpl.Type: // Check stg_unit status and number
			if cleanJdevice == true {
				// err = v.IsStgUnitStatusOK()
				err = v.Schedule.RunPhase(v.IsStgUnitStatusOK, schedule.Desc("Check etcd if all stg_unit status=0"))
				if err != nil {
					return err
				}
			}
			// err = v.IsJnlFormatSuccess()
			err = v.Schedule.RunPhase(v.IsJnlFormatSuccess, schedule.Desc("Check if (jdevice_GB/2 -1)=len(etcd stg_unit)"))
			if err != nil {
				return err
			}
		case config.Servicedpl.Type: // Try clean storage_cache in pod  -- aws
			if cleanSC == true {
				v.CleanStorageCache(config.CleanSC.Arg[0], true)
			}
		case config.ES.Type: // wait for volume status=1
			v.WaitBdVolumeStatusExpected(1, "", "", []string{})
		}

	}

	return nil
}

// ApplyServicesImage .
func (v *Vizion) ApplyServicesImage(svArr []config.Service, image string) error {
	svMgr := v.Service()
	for _, sv := range svArr {
		if !collection.Collect(config.DefaultDplServiceArray).Contains(sv) {
			continue
		}
		svContainer := sv.Container
		logger.Infof(">> Apply service image %s %s(%s):%s ...", sv.K8sKind, sv.TypeName, svContainer, image)
		podLabel := sv.GetPodLabel(v.Base)

		switch sv.K8sKind {
		case config.K8sStatefulsets: // Statefulsets
			k8sNameArr, _ := svMgr.GetStatefulSetsNameArrByLabel(podLabel)
			for _, k8sName := range k8sNameArr {
				svMgr.SetStatefulSetsImage(k8sName, svContainer, image)
			}
		case config.K8sDeployment: // Deployment
			k8sNameArr, _ := svMgr.GetDeploymentsNameArrByLabel(podLabel)
			for _, k8sName := range k8sNameArr {
				svMgr.SetDeploymentsImage(k8sName, svContainer, image)
			}
		case config.K8sDaemonsets: // Daemonsets
			k8sNameArr, _ := svMgr.GetDaemonsetsNameArrByLabel(podLabel)
			for _, k8sName := range k8sNameArr {
				svMgr.SetDaemonSetsImage(k8sName, svContainer, image)
			}
		default: // not support
			logger.Errorf("Not supported k8s resource: %s", sv.K8sKind)
		}
	}
	return nil
}

// ApplyDplmanagerShellImage .
func (v *Vizion) ApplyDplmanagerShellImage(image string) error {
	dplmgrPath := config.DplmanagerLocalPath
	lsCmd := fmt.Sprintf("ls %s", dplmgrPath)
	svMgr := v.Service()
	nodeIPs := svMgr.GetAllNodeIPs()

	for _, nodeIP := range nodeIPs {
		node := v.Node(nodeIP)
		_, output := node.RunCmd(lsCmd)
		if strings.Contains(output, "No such file or directory") {
			continue
		}
		err := node.ChangeDplmanagerShellImage(image, dplmgrPath)
		if err != nil {
			return err
		}
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
	logger.Infof("Log path base name list: %s", utils.Prettify(logPathArr))
	for _, nodeIP := range v.Service().GetAllNodeIPs() {
		node := v.Node(nodeIP)
		node.CleanLog(logPathArr)
	}

	return nil
}

// CleanEtcd .
func (v *Vizion) CleanEtcd(prefixArr []string) error {
	logger.Info("> Clean etcd data ...")
	// etcdctlv3 del --prefix /vizion/dpl/add_vol
	cmdArr := []string{}
	for _, prefix := range prefixArr {
		cmdArr = append(cmdArr, "etcdctlv3 del --prefix "+prefix)
		cmdArr = append(cmdArr, "etcdctlv3 get --prefix "+prefix)
	}
	masterNode := v.MasterNode()
	for _, cmd := range cmdArr {
		rc, output := masterNode.RunCmd(cmd)
		logger.Infof("%d, %s", rc, output)
	}
	return nil
}

// FormatJdevice Run dd command on /dev/j_device
func (v *Vizion) FormatJdevice(nodeIP, jdev, jdPodName string) error {
	formatCmd := fmt.Sprintf("dd if=/dev/zero of=%s bs=1k count=4", jdev)
	if jdPodName != "" { // Run format cmd in pod
		// TODO
	} else { // run format cmd on node local
		n := v.Node(nodeIP)
		_, output := n.RunCmd(formatCmd)
		logger.Info(output)
	}
	return nil
}

// CleanJdevice . Format Journal device on vsphere/aws env
func (v *Vizion) CleanJdevice() error {
	logger.Info("> Clean journal device ...")
	_, nodeLabelArr := config.Jddpl.GetNodeLabelArr(v.Base)
	// podLabel := config.Jddpl.GetPodLabel(v.Base)
	k8sSv := v.Service()
	jddplNodeIPs := k8sSv.GetNodeIPArrByLabels(nodeLabelArr)
	if len(jddplNodeIPs) <= 1 {
		return fmt.Errorf("Find jddpl Nodes <= 1")
	}

	jdeviceLsCmd := fmt.Sprintf("ls -lh %s*", config.JDevicePath)
	jdevicePattern := regexp.MustCompile(`/dev/j_device\d*`)
	awsEnv := false
	jdevArr := []string{}
	for _, nodeIP := range jddplNodeIPs {
		n := v.Node(nodeIP)
		_, output := n.RunCmd(jdeviceLsCmd)
		logger.Info(output)
		if strings.Contains(output, "No such file or directory") {
			// aws env, just support format_jdevice on jd_pod
			awsEnv = true
		} else { // vmware env, support format_jdevice on local
			logger.Info("Local disk, Format journal on local ...")
			jdPodName := ""
			matched := jdevicePattern.FindAllStringSubmatch(output, -1)
			for _, match := range matched {
				jdevArr = append(jdevArr, match[0])
			}
			logger.Info(jdevArr)
			for _, jdev := range jdevArr {
				err := v.FormatJdevice(nodeIP, jdev, jdPodName)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	if awsEnv == true {
		k8sSv = v.Service()
		// if servicedpl already started, skip format_jdevice
		dplPodLabel := config.Servicedpl.GetPodLabel(v.Base)
		dplPods, _ := k8sSv.GetPodListByLabel(dplPodLabel)
		if len(dplPods.Items) > 0 {
			logger.Warning("Servicedpl already started, skip format_jdevice ..")
			return nil
		}
		// aws env, support format_jdevice on jddpl pods
		v.StartServices([]config.Service{config.Jddpl}, false, false)
		jdPodLabel := config.Jddpl.GetPodLabel(v.Base)
		jdPods, _ := k8sSv.GetPodListByLabel(jdPodLabel)
		if len(jdPods.Items) == 0 {
			panic("None jddpl pods found")
		}

		for _, pod := range jdPods.Items {
			jdPodName := pod.ObjectMeta.Name
			nodeIP := pod.Status.HostIP
			logger.Infof("Format journal on pod %s..", jdPodName)
			containerName := config.Jddpl.Container
			output, _ := k8sSv.Exec(k8s.ExecInput{
				PodName:       jdPodName,
				ContainerName: containerName,
				Command:       jdeviceLsCmd,
			})
			matched := jdevicePattern.FindAllStringSubmatch(output.Stdout, -1)
			for _, match := range matched {
				jdevArr = append(jdevArr, match[0])
			}
			logger.Info(jdevArr)
			for _, jdev := range jdevArr {
				err := v.FormatJdevice(nodeIP, jdev, jdPodName)
				if err != nil {
					panic(err)
				}
			}
		}
		v.StopServices([]config.Service{config.Jddpl})
	}
	return nil
}

// GetJdeviceStgUnitNumber .
func (v *Vizion) GetJdeviceStgUnitNumber(jdPodName, nodeIP, jdevice string) (stgUnitNum int64) {
	if jdPodName != "" { // Get from pod j_device
		cmdSpec := fmt.Sprintf("fdisk -l %s | grep Disk | awk -F ',' '{print $2}' | awk '{print $1}'", jdevice)
		execInput := k8s.ExecInput{
			PodName:       jdPodName,
			ContainerName: config.Jddpl.Container,
			Command:       cmdSpec,
		}
		output, err := v.Service().Exec(execInput)
		if err != nil {
			panic(err)
		}
		sBytes, _ := strconv.ParseInt(strings.TrimSpace(strings.TrimRight(output.Stdout, "\n")), 10, 64)
		stgUnitNum = sBytes/1024/1024/1024/2 - 1
	} else if nodeIP != "" {
		n := v.Node(nodeIP)
		stgUnitNum = n.GetJdeviceStgUnitNumber(jdevice)
	} else {
		panic("Need args jdPodName or nodeIP")
	}
	return
}

// IsStgUnitStatusOK Check if etcd all stg_unit status=0
func (v *Vizion) IsStgUnitStatusOK() error {
	logger.Info("> Check etcd if all stg_unit status=0 ...")
	etcd := v.Etcd()
	stgUnitArr, _ := etcd.GetStgUnitArr()
	stgUnitNum := 0
	for _, stgUnit := range stgUnitArr {
		if stgUnit.Status != 0 {
			logger.Error(utils.Prettify(stgUnit))
			return fmt.Errorf("stg_unit status!=0")
		}
		stgUnitNum++
	}
	logger.Infof("All %d stg_unit status=0", stgUnitNum)
	return nil
}

// IsJnlFormatSuccess Check if etcd len(stg_unit) == (jdevice_GB/2 -1)
func (v *Vizion) IsJnlFormatSuccess() error {
	logger.Info("> Check if (jdevice_GB/2 -1)=len(etcd stg_unit) ...")
	jdeviceLsCmd := fmt.Sprintf("ls -lh %s*", config.JDevicePath)
	jdevicePattern := regexp.MustCompile(`/dev/j_device\d*`)
	jdPodLabel := config.Jddpl.GetPodLabel(v.Base)
	k8sSv := v.Service()
	jdPods, _ := k8sSv.GetPodListByLabel(jdPodLabel)
	if len(jdPods.Items) == 0 {
		panic("None jddpl pods found")
	}
	var expectStgUnit int64
	for _, pod := range jdPods.Items {
		jdevArr := []string{}
		jdPodName := pod.ObjectMeta.Name
		// nodeIP := pod.Status.HostIP
		logger.Infof("Get j_device size on pod %s..", jdPodName)
		containerName := config.Jddpl.Container
		output, _ := k8sSv.Exec(k8s.ExecInput{
			PodName:       jdPodName,
			ContainerName: containerName,
			Command:       jdeviceLsCmd,
		})
		matched := jdevicePattern.FindAllStringSubmatch(output.Stdout, -1)
		for _, match := range matched {
			jdevArr = append(jdevArr, match[0])
		}
		logger.Info(jdevArr)
		for _, jdev := range jdevArr {
			jdevStgUnit := v.GetJdeviceStgUnitNumber(jdPodName, "", jdev)
			logger.Infof("%s:%s stg_unit=%d", jdPodName, jdev, jdevStgUnit)
			expectStgUnit += jdevStgUnit
		}
	}
	etcdStgUnit, _ := v.Etcd().GetStgUnitNumber()
	logger.Infof("Total expect_stg_unit/etcd_stg_unit: %d/%d", expectStgUnit, etcdStgUnit)
	if etcdStgUnit != expectStgUnit {
		return fmt.Errorf("etcdStgUnit != jdeviceStgUnit, IsJnlFormatSuccess Fail")
	}
	return nil
}

// CleanStorageCache .
func (v *Vizion) CleanStorageCache(scPath string, podBash bool) error {
	logger.Infof("> Delete %s* on servicedpl nodes ...", scPath)
	_, nodeLabelArr := config.Servicedpl.GetNodeLabelArr(v.Base)
	k8sSv := v.Service()
	servicedplNodeIPs := k8sSv.GetNodeIPArrByLabels(nodeLabelArr)
	if len(servicedplNodeIPs) <= 1 {
		return fmt.Errorf("Find servicedpl Nodes <= 1")
	}

	scLsCmd := fmt.Sprintf("ls -lh %s", scPath)
	for _, nodeIP := range servicedplNodeIPs {
		n := v.Node(nodeIP)
		_, output := n.RunCmd(scLsCmd)
		logger.Info(output)
		if strings.Contains(output, "No such file or directory") {
			// No Storage Cache in local, need delete in pods
			if podBash == true {
				podName := ""
				podLabel := config.Servicedpl.GetPodLabel(v.Base)
				pods, _ := k8sSv.GetPodListByLabel(podLabel)
				for _, pod := range pods.Items {
					podHostIP := pod.Status.HostIP
					if podHostIP == nodeIP {
						podName = pod.ObjectMeta.Name
						logger.Infof("> Delete storage cache files on pod %s ...", podName)
						break
					}
				}
				if podName == "" {
					return fmt.Errorf("None servicedpl pods on node %s", nodeIP)
				}
				containerName := config.Servicedpl.Container
				err := k8sSv.DeleteFilesInPod(scPath, podName, containerName)
				if err != nil {
					return err
				}

			} else {
				logger.Warningf("No Storage Cache on local, Skip delete %s on local ...", scPath)
				continue
			}
		} else {
			if podBash == true {
				logger.Warningf("Already deleted %s on local, Skip delete %s in pod ...", scPath)
			} else {
				err := n.DeleteFiles(scPath)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CleanSubCassTables .
func (v *Vizion) CleanSubCassTables(tableNameArr []string) error {
	logger.Info("> Clean sub-cassandra tables ...")
	cass := v.Cass()
	for _, vsetID := range v.Base.VsetIDs {
		subCass := cass.SetIndex(strconv.Itoa(vsetID))
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
	logger.Info("> Update Master Cassandra Tables ...")
	var err error
	masterCass := v.Cass().SetIndex("0")
	logger.Info("> Updata VPM ...")
	logger.Info("> Updata DPL ...")
	logger.Info("> Updata ANCHSERVER ...")
	logger.Info("> Clean JFS table ...")

	logger.Info("> Insert index_map table ...")
	insertIdxMapCmdArr := []string{
		"insert into vizion.index_map (id, idx) VALUES (00000000-0000-0000-0000-111111111111, 432345564228567616)",
		"insert into vizion.index_map (id, idx) VALUES (00000000-0000-0000-0000-222222222222, 1000000)",
		"insert into vizion.index_map (id, idx) VALUES (00000000-0000-0000-0000-333333333333, 144115188076855872)",
		"insert into vizion.index_map (id, idx) VALUES (00000000-0000-0000-0000-444444444444, 288230376152711744)",
	}
	for _, cmd := range insertIdxMapCmdArr {
		err = masterCass.Execute(cmd)
		if err != nil {
			return err
		}
	}
	return nil
}

// CleanCdcSubCassCdc .
func (v *Vizion) CleanCdcSubCassCdc(vsetIDs []int) error {
	cmdArr := []string{
		"nodetool drain",
		"rm -rf /var/lib/cassandra/cdc_cache/*",
		"rm -rf /var/lib/cassandra/data/cdc_cache/*",
		"rm -rf /var/lib/cassandra/data/cdc_raw/*",
	}
	subCassContainer := config.SubCass.Container
	subCassPodLabel := config.SubCass.GetPodLabel(v.Base)

	k8sSv := v.Service()
	subCassPods, err := k8sSv.GetPodListByLabel(subCassPodLabel)
	if err != nil {
		return err
	}

	for _, pod := range subCassPods.Items {
		for _, cmdSpec := range cmdArr {
			execInput := k8s.ExecInput{
				PodName:       pod.Name,
				ContainerName: subCassContainer,
				Command:       cmdSpec,
			}
			output, err := k8sSv.Exec(execInput)
			logger.Info(utils.Prettify(output))
			if err != nil {
				return err
			}
		}
		k8sSv.DeletePod(pod.Name)
		utils.SleepProgressBar(20 * time.Second)
		isReadyInput := k8s.IsPodReadyInput{
			PodName: pod.Name,
		}
		k8sSv.WaitForPodReady(isReadyInput, 60)
	}
	return nil
}

// CleanCdcCassMonitor . TODO
func (v *Vizion) CleanCdcCassMonitor() error {
	var err error
	cdcPath := "/var/cassandra/monitor/cdc/updated_volume"
	logger.Infof("Clean cdc data on Cassandra Monitor, path:%s", cdcPath)
	cassMonitorPodLabel := config.CassMonitor.GetPodLabel(v.Base)
	cassMonitorContainer := config.CassMonitor.Container
	k8sSv := v.Service()
	cassMonitorPods, err := k8sSv.GetPodListByLabel(cassMonitorPodLabel)
	if err != nil {
		return err
	}
	if len(cassMonitorPods.Items) == 0 {
		return fmt.Errorf("None of cassandra-monitor pods found")
	}
	// clean cdcgc by cdcgcPath
	cmdRm := "rm -rf " + cdcPath
	cmdLs := "ls -lh " + path.Dir(cdcPath)
	for _, pod := range cassMonitorPods.Items {
		rmInput := k8s.ExecInput{
			PodName:       pod.Name,
			ContainerName: cassMonitorContainer,
			Command:       cmdRm,
		}
		output, err := k8sSv.Exec(rmInput)
		logger.Info(utils.Prettify(output))
		if err != nil {
			return err
		}

		lsInput := k8s.ExecInput{
			PodName:       pod.Name,
			ContainerName: cassMonitorContainer,
			Command:       cmdLs,
		}
		output, _ = k8sSv.Exec(lsInput)
		logger.Info(utils.Prettify(output))

		k8sSv.DeletePod(pod.Name)
		// Wait For Pod Ready  -- SKIP
	}
	return nil
}

// CleanCdcgc .
/*If bd-cdcgc-xxx / s3-cdcgc-xxx resource exist:
1. Clean CDC data in cassandra-vset-xxx pod:
	nodetool drain
	rm -rf /var/lib/cassandra/cdc_cache/*
	rm -rf /var/lib/cassandra/data/cdc_cache/*
	rm -rf /var/lib/cassandra/data/cdc_raw/*
2. restart cassandra-vset-xxx pod
3. Clean CDC data in cassandra-monitor-xxx pod:
	rm -rf /var/cassandra/monitor/cdc/updated_volume
	find /var/cassandra/monitor/cdc/updated_volume/result/ -type f -exec rm -rf {} \\; -print
4. restart cassandra-monitor-xxx pod
*/
func (v *Vizion) CleanCdcgc() error {
	logger.Info("> Clean cdcgc data on gc/cassandra-monitor pods ...")
	var err error
	var base types.BaseInput
	k8sSv := v.Service()
	cdcgcVsetIDs := []int{}
	for _, vsetID := range v.Base.VsetIDs {
		utils.DeepCopy(v.Base, base)
		base.VsetIDs = []int{vsetID}
		cdcgcBdPodLabel := config.Cdcgcbd.GetPodLabel(base)
		cdcgcS3PodLabel := config.Cdcgcs3.GetPodLabel(base)
		cdcgcBdK8sArr, err := k8sSv.GetDeploymentsNameArrByLabel(cdcgcBdPodLabel)
		if err != nil {
			return err
		}
		cdcgcS3K8sArr, err := k8sSv.GetDeploymentsNameArrByLabel(cdcgcS3PodLabel)
		if err != nil {
			return err
		}
		cdcgcK8sArr := append(cdcgcBdK8sArr, cdcgcS3K8sArr...)
		if len(cdcgcK8sArr) > 0 {
			cdcgcVsetIDs = append(cdcgcVsetIDs, vsetID)
		}
	}

	if len(cdcgcVsetIDs) == 0 {
		logger.Warningf("None cdcgc config on vsets%v, skipped!!", v.Base.VsetIDs)
		return nil
	}

	err = v.CleanCdcSubCassCdc(cdcgcVsetIDs)
	if err != nil {
		return err
	}
	err = v.CleanCdcCassMonitor()
	if err != nil {
		return err
	}
	return nil
}

// RmmodDplOnBD "rmmod dpl" on all bd nodes
func (v *Vizion) RmmodDplOnBD() error {
	bdNodeIps := v.Service().BdNodeIPArr()
	logger.Info(bdNodeIps)
	for _, nodeIP := range bdNodeIps {
		node := v.Node(nodeIP)
		if err := node.RmModDpl(); err != nil {
			return err
		}
	}
	return nil
}

// SetBdVolumeKV . TODO Ctime
func (v *Vizion) SetBdVolumeKV(kvArr []string) error {
	var err error
	cass := v.Cass()
	bdServiceArr, err := cass.SetIndex("0").GetServiceByType(config.Dpldagent.Type)
	if err != nil {
		return err
	}
	bdIDs := []string{}
	for _, bdSv := range bdServiceArr {
		bdIDs = append(bdIDs, bdSv.ID)
	}
	for _, vsetID := range v.Base.VsetIDs {
		subCass := cass.SetIndex(string(vsetID))
		volumeArr, err := subCass.GetVolume()
		if err != nil {
			return err
		}
		for _, vol := range volumeArr {
			if vol.Status == 0 || (vol.BlockDeviceService != "" && collection.Collect(bdIDs).Contains(vol.BlockDeviceService)) {
				continue
			} else {
				ctime := vol.Ctime // TODO
				for _, kv := range kvArr {
					logger.Infof("> Set vizion.volume: %s ...", kv)
					cmdSpec := fmt.Sprintf("UPDATE vizion.volume SET %s WHERE type=0 AND name=%s AND c_time=%s", kv, vol.Name, ctime)
					err = subCass.Execute(cmdSpec)
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// FormatBdVolume Format SubCass Volume Table
func (v *Vizion) FormatBdVolume() error {
	var err error
	// vizion.volume
	kvArr := []string{
		"format=False",
		"status=2",
		"block_device_service=null",
	}
	err = v.SetBdVolumeKV(kvArr)
	if err != nil {
		return err
	}
	return nil
}

// ============ GitLab / Git / Image ============
// TODO

// ============ Expansion ============

// GetS3ChannelDplUUIDs .
func (v *Vizion) GetS3ChannelDplUUIDs() []string {
	return nil
}
