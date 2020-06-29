package maintenance

import (
	"errors"
	"fmt"
	"os"
	"path"
	"putt/config"
	"putt/libs/convert"
	"putt/libs/git"
	"putt/libs/runner/schedule"
	"putt/types"
	"putt/vizion/resources"
	"strings"
	"time"

	"github.com/chenhg5/collection"
	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// Maintainer for maintenance ops
type Maintainer interface {
	Cleanup() error
	Stop() error
	StopC() error
	Start() error
	Restart() error
	MakeBinary() error
	MakeImage() error
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
	GitCfg            GitInput
	ServiceNameArr    []string
	BinaryNameArr     []string
	CleanNameArr      []string
	Schedule          schedule.Schedule
}

// GitInput ...
type GitInput struct {
	BuildServerIP   string // the git project server IP address
	BuildServerUser string
	BuildServerPwd  string
	BuildServerKey  string
	BuildPath       string // the git procject path
	BuildNum        string // the build number for tag, eg: 2.1.0.133, used in JENKINS
	Pull            bool   // git pull before tag?
	Tag             bool   // git tag && push  ?
	Make            bool   // exec make file ?
	LocalBinPath    string // local path for store dpl binarys
}

// MaintTestInput .
type MaintTestInput struct {
	SvNameArr         []string // service Name array
	ExculdeSvNameArr  []string // service Name array
	BinNameArr        []string //  binary Name array
	ExculdeBinNameArr []string //  binary Name array
	CleanNameArr      []string //  clean item Name array
	Image             string   // eg: registry.ai/stable:tag
	GitCfg            GitInput // The build number for image tag name, used in JENKINS
}

// NewMaint returns a Nodes
func NewMaint(base types.VizionBaseInput, mt MaintTestInput) *Maint {
	var svArr, binArr []config.Service
	var cleanArr []config.CleanItem

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

	cleanNameArr := []string{}
	for _, clean := range cleanArr {
		cleanNameArr = append(cleanNameArr, clean.Name)
	}

	// service Array
	if len(mt.SvNameArr) == 0 {
		svArr = config.DefaultCoreServiceArray
		if !collection.Collect(mt.CleanNameArr).Contains("all") {
			// Skip Opt cdcgc pods
			newSvArr := []config.Service{}
			cdcgcTypeArr := []int{config.Cdcgcbd.Type, config.Cdcgcs3.Type}
			for _, sv := range svArr {
				if !collection.Collect(cdcgcTypeArr).Contains(sv.Type) {
					newSvArr = append(newSvArr, sv)
				}
			}
			svArr = newSvArr
		}
	} else {
		for _, item := range config.DefaultCoreServiceArray {
			if collection.Collect(mt.SvNameArr).Contains(item.Name) {
				svArr = append(svArr, item)
			}
		}
	}

	svNameArr := []string{}
	for _, sv := range svArr {
		svNameArr = append(svNameArr, sv.Name)
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

	binNameArr := []string{}
	for _, bin := range binArr {
		binNameArr = append(binNameArr, bin.Name)
	}

	return &Maint{
		Vizion:         resources.Vizion{Base: base},
		ServiceArr:     svArr,
		BinaryArr:      binArr,
		CleanArr:       cleanArr,
		Image:          mt.Image,
		GitCfg:         mt.GitCfg,
		ServiceNameArr: svNameArr,
		BinaryNameArr:  binNameArr,
		CleanNameArr:   cleanNameArr,
	}
}

// isImageOK - Check is image OK on gitlab
func (maint *Maint) isImageOK() error {
	logger.Infof("Wait for Image Availabel: %s", maint.Image)
	if maint.Image == "" {
		return errors.New("image name is nul")
	}
	tagName := strings.Split(maint.Image, ":")[1]

	// wait for image OK on gitlab
	cfg := git.GitlabConfig{
		BaseURL: "http://gitlab.panzura.com",
		Token:   "xjB1FHHyJHNQUhgy7K7t",
	}
	projectID := 25
	gitlabMgr := git.NewGitlabClient(cfg)
	err := gitlabMgr.IsPipelineJobsSuccess(projectID, tagName)
	if err != nil {
		return err
	}
	logger.Infof("Image Availabel: %s", maint.Image)
	return nil
}

// setImage - maint set sts/deployment/xxx Image
func (maint *Maint) setImage() error {
	var err error
	err = maint.Vizion.ApplyServicesImage(maint.ServiceArr, maint.Image)
	if err != nil {
		return err
	}

	err = maint.Vizion.ApplyDplmanagerShellImage(maint.Image)
	if err != nil {
		return err
	}
	return nil
}

// MakeBinary - maint
func (maint *Maint) MakeBinary() error {
	var err error
	strTime := time.Now().Format("2006-01-02-15-04-05")
	// Get build path branch Name, joint tagName
	gitMgr := git.NewGitMgr(
		maint.GitCfg.BuildServerIP,
		maint.GitCfg.BuildServerUser,
		maint.GitCfg.BuildServerPwd,
		maint.GitCfg.BuildServerKey,
	)
	branchName := gitMgr.GetCurrentBranch(maint.GitCfg.BuildPath)
	tagName := strTime + "-" + branchName
	if maint.GitCfg.BuildNum == "" {
		tagName = tagName + "_private"
	} else {
		tagName = tagName + "_" + maint.GitCfg.BuildNum
	}

	localBinPath := path.Join(maint.GitCfg.LocalBinPath, tagName)
	err = os.MkdirAll(localBinPath, os.ModePerm)
	if err != nil {
		logger.Panic(err)
	}

	// pull && save changelog
	if maint.GitCfg.Pull == true {
		if err = gitMgr.Pull(maint.GitCfg.BuildPath); err != nil {
			return err
		}
		// change.log
		date, changeLog := gitMgr.GetChangeLog(maint.GitCfg.BuildPath)
		changeLogFile := path.Join(localBinPath, "change.log")
		logger.Info(changeLogFile)
		file, err := os.OpenFile(changeLogFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			logger.Panic(err)
		}
		defer file.Close()
		file.WriteString(date + "\n")
		file.WriteString("Version:" + tagName + "\n")
		file.WriteString("Change logs:\n" + changeLog)
	}

	// make file
	if maint.GitCfg.Make == true {
		for _, bin := range maint.BinaryArr {
			binGitPath := path.Join(maint.GitCfg.BuildPath, bin.GitPath)
			binName := bin.Name
			binGitPathName := path.Join(binGitPath, binName)
			binLocalPathName := path.Join(localBinPath, binName)
			if md5sum := gitMgr.MakeFile(binGitPath, binName); md5sum == "" {
				return fmt.Errorf("%s make failed", binName)
			}
			gitMgr.ConnectSftpClient()
			if err = gitMgr.ScpGet(binLocalPathName, binGitPathName); err != nil {
				return err
			}
		}
	}
	logger.Infof("Local Binary Path: %s", localBinPath)
	return nil
}

// MakeImage - maint make image by tag to gitlab
func (maint *Maint) MakeImage() error {
	var err error
	strTime := time.Now().Format("2006-01-02-15-04-05")
	// Get build path branch Name, joint tagName
	gitMgr := git.NewGitMgr(
		maint.GitCfg.BuildServerIP,
		maint.GitCfg.BuildServerUser,
		maint.GitCfg.BuildServerPwd,
		maint.GitCfg.BuildServerKey,
	)
	branchName := gitMgr.GetCurrentBranch(maint.GitCfg.BuildPath)
	tagName := strTime + "-" + branchName
	if maint.GitCfg.BuildNum == "" {
		tagName = tagName + "_notest"
	} else {
		tagName = tagName + "_" + maint.GitCfg.BuildNum
	}

	// pull && save changelog
	if maint.GitCfg.Pull == true {
		if err = gitMgr.Pull(maint.GitCfg.BuildPath); err != nil {
			return err
		}
		// change.log
		date, changeLog := gitMgr.GetChangeLog(maint.GitCfg.BuildPath)
		localBinPath := path.Join(maint.GitCfg.LocalBinPath, tagName)
		err := os.MkdirAll(localBinPath, os.ModePerm)
		if err != nil {
			logger.Panic(err)
		}
		changeLogFile := path.Join(localBinPath, "change.log")
		logger.Info(changeLogFile)
		file, err := os.OpenFile(changeLogFile, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
		if err != nil {
			logger.Panic(err)
		}
		defer file.Close()
		file.WriteString(date + "\n")
		file.WriteString("Version/Tag:" + tagName + "\n")
		file.WriteString("Change logs:\n" + changeLog)
	}

	// tag && push
	if maint.GitCfg.Tag == true {
		if err = gitMgr.Tag(maint.GitCfg.BuildPath, tagName); err != nil {
			return err
		}
	}

	maint.Image = config.RemoteDplRegistry + ":" + tagName

	// wait for maint.Image OK on gitlab
	if err = maint.isImageOK(); err != nil {
		return err
	}
	// logger.Info(utils.Prettify(maint))
	return nil
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
		case "etcd":
			err = maint.Vizion.CleanEtcd(clean.Arg)
			if err != nil {
				return err
			}
		case "j_device":
			err = maint.Vizion.CleanJdevice()
			if err != nil {
				return err
			}
			err = maint.Vizion.IsJnlFormatSuccess()
			if err != nil {
				return err
			}
		case "storage_cache":
			err = maint.Vizion.CleanStorageCache(clean.Arg[0], false)
			if err != nil {
				return err
			}
		case "master_cassandra":
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

		case "cdcgc":
			err = maint.Vizion.CleanCdcgc()
			if err != nil {
				return err
			}
		}
	}

	if formatBD == true {
		err = maint.Vizion.FormatBdVolume()
		if err != nil {
			return err
		}
	}
	return nil
}

// Stop - maint
func (maint *Maint) Stop() error {
	// stop service order by ReverseServiceArr
	stopServiceArr := config.ReverseServiceArr(maint.ServiceArr)
	return maint.Vizion.StopServices(stopServiceArr)
}

// StopC - maint: Stop -> Cleanup
func (maint *Maint) StopC() error {
	var err error
	// Stop
	stopSvNameArr := strings.Join(convert.ReverseStringArr(maint.ServiceNameArr), ",")
	err = maint.Schedule.RunPhase(maint.Stop, schedule.Desc(stopSvNameArr))
	if err != nil {
		return err
	}

	// Cleanup
	strClNameArr := strings.Join(maint.CleanNameArr, ",")
	skipCleanup := false
	if len(maint.CleanNameArr) < 0 {
		skipCleanup = true
	}
	err = maint.Schedule.RunPhase(maint.Cleanup, schedule.Desc(strClNameArr), schedule.Skip(skipCleanup))
	if err != nil {
		return err
	}

	return err
}

// Start - maint
func (maint *Maint) Start() error {
	// logger.Info(utils.Prettify(maint))
	return maint.Vizion.StartServices(maint.ServiceArr)
}

// Restart - maint: Stop -> Cleanup -> Start
func (maint *Maint) Restart() error {
	var err error

	// Stop
	stopSvNameArr := strings.Join(convert.ReverseStringArr(maint.ServiceNameArr), ",")
	err = maint.Schedule.RunPhase(maint.Stop, schedule.Desc(stopSvNameArr))
	if err != nil {
		return err
	}

	// Cleanup
	strClNameArr := strings.Join(maint.CleanNameArr, ",")
	skipCleanup := false
	if len(maint.CleanNameArr) < 0 {
		skipCleanup = true
	}
	err = maint.Schedule.RunPhase(maint.Cleanup, schedule.Desc(strClNameArr), schedule.Skip(skipCleanup))
	if err != nil {
		return err
	}

	// Start
	startSvNameArr := strings.Join(maint.ServiceNameArr, ",")
	err = maint.Schedule.RunPhase(maint.Start, schedule.Desc(startSvNameArr))
	if err != nil {
		return err
	}

	return nil
}

// ApplyImage - maint: Stop -> Cleanup -> setImage -> Start
func (maint *Maint) ApplyImage() error {
	var err error
	// isImageOK: Wait for image OK on gitlab
	err = maint.Schedule.RunPhase(maint.isImageOK, schedule.Desc(maint.Image))
	if err != nil {
		return err
	}

	// Stop
	stopSvNameArr := strings.Join(convert.ReverseStringArr(maint.ServiceNameArr), ",")
	err = maint.Schedule.RunPhase(maint.Stop, schedule.Desc(stopSvNameArr))
	if err != nil {
		return err
	}

	// Cleanup
	strClNameArr := strings.Join(maint.CleanNameArr, ",")
	skipCleanup := false
	if len(maint.CleanNameArr) < 0 {
		skipCleanup = true
	}
	err = maint.Schedule.RunPhase(maint.Cleanup, schedule.Desc(strClNameArr), schedule.Skip(skipCleanup))
	if err != nil {
		return err
	}

	// setImage
	err = maint.Schedule.RunPhase(maint.setImage, schedule.Desc(maint.Image))
	if err != nil {
		return err
	}

	// Start
	startSvNameArr := strings.Join(maint.ServiceNameArr, ",")
	err = maint.Schedule.RunPhase(maint.Start, schedule.Desc(startSvNameArr))
	if err != nil {
		return err
	}

	return nil
}

// UpgradeCore - maint: MakeImage -> Stop -> Cleanup -> setImage -> Start
func (maint *Maint) UpgradeCore() error {
	var err error

	// MakeImage
	err = maint.Schedule.RunPhase(maint.MakeImage)
	if err != nil {
		return err
	}

	err = maint.ApplyImage()
	if err != nil {
		return err
	}

	return nil
}
