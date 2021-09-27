package logging

import (
	"fmt"
	"strings"
	"sync"
)

const (
	Trace   = "trace"
	Debug   = "debug"
	Info    = "info"
	Warning = "warning"
	Error   = "error"
	Fatal   = "fatal"
)

type Fields map[string]interface{}

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	Tracef(format string, args ...interface{})
	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Trace(args ...interface{})
	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// LogConfig describes configuration to init logging.
type LogConfig struct {
	// LoggerKind is one of the logging type: "logrus". Default value is "logrus".
	LoggerKind string

	// LogLevel is one of logging level: "trace", "debug", "info", "warning", "error".
	LogLevel string

	// LogFilePath is the output file path for log storing.
	LogFilePath string

	// NeedRotate defines if necessary to rotate log file.
	// It doesn't affect if LogFilePath is empty.
	NeedRotate bool

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes. It doesn't affect if NeedRotate is false.
	MaxSize int

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename. It doesn't affect if NeedRotate is false.
	MaxAge int

	// MaxBackups is the maximum number of old log files to retain. The default
	// is to retain all old log files. It doesn't affect if NeedRotate is false.
	MaxBackups int
}

const (
	logrusKind = "logrus"
	mockKind   = "mock"
)

var (
	logger Logger
	once   sync.Once
)

func InitLogger(cfg LogConfig) {
	var err error

	once.Do(func() {
		loggerKind := strings.ToLower(cfg.LoggerKind)
		if loggerKind == "" {
			loggerKind = logrusKind
		}

		switch loggerKind {
		case logrusKind:
			logger, err = getLogrusLogger(cfg)
			if err != nil {
				panic(err)
			}
		case mockKind:
			logger = LoggerMock{}
		default:
			panic(fmt.Sprintf("unsupported logger kind %q", cfg.LoggerKind))
		}
	})
}

// GetLogger returns logger if it isn't empty. Otherwise, it will panic.
func GetLogger() Logger {
	if logger == nil {
		panic("failed to get logger before initialization, call InitLogger()")
	}

	return logger
}
