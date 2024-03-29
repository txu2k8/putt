package git

import (
	"fmt"
	"path"
	"platform/libs/sshmgr"
	"strings"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// ManagerGetter has a method to return a NodeInterface.
// A group's client should implement this interface.
type ManagerGetter interface {
	Node(host string) SSHGitManagerInterface
}

// SSHGitManagerInterface has methods to work on GitManager resources.
type SSHGitManagerInterface interface {
	sshmgr.SSHManager
	Pull(projectPath string) error
	Tag(projectPath, tagName string) error
	GetCurrentBranch(projectPath, tagName string) string
	MakeFile(binPath, binName string) string
}

// SSHGitMgr implements ManagerInterface
type SSHGitMgr struct {
	*sshmgr.SSHMgr
}

// NewGitMgr returns a Nodes
func NewGitMgr(host, username, password, keyFile string) *SSHGitMgr {
	sshKey := sshmgr.SSHKey{
		UserName: username,
		Password: password,
		Port:     22,
		KeyFile:  keyFile,
	}
	return &SSHGitMgr{sshmgr.NewSSHMgr(host, sshKey)}
}

// Pull ...
func (g *SSHGitMgr) Pull(projectPath string) error {
	branchCmd := fmt.Sprintf("cd %s && git branch", projectPath)
	rc, output := g.RunCmd(branchCmd)
	logger.Infof("%d, %s", rc, output)

	pullCmd := fmt.Sprintf("cd %s && git pull", projectPath)
	rc, output = g.RunCmd(pullCmd)
	logger.Infof("%d, %s", rc, output)
	if strings.Contains(output, "error") || strings.Contains(output, "fatal") {
		return fmt.Errorf("Git pull failed")
	}
	return nil
}

// Tag ...
func (g *SSHGitMgr) Tag(projectPath, tagName string) error {
	tagCmd := fmt.Sprintf("cd %s && git tag -a %s -m \"tag for test build\"", projectPath, tagName)
	rc, output := g.RunCmd(tagCmd)
	logger.Infof("%d, %s", rc, output)

	pushTagCmd := fmt.Sprintf("cd %s && git push origin %s", projectPath, tagName)
	rc, output = g.RunCmd(pushTagCmd)
	logger.Infof("%d, %s", rc, output)

	return nil
}

// GetCurrentBranch ...
func (g *SSHGitMgr) GetCurrentBranch(projectPath string) string {
	cmdSpec := fmt.Sprintf("cd %s && git rev-parse --abbrev-ref HEAD", projectPath)
	rc, output := g.RunCmd(cmdSpec)
	logger.Infof("%d, %s", rc, output)

	branch := strings.TrimRight(output, "\n")
	return branch
}

// GetChangeLog ...
func (g *SSHGitMgr) GetChangeLog(projectPath string) (date, changeLog string) {
	rc, output := g.RunCmd("date")
	logger.Infof("%d, %s", rc, output)
	date = strings.Trim(output, "\n")

	getLatestTag := fmt.Sprintf("cd %s && git describe --tags", projectPath)
	rc, output = g.RunCmd(getLatestTag)
	logger.Infof("%d, %s", rc, output)
	latestTag := strings.Trim(output, "\n")

	cmdSpec := fmt.Sprintf("cd %s && git log --pretty=oneline  ORIG_HEAD..%s", projectPath, latestTag)
	rc, output = g.RunCmd(cmdSpec)
	logger.Infof("%d, %s", rc, output)
	changeLog = strings.Trim(output, "\n")
	return
}

// MakeFile ...
func (g *SSHGitMgr) MakeFile(binPath, binName string) string {
	// make realclean
	realcleanCmd := fmt.Sprintf("cd %s && make realclean", binPath)
	rc, output := g.RunCmd(realcleanCmd)
	logger.Infof("%d, %s", rc, output)

	// grep binary, make sure "make realclean" success
	grepCmd := fmt.Sprintf("ls %s | grep -w ^%s$", binPath, binName)
	rc, output = g.RunCmd(grepCmd)
	logger.Infof("%d, %s", rc, output)

	// make new binary
	makeCmd := fmt.Sprintf("cd %s && make -j8", binPath)
	_, output = g.RunCmd(makeCmd)
	if strings.Contains(output, "Error") || strings.Contains(output, "error") {
		logger.Warning(output)
	} else {
		logger.Debug(output)
	}

	// get new binary MD5
	md5Cmd := "md5sum " + path.Join(binPath, binName)
	rc, output = g.RunCmd(md5Cmd)
	logger.Infof("%d, %s, %s", rc, binName, md5Cmd)
	md5sum := strings.Split(strings.TrimRight(output, "\n"), " ")[0]
	return md5sum
}
