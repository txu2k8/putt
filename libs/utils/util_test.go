package utils

import (
	_ "gtest/testinit"
	"testing"
)

func TestUtilFunc(t *testing.T) {
	logger.Info(SizeToName(12 * 1024 * 1024 * 1024))
	// 2020-04-01T13:46:29 test INFO: 12GB
}
