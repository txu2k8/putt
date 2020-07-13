package testcase

import (
	"errors"
	"fmt"
	"math/rand"
	"platform/libs/utils"
	"time"
)

// ESTester ...
type ESTester interface {
	ESIndex() error
	ESSearch() error
	ESCleanup() error
	ESStress() error
}

// ESTestInput ...
type ESTestInput struct {
	IP              string
	UserName        string
	Password        string
	Port            int
	URL             string // Parse(IP,Port) --> URL
	IndexNamePrefix string // index name prefix
	Indices         int    // Number of indices to write
	Documents       int    // Number of template documents that hold the same mapping
	BulkSize        int    // How many documents each bulk request should contain
	Workers         int    // Number of workers.
}

// ParseESInput ...
func (conf *ESTestInput) ParseESInput() {
	// Parse ES Ip Port to conf.URL
	conf.URL = fmt.Sprintf("http://%s:%d", conf.IP, conf.Port)
	logger.Debugf("ESTestInput:%v", utils.Prettify(conf))
}

// ESIndex ...
func (conf *ESTestInput) ESIndex() error {
	conf.ParseESInput()
	logger.Info("ES Index test start ...")
	logger.Info("ES Index test complete ...")
	if rand.Intn(1) == 0 {
		return errors.New("job error")
	}
	utils.SleepProgressBar(10 * time.Second)
	return nil
}

// ESSearch ...
func (conf *ESTestInput) ESSearch() error {
	logger.Info("ES Search test start ...")
	logger.Info("ES Search test complete ...")
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}

// ESCleanup ...
func (conf *ESTestInput) ESCleanup() error {
	logger.Info("ES Cleanup test start ...")
	logger.Info("ES Cleanup test complete ...")
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}

// ESStress : Index && Search
func (conf *ESTestInput) ESStress() error {
	logger.Info("ES Stress(Index && Search) test start ...")
	logger.Info("ES Stress:(Index && Search) test complete ...")
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}
