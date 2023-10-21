package log

import (
	"go.uber.org/zap"
)

type Logger struct {
	sugar *zap.SugaredLogger
}

func (l *Logger) Debug(msg string, args ...any) {
	l.sugar.Debugw(msg, args...)
}

func (l *Logger) Debugf(format string, args ...any) {
	l.sugar.Debugf(format, args...)
}

func (l *Logger) Info(msg string, args ...any) {
	l.sugar.Infow(msg, args...)
}

func (l *Logger) Infof(format string, args ...any) {
	l.sugar.Infof(format, args...)
}

func (l *Logger) Warn(msg string, args ...any) {
	l.sugar.Warnw(msg, args...)
}

func (l *Logger) Warnf(format string, args ...any) {
	l.sugar.Warnf(format, args...)
}

func (l *Logger) Error(msg string, args ...any) {
	l.sugar.Errorw(msg, args...)
}

func (l *Logger) Errorf(format string, args ...any) {
	l.sugar.Errorf(format, args...)
}

func (l *Logger) Fatal(msg string, args ...any) {
	l.sugar.Fatalw(msg, args...)
}

func (l *Logger) Fatalf(format string, args ...any) {
	l.sugar.Fatalf(format, args...)
}

func (l *Logger) With(args ...any) *Logger {
	return &Logger{sugar: l.sugar.With(args...)}
}

func (l *Logger) WithError(err error) *Logger {
	return &Logger{sugar: l.sugar.Desugar().With(zap.Error(err)).Sugar()}
}

func Debug(msg string, args ...any) {
	zap.S().Debugw(msg, args...)
}

func Debugf(format string, args ...any) {
	zap.S().Debugf(format, args...)
}

func Info(msg string, args ...any) {
	zap.S().Infow(msg, args...)
}

func Infof(format string, args ...any) {
	zap.S().Infof(format, args...)
}

func Warn(msg string, args ...any) {
	zap.S().Warnw(msg, args...)
}

func Warnf(format string, args ...any) {
	zap.S().Warnf(format, args...)
}

func Error(msg string, args ...any) {
	zap.S().Errorw(msg, args...)
}

func Errorf(format string, args ...any) {
	zap.S().Errorf(format, args...)
}

func Fatal(msg string, args ...any) {
	zap.S().Fatalw(msg, args...)
}

func Fatalf(format string, args ...any) {
	zap.S().Fatalf(format, args...)
}

func With(args ...any) *Logger {
	return &Logger{sugar: zap.S().With(args...)}
}

func WithError(err error) *Logger {
	return &Logger{sugar: zap.L().With(zap.Error(err)).Sugar()}
}
