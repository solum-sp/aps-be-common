package logger

import "context"

type Level string

const (
	DebugLv Level = "debug"
	InfoLv  Level = "info"
	WarnLv  Level = "warn"
	ErrorLv Level = "error"
)

type ILogger interface {
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	With(fields ...Field) ILogger
}

type Field struct {
	Key string
	Val interface{}
}

type IDecorator interface {
	Decorate(ILogger) ILogger
}
