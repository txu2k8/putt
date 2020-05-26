package maintenance

import (
	"pzatest/types"
	"pzatest/vizion/resources"

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
	Vizion         resources.VizionBase
	ServiceTypeArr []int
}

// NewMaint returns a Nodes
func NewMaint(conf types.VizionBaseInput) *Maint {
	return &Maint{Vizion: resources.VizionBase{VizionBaseInput: conf}, ServiceTypeArr: []int{1024}}
}

// Stop - maint
func (maint *Maint) Stop() error {
	maint.Vizion.StopService(maint.ServiceTypeArr)
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
