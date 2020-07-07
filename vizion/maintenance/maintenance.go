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
	CheckHealth() error
	Cleanup() error
	Stop() error
	StopC() error
	Start() error
	Restart() error
	CreateImageByBinary() error // docker build: make
	MakeImage() error           // gitlab build: tag
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
	Check             bool
	ServiceNameArr    []string
	BinaryNameArr     []string
	CleanNameArr      []string
	SkipSteps         []string
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
	Check             bool     // check health before stop and after start services
	SkipSteps         []string // Skip steps "stop | cleanup | start" for apply_image and upgradecore
}

// NewMaint returns a Nodes
func NewMaint(base types.BaseInput, mt MaintTestInput) *Maint {
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
		Check:          mt.Check,
		ServiceNameArr: svNameArr,
		BinaryNameArr:  binNameArr,
		CleanNameArr:   cleanNameArr,
		SkipSteps:      mt.SkipSteps,
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

// tagImageLocal - tag the image to local
func (maint *Maint) tagImageLocal() error {
	if config.LocalDplRegistry == "" {
		return nil
	}

	localImage := strings.Replace(maint.Image, config.RemoteDplRegistry, config.LocalDplRegistry, 1)
	logger.Infof("Tag Image local: %s -> %s", maint.Image, localImage)

	// docker pull
	node := maint.Vizion.MasterNode()
	_, output := node.RunCmd("docker pull " + maint.Image)
	if strings.Contains(output, "Error") {
		return fmt.Errorf(output)
	}

	// docker tag -> push
	if localImage != maint.Image {
		_, output = node.RunCmd(fmt.Sprintf("docker tag %s %s", maint.Image, localImage))
		if strings.Contains(output, "Error") {
			return fmt.Errorf(output)
		}

		_, output = node.RunCmd(fmt.Sprintf("docker push %s", localImage))
		if strings.Contains(output, "Error") {
			return fmt.Errorf(output)
		}
		maint.Image = localImage
		logger.Infof("Local Image Availabel: %s", maint.Image)
	}

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

// makeBinary - maint
func (maint *Maint) makeBinary() (localBinPath string, err error) {
	logger.Info("Make binarys ...")
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

	localBinPath = path.Join(maint.GitCfg.LocalBinPath, tagName)
	err = os.MkdirAll(localBinPath, os.ModePerm)
	if err != nil {
		logger.Panic(err)
	}

	// pull && save changelog
	if maint.GitCfg.Pull == true {
		if err = gitMgr.Pull(maint.GitCfg.BuildPath); err != nil {
			return
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
				err = fmt.Errorf("%s make failed", binName)
				return
			}
			gitMgr.ConnectSftpClient()
			if err = gitMgr.ScpGet(binLocalPathName, binGitPathName); err != nil {
				return
			}
		}
	}
	logger.Infof("Local Binary Path: %s", localBinPath)
	return
}

// CreateImageByBinary - maint Make Image by docker build with Binrarys and Dockerfile --TODO
func (maint *Maint) CreateImageByBinary() error {
	localBinPath, err := maint.makeBinary()
	if err != nil {
		return err
	}
	logger.Info("Docker build image with binarys and Dockerfile ...")
	tagName := path.Base(localBinPath)
	maint.Image = config.RemoteDplRegistry + ":" + tagName
	return nil
}

// MakeImage - maint Make image by push tag to gitlab
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
	if err = gitMgr.Tag(maint.GitCfg.BuildPath, tagName); err != nil {
		return err
	}

	maint.Image = config.RemoteDplRegistry + ":" + tagName

	// wait for maint.Image OK on gitlab
	if err = maint.isImageOK(); err != nil {
		return err
	}
	// logger.Info(utils.Prettify(maint))
	return nil
}

