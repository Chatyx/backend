package log

import (
	"context"

	"go.uber.org/zap"
)

type ctxKey struct{}

func FromContext(ctx context.Context) *Logger {
	if logger, ok := ctx.Value(ctxKey{}).(*Logger); ok {
		return logger
	}

	return &Logger{sugar: zap.S()}
}

func WithLogger(ctx context.Context, logger *Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, logger)
}
