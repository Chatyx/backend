package logging

type LoggerMock struct{}

func (l LoggerMock) WithField(_ string, _ interface{}) Logger { return l }

func (l LoggerMock) WithFields(_ Fields) Logger { return l }

func (l LoggerMock) WithError(_ error) Logger { return l }

func (l LoggerMock) Tracef(_ string, _ ...interface{}) {}

func (l LoggerMock) Debugf(_ string, _ ...interface{}) {}

func (l LoggerMock) Infof(_ string, _ ...interface{}) {}

func (l LoggerMock) Warningf(_ string, _ ...interface{}) {}

func (l LoggerMock) Errorf(_ string, _ ...interface{}) {}

func (l LoggerMock) Fatalf(_ string, _ ...interface{}) {}

func (l LoggerMock) Trace(_ ...interface{}) {}

func (l LoggerMock) Debug(_ ...interface{}) {}

func (l LoggerMock) Info(_ ...interface{}) {}

func (l LoggerMock) Warning(_ ...interface{}) {}

func (l LoggerMock) Error(_ ...interface{}) {}

func (l LoggerMock) Fatal(_ ...interface{}) {}
