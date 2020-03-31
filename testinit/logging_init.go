package testinit

import (
	"io"
	"os"

	"github.com/op/go-logging"
)

// define string format
// time:2006-01-02T15:04:05.999Z-07:00
const (
	logLevel    = logging.DEBUG
	InfoFormat  = `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: %{color:reset}%{message}`
	DebugFormat = `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: (%{shortfile}) %{color:reset}%{message}`
)

func init() {

	// string format: DebugFormat if level>=DEBUG, else InfoFormat
	strformat := InfoFormat
	if logLevel >= logging.DEBUG {
		strformat = DebugFormat
	}
	var format = logging.MustStringFormatter(strformat)

	// backend-1 output to Console
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
	consoleBackendFormator := logging.NewBackendFormatter(consoleBackend, format)
	consoleBackendLeveled := logging.AddModuleLevel(consoleBackendFormator)
	consoleBackendLeveled.SetLevel(logLevel, "")

	// backend-2 output to log file && Console
	file, err := os.OpenFile("test.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	fileBackend := logging.NewLogBackend(io.Writer(file), "", 0)
	fileBackendFormator := logging.NewBackendFormatter(fileBackend, format)
	fileBackendLeveled := logging.AddModuleLevel(fileBackendFormator)
	fileBackendLeveled.SetLevel(logLevel, "")

	// Set the backends to be used.
	logging.SetBackend(consoleBackendLeveled, fileBackendLeveled)
}

func test() {
	var log = logging.MustGetLogger("test")

	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
}
