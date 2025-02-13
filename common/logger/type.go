package logger

type Level string

const (
	DebugLv Level = "debug"
	InfoLv  Level = "info"
	WarnLv  Level = "warn"
	ErrorLv Level = "error"
)

type ILogger interface {
	Debug(msg string, fields ...interface{})
	Info(msg string, fields ...interface{})
	Warn(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
	Fatal(msg string, fields ...interface{})
	With(fields ...interface{}) ILogger
}

type Field struct {
	Key string
	Val interface{}
}

type IDecorator interface {
	Decorate(ILogger) ILogger
}
