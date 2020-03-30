package main

import (
	// "gtest/config"
	_ "gtest/testinit"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("test")

func main() {
	// config.LoadConfig()
	// db.test(4)
	log.Info("------info")
	log.Notice("------notice")
	log.Warning("------warning")
	log.Error("------err")
	log.Critical("------crit")
}
