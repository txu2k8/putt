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
	"runtime"
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
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fName := f.Name()

	logger.Infof("Enter %s ...", fName)
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
	// err = v.Schedule.RunPhase(v.WaitForPingOK, schedule.Desc("Check if nodes ping OK"))
	// if err != nil {
	// 	return err
	// }

	// err = v.Schedule.RunPhase(v.IsNodeCrashed, schedule.Desc("Check nodes /var/crash/ files"))
	// if err != nil {
	// 	return err
	// }

	// err = v.Schedule.RunPhase(v.IsServiceCoreDump, schedule.Desc("Check if service has core dump files"))
	// if err != nil {
	// 	return err
	// }

	// err = v.Schedule.RunPhase(v.WaitForEtcdOK, schedule.Desc("Check if etcd members stared"))
	// if err != nil {
	// 	return err
	// }

	// err = v.Schedule.RunPhase(v.WaitForCassOK, schedule.Desc("Check if cassandra 'nodetool status' UN"))
	// if err != nil {
	// 	return err
	// }

	// err = v.Schedule.RunPhase(v.WaitForDplHeloOK, schedule.Desc("Check if 'dplmanager -mdpl helo' OK"))
	// if err != nil {
	// 	return err
	// }

	err = v.Schedule.RunPhase(v.IsAllJnsPrimary, schedule.Desc("Check if all 'dplmanager -mjns stat' PRIMARY"))
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

// IsCassOK run "nodetool status" in cassandra pods  TODO
func (v *Vizion) IsCassOK() error {
	podNameArr := []string{}
	for _, vsetID := range v.Base.VsetIDs {
		vsetCassPodName := fmt.Sprintf("cassandra-vset%d-0", vsetID)
		podNameArr = append(podNameArr, vsetCassPodName)
	}

	cmdSpec := "/usr/bin/nodetool status | grep rack1"
	pattern := regexp.MustCompile(`(\S+)\s+(\d+.+rack\d)`)
	vk8s := v.Service()
	for _, podName := range podNameArr {
		execInput := k8s.ExecInput{
			PodName:       podName,
			ContainerName: "cassandra",
			Command:       cmdSpec,
		}
		output, err := vk8s.Exec(execInput)
		logger.Info(output)
		if err != nil {
			return err
		}
		matched := pattern.FindAllStringSubmatch(output.Stdout, -1)
		logger.Info(matched)
		nodeToolStatusArr := []string{}
		for _, stat := range nodeToolStatusArr {
			if stat != "UN" {
				logger.Warning(stat)
				return fmt.Errorf("Has node cassandra not UN")
			}
		}
	}

	return nil
}

// WaitForCassOK .
func (v *Vizion) WaitForCassOK() error {
	return v.RetryCheck(v.IsCassOK)
}

// IsMysqlOK . TODO
func (v *Vizion) IsMysqlOK() error {
	return nil
}

// IsDplHeloOK .
func (v *Vizion) IsDplHeloOK() error {
	var err error
	mCass := v.Cass().SetIndex("0")
	dplmanager := v.DplMgr(v.VaildMasterIP())
	for _, vsetID := range v.Base.VsetIDs {
		for _, sv := range config.DefaultDplServiceArray {
			switch sv.Type {
			case config.Dplmanager.Type:
				// skip
			default:
				dplArr, _ := mCass.GetServiceByTypeVsetID(sv.Type, vsetID)
				for _, dpl := range dplArr {
					logger.Infof(utils.Prettify(dpl))
					err = dplmanager.DplHelo(dpl.IP, dpl.Port)
					if err != nil {
						return err
					}
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
func (v *Vizion) IsAllJnsPrimary() error {
	// mCass := v.Cass().SetIndex("0")
	// // dplmanager := v.DplMgr(v.VaildMasterIP())
	// for _, vsetID := range v.Base.VsetIDs {
	// 	anchordplArr, _ := mCass.GetServiceByTypeVsetID()
	// 	// dplmanager.GetJnsStat()
	// }

	return nil
}

// IsAnyJnsPrimary .
func (v *Vizion) IsAnyJnsPrimary() error {
	return nil
}

// IsZpoolStatusOK .
func (v *Vizion) IsZpoolStatusOK() error {
	return nil
}

// PrintDplChannels .
func (v *Vizion) PrintDplChannels() error {
	return nil
}

// PrintDplBalance .
func (v *Vizion) PrintDplBalance() error {
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
		subCass := v.Cass().SetIndex(string(vsetID))
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
