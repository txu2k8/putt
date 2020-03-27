package docs

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

	if !fileExists("./logfile") {
		createFile("./logfile")
	}
	f, err := os.Open("./logfile")
	if err != nil {
		panic(err)
	}

	backend1 := logging.NewLogBackend(io.MultiWriter(os.Stderr, f), "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)
}

// is file exists
func fileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// create file
func createFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		fo.Close()
	}()
	return nil
}

func test() {
	var log = logging.MustGetLogger("test")

	log.Info("info")
	log.Notice("notice")
	log.Warning("warning")
	log.Error("err")
	log.Critical("crit")
}
