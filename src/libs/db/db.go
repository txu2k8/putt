package db

import (
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("test")

func test(a int) int {
	b := a * 2
	log.Info(b)
	return b
}
