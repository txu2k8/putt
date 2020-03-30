package main

import (
	_ "gtest/config"
	_ "gtest/testinit"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

func main() {
	// db.test(4)
	logger.Info("------info")
	logger.Notice("------notice")
	logger.Warning("------warning")
	logger.Error("------err")
	logger.Critical("------crit")
}
