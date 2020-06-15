package resources

import (
	"fmt"
	"pzatest/config"
	"pzatest/libs/retry"
	"pzatest/libs/retry/strategy"
	"pzatest/libs/utils"
	"strings"
	"time"

	"github.com/chenhg5/collection"
)

// IsNodeCrashed ...
func (v *Vizion) IsNodeCrashed() error {
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
