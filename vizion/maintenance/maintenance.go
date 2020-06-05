package maintenance

import (
	"pzatest/config"
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
	ApplyImage() error
	UpgradeCore() error
}

// Maint is used to interact with features provided by the  group.
type Maint struct {
	Vizion            resources.Vizion
	ServiceArr        []config.Service
	ExculdeServiceArr []config.Service
	BinaryArr         []config.Service
	CleanArr          []config.CleanItem
	Image             string
}

// MaintTestInput .
type MaintTestInput struct {
	SvNameArr        []string // service Name array
	ExculdeSvNameArr []string // service Name array
	BinNameArr       []string //  binary Name array
	CleanNameArr     []string //  clean item Name array
	Image            string
}

// NewMaint returns a Nodes
func NewMaint(base types.VizionBaseInput, mt MaintTestInput) *Maint {
	var svArr, binArr []config.Service
	var cleanArr []config.CleanItem

	// service Array
	if len(mt.SvNameArr) == 0 {
		svArr = config.DefaultCoreServiceArray
	} else {
		for _, item := range config.DefaultCoreServiceArray {
			if collection.Collect(mt.SvNameArr).Contains(item.Name) {
				svArr = append(svArr, item)
			}
		}
	}

	// binary Array
	if len(mt.BinNameArr) == 0 {
		binArr = svArr
	} else {
		for _, item := range config.DefaultDplBinaryArray {
			if collection.Collect(mt.BinNameArr).Contains(item.Name) {
				binArr = append(binArr, item)
			}
		}
	}

	// clean Array
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
		Image:      mt.Image,
	}
}

// Cleanup - maint
func (maint *Maint) Cleanup() error {
	var err error
	formatBD := false
	for _, clean := range maint.CleanArr {
		switch clean.Name {
		case "log":
			err = maint.Vizion.CleanLog(maint.ServiceArr)
			if err != nil {
				return err
			}
		case "journal":
			err = maint.Vizion.CleanEtcd(clean.Arg)
			if err != nil {
				return err
			}
			err = maint.Vizion.CleanJournal()
			if err != nil {
				return err
			}
		case "storage_cache":
			err = maint.Vizion.CleanStorageCache(clean.Arg[0], false)
			if err != nil {
				return err
			}
		case "master_cassandra":
			formatBD = true
			err = maint.Vizion.UpdateMasterCassTables()
			if err != nil {
				return err
			}
		case "sub_cassandra":
			formatBD = true
			err = maint.Vizion.CleanSubCassTables(clean.Arg)
			if err != nil {
				return err
			}
		case "etcd":
			formatBD = true
			err = maint.Vizion.CleanEtcd(clean.Arg)
			if err != nil {
				return err
			}
		case "cdcgc":
			err = maint.Vizion.CleanCdcgc()
			if err != nil {
				return err
			}
		}
	}

	if formatBD == true {
		err = maint.Vizion.UpdateSubCassTables()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop - maint
func (maint *Maint) Stop() error {
	var err error
	err = maint.Vizion.StopServices(maint.ServiceArr)
	if err != nil {
		return err
	}

	err = maint.Cleanup()
	if err != nil {
		return err
	}

	return nil
}

// Start - maint
func (maint *Maint) Start() error {
	// logger.Info(utils.Prettify(maint))
	err := maint.Vizion.StartServices(maint.ServiceArr)
	return err
}

// Restart - maint
func (maint *Maint) Restart() error {
	err := maint.Stop()
	if err != nil {
		return err
	}
	err = maint.Start()
	if err != nil {
		return err
	}

	return nil
}

// ApplyImage - maint TODO
func (maint *Maint) ApplyImage() error {
	err := maint.Stop()
	if err != nil {
		return err
	}

	err = maint.Vizion.ApplyServicesImage(maint.ServiceArr, maint.Image)
	if err != nil {
		return err
	}

	err = maint.Start()
	if err != nil {
		return err
	}

	return nil
}

// UpgradeCore - maint
func (maint *Maint) UpgradeCore() error {
	err := maint.Stop()
	if err != nil {
		return err
	}

	err = maint.ApplyImage()
	if err != nil {
		return err
	}

	return nil
}
