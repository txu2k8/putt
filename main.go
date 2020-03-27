package main

import (
	_ "config"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("test")

func test(a int) int {
	b := a * 2
	log.Info(b)
	return b
}

func main() {

	test(4)
	log.Info("This is info message")
	log.Infof("This is info message: %v", 12345)

	log.Warning("This is warning message")
	log.Warningf("This is warning message: %v", 12345)

	log.Error("This is error message")
	log.Errorf("This is error message: %v", 12345)
}
