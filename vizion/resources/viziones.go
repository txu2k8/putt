package resources

import (
	"errors"
	"pzatest/libs/utils"
	"math/rand"
	"time"
)

// ESIndex ...
func ESIndex() error {
	logger.Info("ES Index test start ...")
	logger.Info("ES Index test complete ...")
	if rand.Intn(1) == 0 {
		return errors.New("job error")
	}
	utils.SleepProgressBar(10 * time.Second)
	return nil
}

// ESSearch ...
func ESSearch() error {
	logger.Info("ES Search test start ...")
	logger.Info("ES Search test complete ...")
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}

// ESCleanup ...
func ESCleanup() error {
	logger.Info("ES Cleanup test start ...")
	logger.Info("ES Cleanup test complete ...")
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}