// CheckHealth - maint CheckHealth before stop, and after start
func (maint *Maint) CheckHealth() error {
	// Check Health
	if maint.Check == true {
		return maint.Vizion.CheckHealth()
	}
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
		case "j_device":
			err = maint.Vizion.CleanJdevice()
			if err != nil {
				return err
			}
		case "etcd":
			err = maint.Vizion.CleanEtcd(clean.Arg)
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

// Start - maint
func (maint *Maint) Start() error {
	// logger.Info(utils.Prettify(maint))
	cleanJdevice, cleanSC := false, false
	for _, clean := range maint.CleanArr {
		switch clean.Name {
		case "j_device":
			cleanJdevice = true
		case "storage_cache":
			cleanSC = true
		default:
			continue
		}
	}
	return maint.Vizion.StartServices(maint.ServiceArr, cleanJdevice, cleanSC)
}

// StopC - maint: Stop -> Cleanup
func (maint *Maint) StopC() error {
	var err error
	// Check Health
	err = maint.CheckHealth()
	if err != nil {
		return err
	}

	// Stop
	stopSvNameArr := strings.Join(convert.ReverseStringArr(maint.ServiceNameArr), ",")
	err = maint.Vizion.Schedule.RunPhase(maint.Stop, schedule.Desc(stopSvNameArr))
	if err != nil {
		return err
	}

	// Cleanup
	strClNameArr := strings.Join(maint.CleanNameArr, ",")
	skipCleanup := false
	if len(maint.CleanNameArr) == 0 {
		skipCleanup = true
	}
	err = maint.Vizion.Schedule.RunPhase(maint.Cleanup, schedule.Desc(strClNameArr), schedule.Skip(skipCleanup))
	if err != nil {
		return err
	}

	return err
}

// Restart - maint: Stop -> Cleanup -> Start
func (maint *Maint) Restart() error {
	var err error

	// Check Health
	err = maint.CheckHealth()
	if err != nil {
		return err
	}

	// Stop
	stopSvNameArr := strings.Join(convert.ReverseStringArr(maint.ServiceNameArr), ",")
	err = maint.Vizion.Schedule.RunPhase(maint.Stop, schedule.Desc(stopSvNameArr))
	if err != nil {
		return err
	}

	// Cleanup
	strClNameArr := strings.Join(maint.CleanNameArr, ",")
	skipCleanup := false
	if len(maint.CleanNameArr) == 0 {
		skipCleanup = true
	}
	err = maint.Vizion.Schedule.RunPhase(maint.Cleanup, schedule.Desc(strClNameArr), schedule.Skip(skipCleanup))
	if err != nil {
		return err
	}

	// Start
	startSvNameArr := strings.Join(maint.ServiceNameArr, ",")
	err = maint.Vizion.Schedule.RunPhase(maint.Start, schedule.Desc(startSvNameArr))
	if err != nil {
		return err
	}

	// Check Health
	err = maint.CheckHealth()
	if err != nil {
		return err
	}

	return nil
}

// ApplyImage - maint: Stop -> Cleanup -> setImage -> Start
func (maint *Maint) ApplyImage() error {
	var err error
	// isImageOK: Wait for image OK on gitlab
	err = maint.Vizion.Schedule.RunPhase(maint.isImageOK, schedule.Desc(maint.Image))
	if err != nil {
		return err
	}

	// Check Health
	err = maint.CheckHealth()
	if err != nil {
		return err
	}

	// Stop
	if !collection.Collect(maint.SkipSteps).Contains("stop") {
		stopSvNameArr := strings.Join(convert.ReverseStringArr(maint.ServiceNameArr), ",")
		err = maint.Vizion.Schedule.RunPhase(maint.Stop, schedule.Desc(stopSvNameArr))
		if err != nil {
			return err
		}
	}

	// Cleanup
	if !collection.Collect(maint.SkipSteps).Contains("cleanup") {
		strClNameArr := strings.Join(maint.CleanNameArr, ",")
		skipCleanup := false
		if len(maint.CleanNameArr) == 0 {
			skipCleanup = true
		}
		err = maint.Vizion.Schedule.RunPhase(maint.Cleanup, schedule.Desc(strClNameArr), schedule.Skip(skipCleanup))
		if err != nil {
			return err
		}
	}

	// tagImageLocal
	err = maint.Vizion.Schedule.RunPhase(maint.tagImageLocal, schedule.Desc(maint.Image))
	if err != nil {
		return err
	}
	// setImage
	err = maint.Vizion.Schedule.RunPhase(maint.setImage, schedule.Desc(maint.Image))
	if err != nil {
		return err
	}

	// Start
	startSvNameArr := strings.Join(maint.ServiceNameArr, ",")
	err = maint.Vizion.Schedule.RunPhase(maint.Start, schedule.Desc(startSvNameArr))
	if err != nil {
		return err
	}

	// Check Health
	err = maint.CheckHealth()
	if err != nil {
		return err
	}

	return nil
}

// UpgradeCore - maint: MakeImage -> Stop -> Cleanup -> setImage -> Start
func (maint *Maint) UpgradeCore() error {
	var err error

	if maint.GitCfg.Tag == true {
		// Make Image by git tag
		err = maint.Vizion.Schedule.RunPhase(maint.MakeImage, schedule.Desc("Make Image by git tag"))
		if err != nil {
			return err
		}
	} else if maint.GitCfg.Make == true {
		// Make Image by docker build with Binrarys and Dockerfile
		err = maint.Vizion.Schedule.RunPhase(maint.CreateImageByBinary, schedule.Desc("Make Image by docker build with Binrarys and Dockerfile"))
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("--tag or --make Flags required: --tag(make image by push tag to gitlab), --make(make image by make_binary->docker_build)")
	}

	err = maint.ApplyImage()
	if err != nil {
		return err
	}

	return nil
}
