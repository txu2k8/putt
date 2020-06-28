package testcase

import (
	"fmt"
	"pzatest/config"
	"pzatest/libs/convert"
	"pzatest/libs/random"
	"pzatest/libs/runner/schedule"
	"pzatest/libs/utils"
	"pzatest/types"
	"pzatest/vizion/resources"
	"strings"
)

// HATester ...
type HATester interface {
	Run() error
}

// HABase ...
type HABase struct {
	PhaseList      []string
	S3OK           bool         // S3 service status
	DagentOK       bool         // dagent service status
	ESOK           bool         // ES service status (DagentOK && ESOK)
	NFSOK          bool         // NFS service status
	SurviveS3Files []UploadFile // S3 upload files info after powerdown node/service/ch ...
	SurviveBDFiles []UploadFile // BD write files info after powerdown node/service/ch ...
	Schedule       schedule.Schedule
}

func action() error { return nil }

func (ha *HABase) diffPods() error {
	logger.Info("HABase: diffPods")
	return nil
}

func (ha *HABase) surviveS3Upload() error {
	logger.Info("HABase: surviveS3Upload")
	ha.SurviveS3Files = append(ha.SurviveS3Files, UploadFile{FileName: "testfile-1.txt"})
	return nil
}

func (ha *HABase) surviveS3Download() error {
	logger.Info("HABase: surviveS3Download")
	return nil
}

func (ha *HABase) surviveESIndex() error {
	logger.Info("HABase: surviveESIndex")
	return nil
}

func (ha *HABase) surviveBDWrite() error {
	logger.Info("HABase: surviveBDWrite")
	return nil
}

func (ha *HABase) surviveStress() error {
	logger.Info("HABase: surviveStress")
	return nil
}

// HAWorkflow ...
func (ha *HABase) HAWorkflow(fn func() error) error {
	// logger.Info("Eneter HAWorkflow ...")
	ha.Schedule.RunPhase(action, schedule.Desc("Eneter HAWorkflow ..."))
	// logger.Info(utils.Prettify(&ha))
	// logger.Info("STEP1: Check health before HA test...")
	// logger.Info("STEP2: S3 upload/download before HA test ...")
	// logger.Info("STEP3: ES index before HA test ...")
	// logger.Info("STEP4: BlockDevice Write/Read before HA test ...")
	ha.Schedule.RunPhase(action, schedule.Desc("Check health before HA test ..."))
	ha.Schedule.RunPhase(action, schedule.Desc("S3 upload/download before HA test ..."))
	ha.Schedule.RunPhase(action, schedule.Desc("ES index before HA test ..."))
	ha.Schedule.RunPhase(action, schedule.Desc("BlockDevice Write/Read before HA test ..."))

	// logger.Info("STEP5: HA and Survive opt ...")
	ha.Schedule.RunPhase(action, schedule.Desc("Eneter HA and Survive opt ..."))
	ha.Schedule.RunPhase(ha.diffPods, schedule.Desc("diffPods ..."))
	err := fn()
	if err != nil {
		return err
	}
	// logger.Info("STEP6: Check health after HA test ...")
	ha.Schedule.RunPhase(action, schedule.Desc("Check health after HA test ..."))
	ha.Schedule.RunPhase(ha.diffPods, schedule.Desc("diffPods ..."))
	// logger.Info("STEP7: Download and Check S3 original data(upload by STEP2) MD5 after HA test ...")
	// logger.Info("STEP8: Read and Check block device data(write by STEP4) MD5 after HA test ...")
	// logger.Info("STEP9: Check && Make sure es OK after HA test ...")
	ha.Schedule.RunPhase(action, schedule.Desc("Download and Check S3 original data(upload by STEP2) MD5 after HA test ..."))
	ha.Schedule.RunPhase(action, schedule.Desc("Read and Check block device data(write by STEP4) MD5 after HA test ..."))
	ha.Schedule.RunPhase(action, schedule.Desc("Check && Make sure es OK after HA test ..."))

	return nil
}

// ============================= RestartNode =============================

// RestartNodeTestInput ...
type RestartNodeTestInput struct {
	NodeIPs       []string      // To restart node IP address Array
	VMNames       []string      // To restart node VM name Array
	Platform      string        // Test VM platfor: vsphere | aws
	PowerOpts     []string      // Power opts: shoutdwon|poweroff|reset|reboot
	RestartNumMin int           // Min Restart VM number
	RestartNumMax int           // Max Restart VM number
	Vsphere       types.Vsphere // Vsphere
}

// RestartNodeInfo ...
type RestartNodeInfo struct {
	VMName   string
	VMIP     string
	Role     string
	PowerOpt string
}

// HANode .
type HANode struct {
	HABase
	RestartNodeTestInput
	TestNodes       []RestartNodeInfo
	RandomTestNodes []RestartNodeInfo
}

// NewHANode ...
func NewHANode(input RestartNodeTestInput) *HANode {
	return &HANode{
		RestartNodeTestInput: input,
	}
}

