package tlog

// Config for logging

import (
	"io"
	"os"
	"path"

	"github.com/op/go-logging"
)

// Config ...
type Config struct {
	FileLogPath     string
	FileLogLevel    logging.Level // log level for output to file
	ConsoleLogLevel logging.Level // log level for output to console
	InfoFormat      string        // define string format, time:2006-01-02T15:04:05.999Z-07:00
	DebugFormat     string
}

// Option is the type all options need to adhere to
type Option func(p *Config)

// OptionSetFileLogPath sets the logging output to file path
func OptionSetFileLogPath(filePath string) Option {
	return func(p *Config) {
		p.FileLogPath = filePath
	}
}

// NewOptions ...
func NewOptions(options ...Option) *Config {
	dir, _ := os.Getwd()
	c := Config{
		FileLogPath:     path.Join(dir, "log", "test1.log"),
		FileLogLevel:    logging.DEBUG,
		ConsoleLogLevel: logging.INFO,
		InfoFormat:      `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: %{color:reset}%{message}`,
		DebugFormat:     `%{color}%{time:2006-01-02T15:04:05} %{module} %{level:.4s}: (%{shortfile}) %{color:reset}%{message}`,
	}
	for _, o := range options {
		o(&c)
	}

	return &c
}

// InitLogging Config ...
func (conf *Config) InitLogging() {
	// string format: DebugFormat if level>=DEBUG, else InfoFormat
	fileStrformat := conf.InfoFormat
	consoleStrformat := conf.InfoFormat
	if conf.FileLogLevel >= logging.DEBUG {
		fileStrformat = conf.DebugFormat
	}
	if conf.ConsoleLogLevel >= logging.DEBUG {
		consoleStrformat = conf.DebugFormat
	}
	var (
		fileFormat    = logging.MustStringFormatter(fileStrformat)
		consoleFormat = logging.MustStringFormatter(consoleStrformat)
	)

	// backend-1 output to Console
	consoleBackend := logging.NewLogBackend(os.Stdout, "", 0)
	consoleBackendFormator := logging.NewBackendFormatter(consoleBackend, consoleFormat)
	consoleBackendLeveled := logging.AddModuleLevel(consoleBackendFormator)
	consoleBackendLeveled.SetLevel(conf.ConsoleLogLevel, "")

	// backend-2 output to log file && Console
	fileDir := path.Dir(conf.FileLogPath)
	err := os.MkdirAll(fileDir, os.ModePerm)
	file, err := os.OpenFile(conf.FileLogPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	fileBackend := logging.NewLogBackend(io.Writer(file), "", 0)
	fileBackendFormator := logging.NewBackendFormatter(fileBackend, fileFormat)
	fileBackendLeveled := logging.AddModuleLevel(fileBackendFormator)
	fileBackendLeveled.SetLevel(conf.FileLogLevel, "")

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