package schedule

import (
	"pzatest/libs/prettytable"
	"reflect"
	"runtime"
	"strconv"
	"strings"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// Action defines a callable function that package retry can handle.
type Action func() error

// Scheduler .
type Scheduler interface {
	SetUp(options ...OptionFunc) error
	TearDown(options ...OptionFunc) error
	RunPhase(action Action, options ...OptionFunc) error
}

// Input .
type Input struct {
	Verbosity int    // Print Phase in Teardown If > 0
	Skip      bool   // Skip the phase if true
	Desc      string // The phase description
}

// Phase .
type Phase struct {
	Idx    int    // The phase index, start with 1
	Name   string // The phase name
	Status string // The phase running status
	Desc   string // The phase description
}

// Schedule .
type Schedule struct {
	Input    Input   // Input args
	PhaseArr []Phase // Store the running phase list
}

// PrintPhase .
func (sc *Schedule) PrintPhase() error {
	lenIdx := 3
	lenName := 5
	lenStatus := 6
	lenDesc := 12
	for _, p := range sc.PhaseArr {
		if len(strconv.Itoa(p.Idx)) > lenIdx {
			lenIdx = len(strconv.Itoa(p.Idx))
		}
		if len(p.Name) > lenName {
			lenName = len(p.Name)
		}
		if len(p.Status) > lenStatus {
			lenStatus = len(p.Status)
		}
		if len(p.Desc) > lenDesc {
			lenDesc = len(p.Desc)
		}
	}

	table, _ := prettytable.NewTable(
		prettytable.Column{Header: "No.", AlignRight: false, MinWidth: lenIdx},
		prettytable.Column{Header: "Step", AlignRight: false, MinWidth: lenName},
		prettytable.Column{Header: "Result", AlignRight: false, MinWidth: lenStatus},
		prettytable.Column{Header: "Description", AlignRight: false, MinWidth: lenDesc},
	)
	table.Separator = " | "
	for _, p := range sc.PhaseArr {
		table.AddRow(p.Idx, p.Name, p.Status, p.Desc)
	}
	if len(table.Rows) > 0 {
		logger.Infof("Test Progress:\n%s", table.String())
	}

	return nil
}

// ApplyOptions Apply any given schedule options.
func (sc *Schedule) ApplyOptions(options ...OptionFunc) error {
	for _, fn := range options {
		if fn == nil {
			continue
		}
		if err := fn(sc); err != nil {
			return err
		}
	}
	return nil
}

// SetUp .
func (sc *Schedule) SetUp(options ...OptionFunc) error {
	// Initialize Input
	sc.Input = Input{
		Verbosity: 1,
	}
	sc.ApplyOptions(options...)

	status := "START"
	if sc.Input.Skip == true {
		status = "SKIP"
	}
	description := sc.Input.Desc
	idx := len(sc.PhaseArr) + 1
	pc, _, _, _ := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fName := f.Name()
	fNameSplit := strings.Split(fName, ".")
	pName := fNameSplit[len(fNameSplit)-1]
	phase := Phase{
		Idx:    idx,
		Name:   pName,
		Status: status,
		Desc:   description,
	}

	if idx <= 1 {
		sc.PhaseArr = append(sc.PhaseArr, phase)
	} else {
		lastPhase := sc.PhaseArr[idx-2]
		if lastPhase.Name != fName && lastPhase.Status != status {
			sc.PhaseArr = append(sc.PhaseArr, phase)
		}
	}

	if sc.Input.Verbosity > 0 {
		sc.PrintPhase()
	} else {
		logger.Infof("%s: %s", status, fName)
	}

	return nil
}

// TearDown .
func (sc *Schedule) TearDown(options ...OptionFunc) error {
	// Initialize Input
	sc.Input = Input{
		Verbosity: 1,
	}
	sc.ApplyOptions(options...)
	if sc.Input.Skip == true {
		return nil
	}

	pc, _, _, ok := runtime.Caller(1)
	f := runtime.FuncForPC(pc)
	fName := f.Name()
	fNameSplit := strings.Split(fName, ".")
	pName := fNameSplit[len(fNameSplit)-1]
	status := "PASS"
	if ok == false {
		status = "FAIL"
	}

	description := sc.Input.Desc
	idx := len(sc.PhaseArr) + 1

	phase := Phase{
		Idx:    idx,
		Name:   pName,
		Status: status,
		Desc:   description,
	}
	sc.PhaseArr = append(sc.PhaseArr, phase)

	if sc.Input.Verbosity > 0 {
		sc.PrintPhase()
	} else {
		logger.Infof("%s: %s", status, fName)
	}

	return nil
}

// RunPhase .
func (sc *Schedule) RunPhase(action Action, options ...OptionFunc) error {
	// Initialize Input
	sc.Input = Input{
		Verbosity: 1,
	}
	sc.ApplyOptions(options...)

	status := "START"
	if sc.Input.Skip == true {
		status = "SKIP"
	}
	description := sc.Input.Desc
	idx := len(sc.PhaseArr) + 1
	fName := strings.TrimSuffix(runtime.FuncForPC(reflect.ValueOf(action).Pointer()).Name(), "-fm")
	fNameSplit := strings.Split(fName, ".")
	pName := fNameSplit[len(fNameSplit)-1]
	phase := Phase{
		Idx:    idx,
		Name:   pName,
		Status: status,
		Desc:   description,
	}

	if idx <= 1 {
		sc.PhaseArr = append(sc.PhaseArr, phase)
	} else {
		lastPhase := sc.PhaseArr[idx-2]
		if lastPhase.Name != pName && lastPhase.Status != status {
			sc.PhaseArr = append(sc.PhaseArr, phase)
		}
	}

	if sc.Input.Verbosity > 0 {
		sc.PrintPhase()
	} else {
		logger.Infof("%s: %s", status, fName)
	}
	// Run func
	err := action()
	status = "PASS"
	if err != nil {
		status = "FAIL"
	}
	sc.PhaseArr[idx-1].Status = status
	if sc.Input.Verbosity > 0 {
		sc.PrintPhase()
	} else {
		logger.Infof("%s: %s", status, fName)
	}
	return err
}
