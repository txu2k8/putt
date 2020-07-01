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
		BaseURL: "http://gitlab.xx.com",
		Token:   "xjB1FHHyJHNQUhgy7K7t11",
	}
	gitMgr := NewGitlabClient(config)
	// descV110 := "v1.1.0:\n" +
	// 	"StressTest | Develop | Maintenance | Tools\n" +
	// 	"1. Support for vizion old_setup v1.1.0(setup with python raw_input)\n" +
	// 	"2. old_setup services depends on master-cassandra\n" +
	// 	"3. old_setup ServiceManager with /opt/xxx python scripts\n"

	descV22 := "v2.2:\n" +
		"StressTest | Develop | Maintenance | Tools\n" +
		"1. Support for vizion new setup(JIRA VZ-8990)\n" +
		"2. Support for vizion dpl v2.2.x:\n" +
		"3. No master cassandra, scripts get services depends on k8s\n" +
		"4. Service VPM removed\n" +
		"5. Service anchor -> mjcachedpl + djacahedpl (with group id)\n" +
		"6. Service cmapdpl -> cmapmcdpl + cmapdcdpl (with group id)\n" +
		"7. Others ..."

	_, err := gitMgr.UpdateRelease(47, "v2.2", "v2.2", descV22)
	logger.Info(err)
}
