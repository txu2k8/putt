package stress

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

// Job struct
type Job struct {
	Fn          func() error
	Name        string
	RunTimes    int
	MaxParallel int
}

// Result struct
type Result struct {
	Error   error
	Name    string
	JobNr   int
	Start   time.Time
	End     time.Time
	Summary string
}

type runner struct {
	wg      sync.WaitGroup
	done    chan struct{}
	results chan Result

	logWriter io.Writer

	maxParallel int
	jobs        []Job
	aggregate   func(Result)
}

// New job
func New(maxParallel int, jobs []Job, aggregate func(Result)) *runner {
	return &runner{
		done:        make(chan struct{}),
		results:     make(chan Result, maxParallel),
		maxParallel: maxParallel,
		jobs:        jobs,
		aggregate:   aggregate,
		logWriter:   os.Stdout,
	}
}

func (r *runner) SetLogWriter(w io.Writer) {
	r.logWriter = w
}

func (r *runner) Stop() {
	close(r.done)
}

func (r *runner) Start() {
	for _, job := range r.jobs {
		r.wg.Add(1)
		go func(job Job) {
			r.runJob(job)
			r.wg.Done()
		}(job)
	}
	go func() {
		r.wg.Wait()
		logger.Info("All jobs finished\n")
		close(r.results)
	}()
	r.listenResults()
}

func (r *runner) listenResults() {
	for res := range r.results {
		r.aggregate(res)
	}
}

func (r *runner) runJob(job Job) {
	maxParallel := job.MaxParallel
	if maxParallel <= 0 {
		maxParallel = r.maxParallel
	}
	ch := make(chan struct{}, maxParallel)
	wg := sync.WaitGroup{}
	forever := job.RunTimes <= 0
loop:
	for i := 1; forever || job.RunTimes-i >= 0; i++ {
		logger.Warning(i, job.RunTimes)
		time.After(30 * time.Second)
		select {
		case ch <- struct{}{}:
			wg.Add(1)
			go func(i int) {
				start := time.Now()
				logger.Infof("[START ] - %s - Iteration: %d", job.Name, i)
				err := job.Fn()
				if err != nil {
					logger.Error(err)
					// r.Stop()
					r.done <- struct{}{}
					job.RunTimes = i
				}
				end := time.Now()
				<-ch
				r.results <- Result{Name: job.Name, Error: err, Start: start, End: end, JobNr: i}
				wg.Done()
			}(i)
		case <-r.done:
			logger.Warningf("Received done in '%s' job, waiting to finish\n", job.Name)
			break loop
		}
	}
	wg.Wait()
	logger.Infof("Job(Name:'%s', RunTimes:%d) is done\n", job.Name, job.RunTimes)
	return
}

// Run Stress Test
func Run(jobs []Job) {
	var result string // PASS | FAIL | ERROR
	var iterationTook, totalTook time.Duration
	// results := make(map[error]int)
	results := []Result{}

	aggregate := func(r Result) {
		switch r.Error {
		case nil:
			result = "[ PASS ]"
		default:
			result = "[ FAIL ]"
		}
		iterationTook = r.End.Sub(r.Start)
		totalTook += iterationTook
		// results[r.Error]++
		r.Summary = fmt.Sprintf("%s - %s - Iteration: %d - Elapsed Time: %s", result, r.Name, r.JobNr, iterationTook)
		results = append(results, r)
		logger.Infof(r.Summary)
	}

	runner := New(1, jobs, aggregate)
	runner.Start()

	logger.Info(strings.Repeat("=", 50))
	for _, res := range results {
		logger.Info(res.Summary)
	}
	logger.Info("Total:", len(results))
	logger.Info("Time Elapsed:", totalTook)
	logger.Info(strings.Repeat("=", 50))
}
