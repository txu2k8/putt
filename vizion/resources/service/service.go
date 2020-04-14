package service

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("test")

type Service

func test(a int) int {
	b := a * 2
	log.Info(b)
	return b
}
