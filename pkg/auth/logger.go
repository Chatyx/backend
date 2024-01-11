package auth

type Logger interface {
	Warnf(format string, args ...any)
	Errorf(format string, args ...any)
}

type noOpLogger struct{}

func (l noOpLogger) Warnf(string, ...any) {}

func (l noOpLogger) Errorf(string, ...any) {}
