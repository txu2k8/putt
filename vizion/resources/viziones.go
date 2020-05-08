package resources

import (
	"errors"
	"math/rand"
)

// ESIndex ...
func ESIndex() error {
	logger.Info("ES Index test start ...")
	logger.Info("ES Index test complete ...")
	if rand.Intn(2) == 0 {
		return errors.New("job error")
	}
	return nil
}
