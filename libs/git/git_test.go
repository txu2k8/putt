package git

import (
	"testing"
)

func TestGitMgr(t *testing.T) {
	gitMgr := NewGitMgr("10.199.116.1", "root", "password", "")
	b := gitMgr.GetCurrentBranch("/home/project/dpl/develop/dpl")
	logger.Info(b)
	err := gitMgr.Pull("/home/project/dpl/develop/dpl")
	logger.Info(err)
}

func TestGitlab(t *testing.T) {
	config := GitlabConfig{
		BaseURL: "http://gitlab.panzura.com",
		Token:   "xjB1FHHyJHNQUhgy7K11",
	}
	gitMgr := NewGitlabClient(config)
	err := gitMgr.IsPipelineJobsSuccess(25, "develop.dpl.6.0.0.4535")
	logger.Info(err)
}
