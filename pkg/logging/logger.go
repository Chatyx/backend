package logging

type Fields map[string]interface{}

type Logger interface {
	WithField(key string, value interface{}) Logger
	WithFields(fields Fields) Logger
	WithError(err error) Logger

	Debugf(format string, args ...interface{})
	Infof(format string, args ...interface{})
	Warningf(format string, args ...interface{})
	Errorf(format string, args ...interface{})
	Fatalf(format string, args ...interface{})

	Debug(args ...interface{})
	Info(args ...interface{})
	Warning(args ...interface{})
	Error(args ...interface{})
	Fatal(args ...interface{})
}

// LogConfig describes configuration to init logging.
type LogConfig struct {
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

var gLogger Logger

// GetLogger returns logger if it isn't empty. Otherwise, it will panic.
func GetLogger() Logger {
	if gLogger == nil {
		panic("failed to get logger before initialization")
	}

	return gLogger
}
