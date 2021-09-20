package logging

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/natefinch/lumberjack"
	"github.com/sirupsen/logrus"
)

type LogrusLogger struct {
	entry *logrus.Entry
}

func (l *LogrusLogger) WithField(key string, value interface{}) Logger {
	return &LogrusLogger{entry: l.entry.WithField(key, value)}
}

func (l *LogrusLogger) WithFields(fields Fields) Logger {
	return &LogrusLogger{entry: l.entry.WithFields(logrus.Fields(fields))}
}

func (l *LogrusLogger) WithError(err error) Logger {
	return &LogrusLogger{entry: l.entry.WithError(err)}
}

func (l *LogrusLogger) Tracef(format string, args ...interface{}) {
	l.entry.Tracef(format, args...)
}

func (l *LogrusLogger) Debugf(format string, args ...interface{}) {
	l.entry.Debugf(format, args...)
}

func (l *LogrusLogger) Infof(format string, args ...interface{}) {
	l.entry.Infof(format, args...)
}

func (l *LogrusLogger) Warningf(format string, args ...interface{}) {
	l.entry.Warningf(format, args...)
}

func (l *LogrusLogger) Errorf(format string, args ...interface{}) {
	l.entry.Errorf(format, args...)
}

func (l *LogrusLogger) Fatalf(format string, args ...interface{}) {
	l.entry.Fatalf(format, args...)
}

func (l *LogrusLogger) Trace(args ...interface{}) {
	l.entry.Trace(args...)
}

func (l *LogrusLogger) Debug(args ...interface{}) {
	l.entry.Debug(args...)
}

func (l *LogrusLogger) Info(args ...interface{}) {
	l.entry.Info(args...)
}

func (l *LogrusLogger) Warning(args ...interface{}) {
	l.entry.Warning(args...)
}

func (l *LogrusLogger) Error(args ...interface{}) {
	l.entry.Error(args...)
}

func (l *LogrusLogger) Fatal(args ...interface{}) {
	l.entry.Fatal(args...)
}

func getLogrusLogger(cfg LogConfig) (*LogrusLogger, error) {
	l := logrus.New()
	l.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	logLevel := strings.ToLower(cfg.LogLevel)
	if cfg.LogLevel == "" {
		logLevel = "debug"
	}

	switch logLevel {
	case Trace:
		l.SetLevel(logrus.TraceLevel)
	case Debug:
		l.SetLevel(logrus.DebugLevel)
	case Info:
		l.SetLevel(logrus.InfoLevel)
	case Warning:
		l.SetLevel(logrus.WarnLevel)
	case Error:
		l.SetLevel(logrus.ErrorLevel)
	case Fatal:
		l.SetLevel(logrus.FatalLevel)
	default:
		return nil, fmt.Errorf("unsupported log level %s", cfg.LogLevel)
	}

	writers := []io.Writer{os.Stdout}

	if cfg.LogFilePath != "" {
		logDirPath, err := filepath.Abs(filepath.Dir(cfg.LogFilePath))
		if err != nil {
			return nil, fmt.Errorf("failed to get base log dir: %v", err)
		}

		if err = os.MkdirAll(logDirPath, 0o750); err != nil {
			return nil, fmt.Errorf("failed to create log dir %s: %v", logDirPath, err)
		}

		logFile, err := os.OpenFile(cfg.LogFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o660)
		if err != nil {
			return nil, fmt.Errorf("failed to create or open log file %s: %v", cfg.LogFilePath, err)
		}

		writers = append(writers, logFile)

		if cfg.NeedRotate {
			writers = append(writers, &lumberjack.Logger{
				Filename:   cfg.LogFilePath,
				MaxSize:    cfg.MaxSize,
				MaxAge:     cfg.MaxAge,
				MaxBackups: cfg.MaxBackups,
				LocalTime:  true,
				Compress:   true,
			})
		}
	}

	l.SetOutput(io.MultiWriter(writers...))

	return &LogrusLogger{entry: logrus.NewEntry(l)}, nil
}
