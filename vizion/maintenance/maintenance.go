package maintenance

import (
	"pzatest/config"
	"pzatest/libs/utils"
	"pzatest/types"
	"pzatest/vizion/resources"

	"github.com/chenhg5/collection"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// Maintainer for maintenance ops
type Maintainer interface {
	Stop() error
	Start() error
	Restart() error
	UpgradeCore() error
}

// Maint is used to interact with features provided by the  group.
type Maint struct {
	Vizion            resources.Vizion
	ServiceArr        []config.Service
	ExculdeServiceArr []config.Service
	BinaryArr         []config.Service
	CleanArr          []config.CleanItem
}

// MaintTestInput .
type MaintTestInput struct {
	SvNameArr        []string // service Name array
	ExculdeSvNameArr []string // service Name array
	BinNameArr       []string //  binary Name array
	CleanNameArr     []string //  clean item Name array
}

// NewMaint returns a Nodes
func NewMaint(base types.VizionBaseInput, mt MaintTestInput) *Maint {
	var svArr, binArr []config.Service
	var cleanArr []config.CleanItem
	if len(mt.SvNameArr) == 0 {
		svArr = config.DefaultCoreServiceArray
	} else {
		for _, item := range config.DefaultCoreServiceArray {
			if collection.Collect(mt.SvNameArr).Contains(item.Name) {
				svArr = append(svArr, item)
			}
		}
	}

	if len(mt.BinNameArr) == 0 {
		binArr = svArr
	} else {
		for _, item := range config.DefaultDplBinaryArray {
			if collection.Collect(mt.BinNameArr).Contains(item.Name) {
				binArr = append(binArr, item)
			}
		}
	}

	if len(mt.CleanNameArr) == 0 {
		cleanArr = []config.CleanItem{}
	} else if collection.Collect(mt.CleanNameArr).Contains("all") {
		cleanArr = config.DefaultCleanArray
	} else {
		for _, item := range config.DefaultCleanArray {
			if collection.Collect(mt.CleanNameArr).Contains(item.Name) {
				cleanArr = append(cleanArr, item)
			}
		}
	}

	return &Maint{
		Vizion:     resources.Vizion{Base: base},
		ServiceArr: svArr,
		BinaryArr:  binArr,
		CleanArr:   cleanArr,
	}
}

// Stop - maint
func (maint *Maint) Stop() error {
	logger.Info(utils.Prettify(maint))
	for _, sv := range maint.ServiceArr {
		// logger.Info(utils.Prettify(sv))
		err := maint.Vizion.StopService(sv)
		if err != nil {
			return err
		}
	}
	return nil
}

// Start - maint
func (maint *Maint) Start() error {

	return nil
}

// Restart - maint
func (maint *Maint) Restart() error {
	maint.Stop()
	maint.Start()
	return nil
}

// UpgradeCore - maint
func (maint *Maint) UpgradeCore() error {
	maint.Stop()
	maint.Start()
	return nil
}