func (job *HANode) parseRestartNodes() error {
	powerOpt := random.ChoiceStrArr(job.PowerOpts)
	logger.Info(">> Parse Test Nodes by input args ...")
	inputNodes := []string{}
	switch {
	case len(job.NodeIPs) > 0:
		inputNodes = job.NodeIPs
	case len(job.VMNames) > 0:
		inputNodes = job.VMNames
	default:
		return fmt.Errorf("Please input NodeIPs or VMNames")
	}

	testNodes := []string{}
	if strings.Contains(strings.Join(inputNodes, ""), ",") {
		for _, inputNode := range inputNodes {
			nodeSplit := strings.Split(inputNode, ":")
			rNum := 0
			if len(nodeSplit) >= 2 {
				rNums := convert.StrNumToIntArr(nodeSplit[1], ",", 2)
				rMin := rNums[0]
				rMax := rNums[1]
				rNum = random.RandRangeInt(rMin, rMax)
			} else {
				rNum = random.RandRangeInt(job.RestartNumMin, job.RestartNumMax)
			}
			tmpNodes := strings.Split(nodeSplit[0], ",")
			testNodes = append(testNodes, random.SampleStrArr(tmpNodes, rNum)...)
		}
	} else {
		rNum := random.RandRangeInt(job.RestartNumMin, job.RestartNumMax)
		testNodes = random.SampleStrArr(inputNodes, rNum)
	}
	testNodes = utils.UniqArr(testNodes)
	if job.Platform == "vsphere" {
		// TODO
		for _, node := range testNodes {
			vmName, vmIP := "", ""
			if job.NodeIPs != nil {
				vmIP = node
				vmName = "" // TODO
			} else {
				vmIP = node
				vmName = "" // TODO
			}
			nodeInfo := RestartNodeInfo{
				VMName:   vmName,
				VMIP:     vmIP,
				PowerOpt: powerOpt,
			}
			job.TestNodes = append(job.TestNodes, nodeInfo)
		}
	}

	job.TestNodes = append(job.TestNodes, RestartNodeInfo{VMName: "VMName-1"})
	logger.Info(utils.Prettify(job.TestNodes))
	// logger.Info(utils.Prettify(&job))
	return fmt.Errorf("stop")
}

func (job *HANode) powerDownNodes() error {
	logger.Info("Run powerDownNodes")
	job.PhaseList = append(job.PhaseList, "powerDownNodes")
	return nil
}

func (job *HANode) powerOnNodes() error {
	logger.Info("Run powerOnNodes")
	job.PhaseList = append(job.PhaseList, "powerOnNodes")
	return nil
}

func (job *HANode) isSurviveOK() error {
	logger.Info("Run isSurviveOK")
	job.S3OK = true
	job.ESOK = false
	return nil
}

func (job *HANode) runJob() error {
	var err error
	logger.Info("Run RestartNodeJob")
	if err = job.Schedule.RunPhase(job.parseRestartNodes, schedule.Desc("Parse Restart Nodes Input ...")); err != nil {
		return err
	}
	if err = job.Schedule.RunPhase(job.powerDownNodes, schedule.Desc("Powerdown nodes ...")); err != nil {
		return err
	}
	if err = job.Schedule.RunPhase(job.isSurviveOK, schedule.Desc("Judge is Survived s3/bd/es/nfs shoudld OK ...")); err != nil {
		return err
	}
	if job.S3OK {
		if err = job.Schedule.RunPhase(job.surviveS3Upload, schedule.Desc("Survived S3Upload ...")); err != nil {
			return err
		}
	}
	if job.ESOK {
		if err = job.Schedule.RunPhase(job.surviveESIndex, schedule.Desc("Survived ESIndex ...")); err != nil {
			return err
		}
	}
	if job.DagentOK {
		if err = job.Schedule.RunPhase(job.surviveBDWrite, schedule.Desc("Survived BDWrite ...")); err != nil {
			return err
		}
	}

	if err = job.Schedule.RunPhase(job.powerOnNodes, schedule.Desc("PowerOn nodes ...")); err != nil {
		return err
	}

	return nil
}

// Run RestartNodeJob With HAWorkflow
func (job *HANode) Run() error {
	logger.Info("RestartNodeJob: Run With HAWorkflow")
	err := job.HAWorkflow(job.runJob)
	// logger.Info(utils.Prettify(&job))
	return err
}

// ============================= RestartService =============================

// Debug code
func Debug(base types.VizionBaseInput) error {
	// host := "10.25.119.77"
	vizion := resources.Vizion{Base: base}

	// masterCass := vizion.Cass().SetIndex("0")
	// svArr, _ := masterCass.GetServiceByType(1024)
	// logger.Info(utils.Prettify(svArr))
	// nArr, _ := masterCass.GetServiceByType(33)
	// logger.Info(utils.Prettify(nArr))

	// subCass := vizion.Cass().SetIndex("1")
	// vArr, _ := subCass.GetVolume()
	// logger.Info(utils.Prettify(vArr))

	logPathArr := []string{}
	for _, sv := range config.DefaultServiceArray {
		logArr := sv.GetLogDirArr(base)
		// logger.Info(utils.Prettify(logArr))
		logPathArr = append(logPathArr, logArr...)
	}
	for _, nodeIP := range vizion.Service().GetAllNodeIPs() {
		node := vizion.Node(nodeIP)
		node.CleanLog(logPathArr)
	}

	// return vizion.Check().IsNodeCrashed()
	return nil
}
