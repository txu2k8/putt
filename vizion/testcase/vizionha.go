package testcase

import (
	"pzatest/libs/utils"
	"pzatest/types"
	"pzatest/vizion/resources"
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
}

func (job *HABase) diffPods() {
	logger.Info("HABase: diffPods")
}

func (job *HABase) surviveS3Upload() {
	logger.Info("HABase: surviveS3Upload")
	job.SurviveS3Files = append(job.SurviveS3Files, UploadFile{FileName: "testfile-1.txt"})
}

func (job *HABase) surviveS3Download() {
	logger.Info("HABase: surviveS3Download")
}

func (job *HABase) surviveESIndex() {
	logger.Info("HABase: surviveESIndex")
}

func (job *HABase) surviveBDWrite() {
	logger.Info("HABase: surviveBDWrite")
}

func (job *HABase) surviveStress() {
	logger.Info("HABase: surviveStress")
}

// HAWorkflow ...
func (job *HABase) HAWorkflow(fn func() error) error {
	logger.Info("Eneter HAWorkflow ...")
	logger.Info(utils.Prettify(&job))
	logger.Info("STEP1: Check health before HA test...")
	logger.Info("STEP2: S3 upload/download before HA test ...")
	logger.Info("STEP3: ES index before HA test ...")
	logger.Info("STEP4: BlockDevice Write/Read before HA test ...")

	logger.Info("STEP5: HA and Survive opt ...")
	job.diffPods()
	err := fn()
	if err != nil {
		return err
	}
	logger.Info("STEP6: Check health after HA test ...")
	job.diffPods()
	logger.Info("STEP7: Download and Check S3 original data(upload by STEP2) MD5 after HA test ...")
	logger.Info("STEP8: Read and Check block device data(write by STEP4) MD5 after HA test ...")
	logger.Info("STEP9: Check && Make sure es OK after HA test ...")

	return nil
}

// ============================= RestartNode =============================

// RestartNodeTestInput ...
type RestartNodeTestInput struct {
	// TestInput
	NodeIPs    []string // To restart node IP address Array
	VMNames    []string // To restart node VM name Array
	Platform   string   // Test VM platfor: vsphere | aws
	PowerOpts  []string // Power opts: shoutdwon|poweroff|reset|reboot
	RestartNum int      // Restart VM number

	// job
	HABase
	TestNodes       []RestartNodeInfo
	RandomTestNodes []RestartNodeInfo
}

// RestartNodeInfo ...
type RestartNodeInfo struct {
	VMName   string
	VMIP     string
	Role     string
	PowerOpt string
}

func (job *RestartNodeTestInput) parseRestartNodes() {
	logger.Info("Run parseRestartNodes")
	job.TestNodes = append(job.TestNodes, RestartNodeInfo{VMName: "VMName-1"})
	logger.Info(utils.Prettify(&job))
}

func (job *RestartNodeTestInput) powerDownNodes() {
	logger.Info("Run powerDownNodes")
	job.PhaseList = append(job.PhaseList, "powerDownNodes")
}

func (job *RestartNodeTestInput) powerOnNodes() {
	logger.Info("Run powerOnNodes")
	job.PhaseList = append(job.PhaseList, "powerOnNodes")
}

func (job *RestartNodeTestInput) isSurviveOK() {
	logger.Info("Run isSurviveOK")
	job.S3OK = true
	job.ESOK = false
}

func (job *RestartNodeTestInput) runJob() error {
	logger.Info("Run RestartNodeJob")
	job.parseRestartNodes()
	job.powerDownNodes()
	job.isSurviveOK()
	if job.S3OK {
		job.surviveS3Upload()
	}
	if job.ESOK {
		job.surviveESIndex()
	}
	if job.DagentOK {
		job.surviveBDWrite()
	}
	job.powerOnNodes()
	return nil
}

// Run RestartNodeJob With HAWorkflow
func (job *RestartNodeTestInput) Run() error {
	logger.Info("RestartNodeJob: Run With HAWorkflow")
	job.HAWorkflow(job.runJob)
	logger.Info(utils.Prettify(&job))
	return nil
}

// ============================= RestartService =============================

// Debug code
func Debug(conf types.VizionBaseInput) error {
	host := "10.25.119.77"
	vizion := resources.VizionBase{VizionBaseInput: conf}
	exist := vizion.Node(host).IsDplmodExist()
	logger.Info(exist)
	vizion.Service().K8sEnableNodeLabel("", "", "")

	return nil
}
