package config

import (
	"os"

	"github.com/op/go-logging"
)

// format define
const (
	logLevel    = logging.DEBUG
	InfoFormat  = `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: %{color:reset}%{message}`
	DebugFormat = `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: (%{shortfile}) %{color:reset}%{message}`
)

// Example format string. Everything except the message has a custom color
// which is dependent on the log level. Many fields have a custom output
// formatting too, eg. the time returns the hour down to the milli second.
// time:2006-01-02T15:04:05.999Z-07:00
func init() {
	// string format: DebugFormat if level>=DEBUG, else InfoFormat
	strformat := InfoFormat
	if logLevel >= logging.DEBUG {
		strformat = DebugFormat
	}
	var format = logging.MustStringFormatter(strformat)

	backend1 := logging.NewLogBackend(os.Stderr, "", 0)
	backend2 := logging.NewLogBackend(os.Stdout, "", 0)
	backend2Formatter := logging.NewBackendFormatter(backend2, format)

	// Only errors and more severe messages should be sent to backend1
	backend1Leveled := logging.AddModuleLevel(backend1)
	backend1Leveled.SetLevel(logging.ERROR, "")

	// Set the backends to be used.
	logging.SetBackend(backend1Leveled, backend2Formatter)

	var log = logging.MustGetLogger("test")
	log.Info("logging initialized ...")

}

// is file exists
func FileExists(name string) bool {
	if _, err := os.Stat(name); err != nil {
		if os.IsNotExist(err) {
			return false
		}
	}
	return true
}

// create file
func CreateFile(name string) error {
	fo, err := os.Create(name)
	if err != nil {
		return err
	}
	defer func() {
		fo.Close()
	}()
	return nil
}
