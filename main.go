package main

import (
	_ "config"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("test")

func test(a int) int {
	b := a * 2
	log.Info("--------test")
	return b
}

func main() {
	test(4)
	log.Info("------info")
	log.Notice("------notice")
	log.Warning("------warning")
	log.Error("------err")
	log.Critical("------crit")
}
