package utils

import (
	_ "gtest/testinit"
	"testing"
)

func TestUtilFunc(t *testing.T) {
	logger.Info(SizeCountByte(`12.2MB`))
	// 2020-04-08T17:26:53 test INFO: 12792627

	logger.Info(ByteCountDecimal(12792627))
	// 2020-04-08T16:36:08 test INFO: 12.2 MB
}
