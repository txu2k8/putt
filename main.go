package main

import (
	"os"
	"pzatest/cmd"
	_ "pzatest/config"
	_ "pzatest/testinit"
	"strings"

	"github.com/op/go-logging"
)

var logger = logging.MustGetLogger("test")

func main() {
	logger.Infof("Args: pzatest %s", strings.Join(os.Args[1:], " "))
	cmd.Execute()
}
