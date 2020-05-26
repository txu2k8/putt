package resources

import (
	"pzatest/config"
	"pzatest/libs/utils"
	"pzatest/types"

	"github.com/chenhg5/collection"
)

// CleanLog .
func CleanLog(conf types.VizionBaseInput) {
	vizion := VizionBase{VizionBaseInput: conf}
	logPathArr := []string{}
	for _, sv := range config.DefaultServiceArray {
		logArr := sv.GetLogDirArr(conf)
		// logger.Info(utils.Prettify(logArr))
		logPathArr = append(logPathArr, logArr...)
	}
	for _, nodeIP := range vizion.Service().GetAllNodeIPs() {
		node := vizion.Node(nodeIP)
		node.CleanLog(logPathArr)
	}
}

// StopService .
func (vizion *VizionBase) StopService(serviceTypeArr []int) error {
	for _, sv := range config.DefaultServiceArray {
		// logger.Info(utils.Prettify(sv))
		if collection.Collect(serviceTypeArr).Contains(sv.Type) {
			logger.Infof(">> Stop service %s:%d ...", sv.TypeName, sv.Type)
			ipArr, _ := vizion.Cass().SetIndex("0").GetServiceByType(sv.Type)
			logger.Info(utils.Prettify(ipArr))
		}
	}
	return nil
}
