package resources

import (
	"fmt"
	"putt/config"
	"putt/libs/k8s"
	"putt/libs/retry"
	"putt/libs/retry/strategy"
	"putt/libs/runner/schedule"
	"putt/libs/utils"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/chenhg5/collection"
)

// ============ Check health: Default ============

// Check .
type Check func() error

// RetryCheck .
func (v *Vizion) RetryCheck(check Check) error {
	action := func(attempt uint) error {
		return check()
	}
	err := retry.Retry(
		action,
		strategy.Limit(uint(180)),
		strategy.Wait(30*time.Second),
		// strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// CheckHealth .
func (v *Vizion) CheckHealth() error {
	var err error
	err = v.Schedule.RunPhase(v.WaitForPingOK, schedule.Desc("Check if nodes ping OK"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.IsNodeCrashed, schedule.Desc("Check nodes /var/crash/ files"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.IsServiceCoreDump, schedule.Desc("Check if service has core dump files"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.WaitForAllPodStatusOK, schedule.Desc("Check if All Service pods contaniers ready"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.WaitForEtcdOK, schedule.Desc("Check if etcd members stared >=3"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.WaitForCassOK, schedule.Desc("Check if cassandra 'nodetool status' all UN"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.WaitForMysqlOK, schedule.Desc("Check if Mysql members all ONLINE && PRIMARY >=1"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.WaitForDplHeloOK, schedule.Desc("Check if 'dplmanager -mdpl helo' OK"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.WaitForAllJnsPrimary, schedule.Desc("Check if all 'dplmanager -mjns stat' PRIMARY"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.PrintDplChannels, schedule.Desc("Print Dpl Channel List"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.PrintDplBalance, schedule.Desc("Print Dpl Balance list"))
	if err != nil {
		return err
	}

	err = v.Schedule.RunPhase(v.IsZpoolStatusOK, schedule.Desc("Check if zpool status ONLINE"))
	if err != nil {
		return err
	}

	return nil
}

// WaitForPingOK .
func (v *Vizion) WaitForPingOK() error {
	logger.Info("Enter WaitForPingOK ...")
	var err error
	for _, nodeIP := range v.Service().GetAllNodeIPs() {
		err = utils.IsPingOK(nodeIP)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsNodeCrashed ...
func (v *Vizion) IsNodeCrashed() error {
	logger.Info("Enter IsNodeCrashed ...")
	crashed := false
	crashArrMap := map[string][]string{}
	for _, nodeIP := range v.Service().GetAllNodeIPs() {
		node := v.Node(nodeIP)
		crashArr := node.GetCrashDirs()
		if len(crashArr) > 0 {
			crashed = true
			crashArrMap[nodeIP] = crashArr
			logger.Errorf("%s has crash files\n%s", nodeIP, utils.Prettify(crashArr))
		}
	}
	if crashed == true {
		return fmt.Errorf("Some node has crash files\n%s", utils.Prettify(crashArrMap))
	}
	return nil
}

// IsServiceCoreDump .
func (v *Vizion) IsServiceCoreDump() error {
	logger.Info("Enter IsServiceCoreDump ...")
	coreMaps := map[string][]string{}
	ignoreCoreArr := []string{
		".dplmanager.",
	}

	logPathArr := []string{}
	for _, sv := range config.DefaultServiceArray {
		logArr := sv.GetLogDirArr(v.Base)
		// logger.Info(utils.Prettify(logArr))
		logPathArr = append(logPathArr, logArr...)
	}
	logger.Infof("Log path base name list: %s", utils.Prettify(logPathArr))

	for _, nodeIP := range v.Service().GetAllNodeIPs() {
		node := v.Node(nodeIP)
		allCoreArr := node.GetCoreFiles(logPathArr)
		coreArr := []string{}
		for _, ignoreCore := range ignoreCoreArr {
			for _, corePath := range allCoreArr {
				if !strings.Contains(corePath, ignoreCore) {
					coreArr = append(coreArr, corePath)
				}
			}
		}
		if len(coreArr) > 0 {
			coreMaps[nodeIP] = coreArr
			logger.Errorf("%s has core files\n%s", nodeIP, utils.Prettify(coreMaps))
		}
	}

	if len(coreMaps) > 0 {
		return fmt.Errorf("Core files\n%s", utils.Prettify(coreMaps))
	}
	return nil
}

// WaitForAllPodStatusOK .
func (v *Vizion) WaitForAllPodStatusOK() error {
	k8sSv := v.Service()
	for _, sv := range config.DefaultServiceArray {
		if collection.Collect([]int{config.Dcmapdpl.Type, config.Mcmapdpl.Type}).Contains(sv.Type) {
			continue // skip check cmap pods
		}
		tries := 60
		mysqlTypeArr := []int{
			config.MysqlCluster.Type,
			config.MysqlOperator.Type,
			config.MysqlRouter.Type,
		}
		if collection.Collect(mysqlTypeArr).Contains(sv.Type) {
			tries = 180 // Retry more times to wait msql start
		}
		input := k8s.IsAllPodReadyInput{
			PodLabel:    sv.GetPodLabel(v.Base),
			IgnoreEmpty: true,
		}
		err := k8sSv.WaitForAllPodReady(input, tries)
		if err != nil {
			return err
		}
	}
	return nil
}

// IsEtcdOK .
func (v *Vizion) IsEtcdOK() error {
	etcdMenbers := v.MasterNode().GetEtcdMembers()
	if len(etcdMenbers) < 3 {
		return fmt.Errorf("etcd members < 3")
	}
	for _, etcdM := range etcdMenbers {
		if !strings.Contains(etcdM, "started") {
			return fmt.Errorf("etcd status: %s", etcdM)
		}
	}

	return nil
}

// WaitForEtcdOK .
func (v *Vizion) WaitForEtcdOK() error {
	return v.RetryCheck(v.IsEtcdOK)
}

// IsCassOK .Check if cassandra 'nodetool status' UN  TODO
func (v *Vizion) IsCassOK() error {
	podNameArr := []string{}
	for _, vsetID := range v.Base.VsetIDs {
		vsetCassPodName := fmt.Sprintf("cassandra-vset%d-0", vsetID)
		podNameArr = append(podNameArr, vsetCassPodName)
	}

	cmdSpec := "/usr/bin/nodetool status | grep rack1"
	pattern := regexp.MustCompile(`(\S+)\s+(\d+.+rack\d)`)
	k8sSv := v.Service()
	downCassNum := 0
	for _, podName := range podNameArr {
		execInput := k8s.ExecInput{
			PodName:       podName,
			ContainerName: "cassandra",
			Command:       cmdSpec,
		}
		output, err := k8sSv.Exec(execInput)
		logger.Debug(utils.Prettify(output))
		if err != nil {
			return err
		}
		matched := pattern.FindAllStringSubmatch(output.Stdout, -1)
		for _, statArr := range matched {
			if statArr[1] != "UN" {
				logger.Warning(utils.Prettify(statArr[0]))
				downCassNum++
			}
			logger.Info(utils.Prettify(statArr[0]))
		}
	}
	if downCassNum > 0 {
		return fmt.Errorf("Has node cassandra not UN")
	}
	return nil
}

// WaitForCassOK .
func (v *Vizion) WaitForCassOK() error {
	return v.RetryCheck(v.IsCassOK)
}

// IsMysqlOK . Check if Mysql members ONLINE>= 3 && PRIMARY >=1 TODO
func (v *Vizion) IsMysqlOK() error {
	podNameArr := []string{"mysql-cluster-0"}
	k8sSv := v.Service()
	_, mysqlPwd := k8sSv.GetMysqlUserPassword()
	cmdSpec := fmt.Sprintf("mysql -p%s -e \"select * from performance_schema.replication_group_members;\"", mysqlPwd)

	patternOnline := regexp.MustCompile(`ONLINE`)
	patternOffline := regexp.MustCompile(`OFFLINE`)
	patternPrimary := regexp.MustCompile(`PRIMARY`)

	for _, podName := range podNameArr {
		onLineNum, offLineNum, primaryNum := 0, 0, 0
		execInput := k8s.ExecInput{
			PodName:       podName,
			ContainerName: "mysql",
			Command:       cmdSpec,
		}
		output, err := k8sSv.Exec(execInput)
		if err != nil {
			logger.Debug(utils.Prettify(output))
			return err
		}
		logger.Info(output.Stdout)
		matchedOnline := patternOnline.FindAllStringSubmatch(output.Stdout, -1)
		matchedOffline := patternOffline.FindAllStringSubmatch(output.Stdout, -1)
		matchedPrimary := patternPrimary.FindAllStringSubmatch(output.Stdout, -1)
		onLineNum = len(matchedOnline)
		offLineNum = len(matchedOffline)
		primaryNum = len(matchedPrimary)

		switch {
		case offLineNum > 0:
			return fmt.Errorf("Mysql has OFFLINE nodes")
		case onLineNum < 3:
			return fmt.Errorf("Mysql ONLINE nodes: %d < 3", onLineNum)
		case primaryNum == 0:
			return fmt.Errorf("Mysql has no PRIMARY node")
		default:
			logger.Infof("Check in pod %s: Mysql is OK", podName)
		}
	}
	return nil
}

// WaitForMysqlOK .
func (v *Vizion) WaitForMysqlOK() error {
	return v.RetryCheck(v.IsMysqlOK)
}

// IsDplHeloOK .
func (v *Vizion) IsDplHeloOK() error {
	var err error
	heloDplTypeArr := []int{
		config.Jddpl.Type,
		config.Servicedpl.Type,
		config.Mjcachedpl.Type,
		config.Djcachedpl.Type,
		config.Flushdpl.Type,
		config.Mcmapdpl.Type,
		config.Dcmapdpl.Type,
		config.Dpldagent.Type,
	}
	mCass := v.Cass().SetIndex("0")
	dplmanager := v.DplMgr(v.VaildMasterIP())
	for _, sv := range config.DefaultDplServiceArray {
		if collection.Collect(heloDplTypeArr).Contains(sv.Type) {
			dplArr := mCass.GetAllServices(sv.Type)
			for _, dpl := range dplArr {
				err = dplmanager.DplHelo(dpl.IP, dpl.Port)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// WaitForDplHeloOK .
func (v *Vizion) WaitForDplHeloOK() error {
	return v.RetryCheck(v.IsDplHeloOK)
}

// IsAllJnsPrimary .
func (v *Vizion) IsAllJnsPrimary(anchordplArr []Service) error {
	var err error
	// mCass := v.Cass().SetIndex("0")
	// anchordplArr := mCass.AnchordplServices()
	dplmanager := v.DplMgr(v.VaildMasterIP())
	for _, anchordpl := range anchordplArr {
		err = dplmanager.IsJnsStatPrimary(anchordpl.VsetID, anchordpl.ID, "")
		if err != nil {
			return err
		}
	}

	return nil
}

// WaitForAllJnsPrimary .
func (v *Vizion) WaitForAllJnsPrimary() error {
	mCass := v.Cass().SetIndex("0")
	anchordplArr := mCass.AnchordplServices()
	checkAction := func() error {
		return v.IsAllJnsPrimary(anchordplArr)
	}
	return v.RetryCheck(checkAction)
}

// IsAnyJnsPrimary .
func (v *Vizion) IsAnyJnsPrimary(anchordplArr []Service) error {
	var err error
	// mCass := v.Cass().SetIndex("0")
	// anchordplArr := mCass.AnchordplServices()
	dplmanager := v.DplMgr(v.VaildMasterIP())
	for _, anchordpl := range anchordplArr {
		err = dplmanager.IsJnsStatPrimary(anchordpl.VsetID, anchordpl.ID, "")
		if err != nil {
			continue
		}
		return nil
	}
	return fmt.Errorf("None of JNS is OK")
}

// WaitForAnyJnsPrimary .
func (v *Vizion) WaitForAnyJnsPrimary() error {
	mCass := v.Cass().SetIndex("0")
	anchordplArr := mCass.AnchordplServices()
	checkAction := func() error {
		return v.IsAnyJnsPrimary(anchordplArr)
	}
	return v.RetryCheck(checkAction)
}

// IsS3AnyJnsPrimary .
func (v *Vizion) IsS3AnyJnsPrimary(anchordplArr, servicedplArr []Service) error {
	var err error

	dplmanager := v.DplMgr(v.VaildMasterIP())
	s3dplIDs := []string{}
	for _, servicedpl := range servicedplArr {
		dplChArr, err := dplmanager.GetDplChannels(servicedpl.IP, servicedpl.Port)
		if err != nil {
			return err
		}
		for _, dplCh := range dplChArr {
			if dplCh["Channel_type"] == config.CHTYPES3 {
				s3dplIDs = append(s3dplIDs, servicedpl.ID)
			}
		}
	}
	if len(s3dplIDs) == 0 {
		return fmt.Errorf("Got None s3 channel")
	}

	for _, s3dplID := range s3dplIDs {
		for _, anchordpl := range anchordplArr {
			err = dplmanager.IsJnsStatPrimary(anchordpl.VsetID, anchordpl.ID, s3dplID)
			if err != nil {
				continue
			}
			return nil
		}
		return fmt.Errorf("None of JNS is OK")
	}
	return nil
}

// WaitForS3JnsPrimary .
func (v *Vizion) WaitForS3JnsPrimary() error {
	mCass := v.Cass().SetIndex("0")
	anchordplArr := mCass.AnchordplServices()
	servicedplArr := mCass.ServicedplServices()
	checkAction := func() error {
		return v.IsS3AnyJnsPrimary(anchordplArr, servicedplArr)
	}
	return v.RetryCheck(checkAction)
}

// IsZpoolStatusOK .Check if zpool status ONLINE
func (v *Vizion) IsZpoolStatusOK() error {
	bdNodeIPs := v.Service().BdNodeIPArr()
	for _, bdNodeIP := range bdNodeIPs {
		node := v.Node(bdNodeIP)
		err := node.IsZpoolStatusOK()
		if err != nil {
			return err
		}
	}
	return nil
}

// PrintDplChannels .
func (v *Vizion) PrintDplChannels() error {
	mCass := v.Cass().SetIndex("0")
	servicedplArr := mCass.ServicedplServices()
	dplmanager := v.DplMgr(v.VaildMasterIP())
	for _, servicedpl := range servicedplArr {
		dplmanager.PrintDplChannels(servicedpl.IP, servicedpl.Port)
	}
	return nil
}

// PrintDplBalance .
func (v *Vizion) PrintDplBalance() error {
	v.MasterNode().PrintDplBalance()
	return nil
}

// ============ Check health: Maintenance ============

// IsBdVolumeStatusExpected ...
func (v *Vizion) IsBdVolumeStatusExpected(expectStatus int, nodeName, nodeIP string, pvcNameArr []string) error {
	var err error
	bdServiceArr, err := v.Cass().SetIndex("0").GetServiceByType(config.Dpldagent.Type)
	if err != nil {
		return err
	}
	bdIDs := []string{}
	for _, bdSv := range bdServiceArr {
		bdIDs = append(bdIDs, bdSv.ID)
	}
	expectedVol := 0
	for _, vsetID := range v.Base.VsetIDs {
		subCass := v.Cass().SetIndex(strconv.Itoa(vsetID))
		volumeArr, err := subCass.GetVolume()
		if err != nil {
			return err
		}
		for _, vol := range volumeArr {
			// Filter by nodeName
			if nodeName != "" && nodeName != vol.CsiFlag {
				continue
			}
			// Filter by BlockDeviceService
			if vol.BlockDeviceService != "" && collection.Collect(bdIDs).Contains(vol.BlockDeviceService) {
				continue
			}
			// Filter by nodeIP
			if nodeIP != "" {
				bdServiceIP := ""
				for _, bdSv := range bdServiceArr {
					if bdSv.ID == vol.BlockDeviceService {
						bdServiceIP = bdSv.IP
						break
					}
				}
				if nodeIP != bdServiceIP {
					continue
				}
			}
			// Filter by pvcNameArr
			matchedVolName := []string{}
			for _, pvcName := range pvcNameArr {
				if strings.Contains(vol.Name, pvcName) {
					matchedVolName = append(matchedVolName, vol.Name)
				}
			}
			if len(matchedVolName) == 0 {
				continue
			}

			if vol.Status != expectStatus {
				expectedVol++
			} else {
				return fmt.Errorf("Wait for bd volumes status(acl/exp): %d/%d", vol.Status, expectStatus)
			}
		}
	}
	logger.Infof("All %d bd volumes status as expect: %d", expectedVol, expectStatus)
	return nil
}

// WaitBdVolumeStatusExpected ...
func (v *Vizion) WaitBdVolumeStatusExpected(expectStatus int, nodeName, nodeIP string, pvcNameArr []string) error {
	action := func(attempt uint) error {
		return v.IsBdVolumeStatusExpected(expectStatus, nodeName, nodeIP, pvcNameArr)
	}
	err := retry.Retry(
		action,
		strategy.Limit(30),
		strategy.Wait(30*time.Second),
	)
	return err
}

// WaitBlockDeviceRemoved ...
func (v *Vizion) WaitBlockDeviceRemoved(nodeName, nodeIP string, pvcNameArr []string) error {
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
	expectedVol := 0
	for _, vsetID := range v.Base.VsetIDs {
		subCass := cass.SetIndex(string(vsetID))
		volumeArr, err := subCass.GetVolume()
		if err != nil {
			return err
		}
		for _, vol := range volumeArr {
			// Filter by nodeName
			if nodeName != "" && nodeName != vol.CsiFlag {
				continue
			}
			// Filter by BlockDeviceService
			if vol.BlockDeviceService != "" && collection.Collect(bdIDs).Contains(vol.BlockDeviceService) {
				continue
			}

			// Skip if vol.status in [-1, 0]
			if collection.Collect([]int{-1, 0}).Contains(vol.Status) {
				continue
			}

			// Get bd service IP from volume <-> service
			bdServiceIP := ""
			for _, bdSv := range bdServiceArr {
				if bdSv.ID == vol.BlockDeviceService {
					bdServiceIP = bdSv.IP
					break
				}
			}

			// Filter by nodeIP
			if nodeIP != "" && bdServiceIP != "" && nodeIP != bdServiceIP {
				continue
			}

			// use nodeIP or bdServiceIP
			if bdServiceIP == "" {
				logger.Warning("Not got any bd service ip with: bdSv.ID==vol.BlockDeviceService")
				if nodeIP != "" {
					logger.Warning("Not got bd node ip too, skip check!")
					continue
				}
				bdServiceIP = nodeIP
			}

			// Filter by pvcNameArr
			matchedVolName := []string{}
			for _, pvcName := range pvcNameArr {
				if strings.Contains(vol.Name, pvcName) {
					matchedVolName = append(matchedVolName, vol.Name)
				}
			}
			if len(matchedVolName) == 0 {
				continue
			}

			var blockDeviceName string
			if vol.BlockDeviceName != "" {
				blockDeviceName = vol.BlockDeviceName
			} else {
				blockDeviceName = "/dev/dpl*"
			}

			node := v.Node(bdServiceIP)
			err = node.WaitDplDeviceRemoved(blockDeviceName)
			if err != nil {
				return err
			}
			expectedVol++
		}
	}
	logger.Infof("All %d bd volumes device already removed", expectedVol)
	return nil
}
