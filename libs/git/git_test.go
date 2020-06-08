package git

import (
	"testing"
)

func TestGtlab(t *testing.T) {
	config := GitlabConfig{
		BaseURL: "http://gitlab.panzura.com",
		Token:   "xjB1FHHyJHNQUhgy7K11",
	}

	// newGitlabClient(config)
	gitMgr := NewGitlabClient(config)
	err := gitMgr.IsImageOK(25, "develop.dpl.6.0.0.4535")
	logger.Info(err)
}
