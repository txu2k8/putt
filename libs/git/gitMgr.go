package git

import (
	"fmt"
	"path"
	"pzatest/libs/sshmgr"
	"strings"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// ManagerGetter has a method to return a NodeInterface.
// A group's client should implement this interface.
type ManagerGetter interface {
	Node(host string) ManagerInterface
}

// ManagerInterface has methods to work on GitManager resources.
type ManagerInterface interface {
	sshmgr.SSHManager
	GetKubeConfig(localPath string) error
	GetKubeVipIP(fqdn string) (vIP string)
	GetCrashDirs() (crashArr []string)
}

// nodes implements NodeInterface
type gitMgr struct {
	*sshmgr.SSHMgr
}

// newGitMgr returns a Nodes
func newGitMgr(host, username, password, keyFile string) *gitMgr {
	sshKey := sshmgr.SSHKey{
		UserName: username,
		Password: password,
		Port:     22,
		KeyFile:  keyFile,
	}
	return &gitMgr{sshmgr.NewSSHMgr(host, sshKey)}
}

// Pull ...
func (g *gitMgr) Pull(projectPath string) error {
	branchCmd := fmt.Sprintf("cd %s && git branch -a", projectPath)
	_, output := g.RunCmd(branchCmd)
	logger.Info(output)

	pullCmd := fmt.Sprintf("cd %s && git pull", projectPath)
	_, output = g.RunCmd(pullCmd)
	logger.Info(output)
	if strings.Contains(output, "error") || strings.Contains(output, "fatal") {
		return fmt.Errorf("Git pull failed")
	}
	return nil
}

// tag
func (g *gitMgr) Tag(projectPath, tagName string) error {
	tagCmd := fmt.Sprintf("cd %s && git tag -a %s 0m \"tag for test build\"", projectPath, tagName)
	_, output := g.RunCmd(tagCmd)
	logger.Info(output)

	pushTagCmd := fmt.Sprintf("cd %s && git push origin %s", projectPath, tagName)
	_, output = g.RunCmd(pushTagCmd)
	logger.Info(output)

	return nil
}

// branch
func (g *gitMgr) GetCurrentBranch(projectPath, tagName string) string {
	cmdSpec := fmt.Sprintf("cd %s && git rev-parse --abbrev-ref HEAD", projectPath)
	_, output := g.RunCmd(cmdSpec)
	logger.Info(output)

	branch := strings.TrimRight(output, "\n")
	return branch
}

// make files
func (g *gitMgr) MakeFile(binPath, binName string) string {
	// make realclean
	realcleanCmd := fmt.Sprintf("cd %s && make realclean", binPath)
	_, output := g.RunCmd(realcleanCmd)
	logger.Info(output)

	// grep binary, make sure "make realclean" success
	grepCmd := fmt.Sprintf("ls %s | grep -w ^%s$", binPath, binName)
	_, output = g.RunCmd(grepCmd)
	logger.Info(output)

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
	_, output = g.RunCmd(md5Cmd)
	md5sum := strings.Split(strings.TrimRight(output, "\n"), " ")[0]
	return md5sum
}
