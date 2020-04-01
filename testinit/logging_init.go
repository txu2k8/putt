package testinit

import (
	"io"
	"os"

	"github.com/op/go-logging"
)

// define string format
// time:2006-01-02T15:04:05.999Z-07:00
const (
	fileLogLevel    = logging.DEBUG // log level for output to file
	consoleLogLevel = logging.INFO  // log level for output to console
	InfoFormat      = `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: %{color:reset}%{message}`
	DebugFormat     = `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: (%{shortfile}) %{color:reset}%{message}`
)

func init() {
	// string format: DebugFormat if level>=DEBUG, else InfoFormat
	fileStrformat := InfoFormat
	consoleStrformat := InfoFormat
	if fileLogLevel >= logging.DEBUG {
		fileStrformat = DebugFormat
	}
	if consoleLogLevel >= logging.DEBUG {
		consoleStrformat = DebugFormat
	}
	var (
		fileFormat    = logging.MustStringFormatter(fileStrformat)
		consoleFormat = logging.MustStringFormatter(consoleStrformat)
	)

	// backend-1 output to Console
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
	consoleBackendFormator := logging.NewBackendFormatter(consoleBackend, consoleFormat)
	consoleBackendLeveled := logging.AddModuleLevel(consoleBackendFormator)
	consoleBackendLeveled.SetLevel(consoleLogLevel, "")

	// backend-2 output to log file && Console
	file, err := os.OpenFile("test.log",
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	fileBackend := logging.NewLogBackend(io.Writer(file), "", 0)
	fileBackendFormator := logging.NewBackendFormatter(fileBackend, fileFormat)
	fileBackendLeveled := logging.AddModuleLevel(fileBackendFormator)
	fileBackendLeveled.SetLevel(fileLogLevel, "")

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
