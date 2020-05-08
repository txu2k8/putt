package stress

import (
	"errors"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"testing"
	"time"
)

func aJob() error {
	time.Sleep(10 * time.Millisecond)
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}

func JobB() error {
	time.Sleep(100 * time.Millisecond)
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}

// Run a job for 42 seconds in 100 go routines
func Test1(t *testing.T) {
	results := make(map[error]int)

	jobs := []Job{
		{
			// the actual job which is a `func() error`
			Fn: aJob,
			// job name (optional)
			Name: "A job",
		},
	}

	// aggregation function
	aggregate := func(r Result) {
		if r.JobNr%1000 == 0 {
			fmt.Println("job nr:", r.JobNr)
		}
		results[r.Error]++
	}

	maxParallel := 100

	runner := New(maxParallel, jobs, aggregate)

	// stop after 42 seconds
	time.AfterFunc(42*time.Second, runner.Stop)

	// Start runner
	// this is blocking
	runner.Start()

	fmt.Println(results)
}

// Run a job for 1000 times in 100 go routines
func Test2(t *testing.T) {
	results := make(map[error]int)
	var totalTook time.Duration

	jobs := []Job{
		{
			Fn:   aJob,
			Name: "A job",
			// limit of jobs to run
			RunTimes: 1000,
		},
	}

	aggregate := func(r Result) {
		if r.JobNr%1000 == 0 {
			fmt.Println("job nr:", r.JobNr)
		}
		totalTook += r.End.Sub(r.Start)
		results[r.Error]++
	}

	runner := New(100, jobs, aggregate)

	runner.Start()

	fmt.Println("results:", results)
	fmt.Println("took", totalTook)
	fmt.Println("average time in ms:", float64(totalTook)/float64(1000*time.Millisecond))
}

// Run a job for 100 times and job b for 10 seconds in 100 go routines each
func Test3(t *testing.T) {
	resultsA := make(map[error]int)
	resultsB := make(map[error]int)

	results := map[string]map[error]int{
		"A job": resultsA,
		"job B": resultsB,
	}

	jobs := []Job{
		{
			Fn:       aJob,
			Name:     "A job",
			RunTimes: 100,
		},
		{
			Fn:   JobB,
			Name: "job B",
		},
	}

	aggregate := func(r Result) {
		if r.JobNr%1000 == 0 {
			fmt.Printf("job %s nr: %d\n", r.Name, r.JobNr)
		}
		results[r.Name][r.Error]++
	}

	runner := New(100, jobs, aggregate)

	stop := make(chan os.Signal)
	signal.Notify(stop, os.Interrupt)

	go func() {
		// stop after 10 seconds
		// or by SIGINT
		select {
		case <-time.After(10 * time.Second):
		case <-stop:
		}
		runner.Stop()
	}()

	runner.Start()

	fmt.Println("results:", results)
}
