package auth

type Logger interface {
	Infof(format string, args ...any)
	Warnf(format string, args ...any)
}

type noOpLogger struct{}

func (l noOpLogger) Infof(string, ...any) {}

func (l noOpLogger) Warnf(string, ...any) {}
