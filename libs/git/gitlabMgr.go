package git

import (
	"errors"
	"fmt"
	"log"
	"pzatest/libs/retry"
	"pzatest/libs/retry/backoff"
	"pzatest/libs/retry/strategy"
	"pzatest/libs/utils"
	"time"

	"github.com/xanzy/go-gitlab"
)

// GitlabManager .
type GitlabManager interface {
	GetProjectPipelineArr(projectID int, ref string) ([]*gitlab.PipelineInfo, error)
	GetPipelineJobArr(projectID, pipelineID int) ([]*gitlab.Job, error)
	IsJobStatusExpected(projectID int, tagName, jobName, expectStatus string) error
	WaitJobStatusExpected(projectID int, tagName, jobName, expectStatus string) error
	IsPipelineJobsSuccess(projectID int, tagName string) error
}

// GitlabMgr .
type GitlabMgr struct {
	Client *gitlab.Client // Gitlab client
	Cfg    *GitlabConfig  // Gitlab config
}

// GitlabConfig .
type GitlabConfig struct {
	BaseURL string
	Token   string
}

// NewGitlabClient returns a gitlab client
func NewGitlabClient(c GitlabConfig) *GitlabMgr {
	client, err := gitlab.NewClient(c.Token, gitlab.WithBaseURL(c.BaseURL))
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	return &GitlabMgr{
		Client: client,
		Cfg:    &c,
	}
}

// GetProjectPipelineArr ...
func (g *GitlabMgr) GetProjectPipelineArr(projectID int, ref string) ([]*gitlab.PipelineInfo, error) {
	opt := &gitlab.ListProjectPipelinesOptions{
		// Scope:         gitlab.String("branches"),
		// Status:        gitlab.BuildState(gitlab.Running),
		// Ref:           gitlab.String("master"),
		// YamlErrors:    gitlab.Bool(true),
		// Name:          gitlab.String("name"),
		// Username:      gitlab.String("username"),
		Ref:           gitlab.String(ref),
		UpdatedAfter:  gitlab.Time(time.Now().Add(-24 * 365 * time.Hour)),
		UpdatedBefore: gitlab.Time(time.Now().Add(-7 * 24 * time.Hour)),
		OrderBy:       gitlab.String("status"),
		Sort:          gitlab.String("asc"),
	}

	pipelines, resp, err := g.Client.Pipelines.ListProjectPipelines(projectID, opt)
	if err != nil {
		logger.Errorf(utils.Prettify(resp))
		logger.Fatal(err)
	}

	return pipelines, err
}

// GetPipelineJobArr ...
func (g *GitlabMgr) GetPipelineJobArr(projectID, pipelineID int) ([]*gitlab.Job, error) {
	jobs, resp, err := g.Client.Jobs.ListPipelineJobs(projectID, pipelineID, nil)
	if err != nil {
		logger.Errorf(utils.Prettify(resp))
		logger.Fatal(err)
	}

	return jobs, err
}

// IsJobStatusExpected .
func (g *GitlabMgr) IsJobStatusExpected(projectID int, tagName, jobName, expectStatus string) error {
	pipelines, _ := g.GetProjectPipelineArr(projectID, tagName)
	// logger.Infof("Pipline Array:\n%s", utils.Prettify(pipelines))
	if len(pipelines) <= 0 {
		return fmt.Errorf("Got None pipelines with tag: %s", tagName)
	}

	for _, pipeline := range pipelines {
		jobs, _ := g.GetPipelineJobArr(projectID, pipeline.ID)
		// logger.Info(utils.Prettify(jobs))
		if len(jobs) <= 0 {
			return fmt.Errorf("Got None job in pipeline %d", pipeline.ID)
		}

		for _, job := range jobs {
			if job.Name == jobName {
				switch job.Status {
				case expectStatus:
					logger.Infof("Job:%12s, Status:%s, expect:%s", job.Name, job.Status, expectStatus)
					return nil
				case "canceled":
					if job.Name == "test" {
						logger.Infof("Job:%12s, Status:%s, expect:%s", job.Name, job.Status, expectStatus)
						return nil
					} else if job.Name == "build-image" {
						panic("Job build-image CANCELED")
					}
				}
				logger.Warningf("Job:%12s, Status:%s, expect:%s", job.Name, job.Status, expectStatus)
			}
		}
	}
	return errors.New("Jobs status not expected")
}

// WaitJobStatusExpected ...
func (g *GitlabMgr) WaitJobStatusExpected(projectID int, tagName, jobName, expectStatus string) error {
	action := func(attempt uint) error {
		return g.IsJobStatusExpected(projectID, tagName, jobName, expectStatus)
	}
	err := retry.Retry(
		action,
		strategy.Limit(360),
		strategy.Backoff(backoff.Fibonacci(20*time.Second)),
	)
	return err
}

// IsPipelineJobsSuccess job1: build-image, job2: test
func (g *GitlabMgr) IsPipelineJobsSuccess(projectID int, tagName string) error {
	if err := g.WaitJobStatusExpected(projectID, tagName, "build-image", "success"); err == nil {
		if err := g.WaitJobStatusExpected(projectID, tagName, "test", "success"); err == nil {
			return nil
		}
		return errors.New("build-imag:OK, test:FAIL")
	}
	return errors.New("Pipeline Jobs not all success, Image not Availabel")
}
