package log

import (
	"fmt"
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Level int

const (
	InvalidLevel Level = iota - 1
	DebugLevel
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

func ParseLevel(s string) (Level, error) {
	switch strings.ToLower(s) {
	case "debug":
		return DebugLevel, nil
	case "info":
		return InfoLevel, nil
	case "warn":
		return WarnLevel, nil
	case "error":
		return ErrorLevel, nil
	case "fatal":
		return FatalLevel, nil
	default:
		return InvalidLevel, fmt.Errorf("unsupport log level %q", s)
	}
}

func (l Level) toZap() zapcore.Level {
	switch l {
	case DebugLevel:
		return zapcore.DebugLevel
	case InfoLevel:
		return zapcore.InfoLevel
	case WarnLevel:
		return zapcore.WarnLevel
	case ErrorLevel:
		return zapcore.ErrorLevel
	case FatalLevel:
		return zapcore.FatalLevel
	case InvalidLevel:
		return zapcore.InvalidLevel
	default:
		return zapcore.InvalidLevel
	}
}

type Config struct {
	Level          string
	ProductionMode bool
}

func Configure(conf Config) error {
	level, err := ParseLevel(conf.Level)
	if err != nil {
		return err
	}

	sync := zapcore.AddSync(os.Stderr)

	encConf := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  zapcore.OmitKey,
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var enc zapcore.Encoder

	if conf.ProductionMode {
		enc = zapcore.NewJSONEncoder(encConf)
	} else {
		encConf.EncodeLevel = zapcore.CapitalColorLevelEncoder
		enc = zapcore.NewConsoleEncoder(encConf)
	}

	logger := zap.New(
		zapcore.NewCore(enc, sync, level.toZap()),
		zap.WithCaller(true),
	).WithOptions(zap.AddCallerSkip(1))

	zap.ReplaceGlobals(logger)

	return nil
}
