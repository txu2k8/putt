package utils

import (
	_ "gtest/testinit"
	"testing"
)

func TestUtilFunc(t *testing.T) {
	logger.Info(SizeCountByte(`12 GB`))
	// 2020-04-08T17:26:53 test INFO: 12884901888

	logger.Info(ByteCountDecimal(12884901888))
	// 2020-04-08T16:36:08 test INFO: 11.7 GB
}
