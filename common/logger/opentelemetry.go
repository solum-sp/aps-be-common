package logger

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

type OpenTelemetryDecorator struct{}

func NewOpenTelemetryDecorator() *OpenTelemetryDecorator {
	return &OpenTelemetryDecorator{}
}

func (d *OpenTelemetryDecorator) Decorate(logger ILogger) ILogger {
	return &openTelemetryLogger{
		logger: logger,
	}
}

type openTelemetryLogger struct {
	logger ILogger
}

func (l *openTelemetryLogger) log(level string, msg string, fields ...interface{}) {
	span := trace.SpanFromContext(context.Background())
	if span.SpanContext().IsValid() {
		traceID := span.SpanContext().TraceID().String()
		spanID := span.SpanContext().SpanID().String()

		fields = append(fields, "trace_id", traceID, "span_id", spanID)
	}

	switch level {
	case "debug":
		l.logger.Debug(msg, fields...)
	case "info":
		l.logger.Info(msg, fields...)
	case "warn":
		l.logger.Warn(msg, fields...)
	case "error":
		l.logger.Error(msg, fields...)
	case "fatal":
		l.logger.Fatal(msg, fields...)
	}
}

func (l *openTelemetryLogger) Debug(msg string, fields ...interface{}) {
	l.log("debug", msg, fields...)
}

func (l *openTelemetryLogger) Info(msg string, fields ...interface{}) {
	l.log("info", msg, fields...)
}

func (l *openTelemetryLogger) Warn(msg string, fields ...interface{}) {
	l.log("warn", msg, fields...)
}

func (l *openTelemetryLogger) Error(msg string, fields ...interface{}) {
	l.log("error", msg, fields...)
}

func (l *openTelemetryLogger) Fatal(msg string, fields ...interface{}) {
	l.log("fatal", msg, fields...)
}

func (l *openTelemetryLogger) With(fields ...interface{}) ILogger {
	return &openTelemetryLogger{
		logger: l.logger.With(fields...),
	}
}
