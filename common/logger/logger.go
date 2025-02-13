package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type zapLogger struct {
	logger  *zap.Logger
	service string
}

type Config struct {
	Service string
	Level   Level
}

func NewLogger(config Config) (*zapLogger, error) {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	encoder := zapcore.NewJSONEncoder(encoderConfig)

	var level zapcore.Level
	switch config.Level {
	case DebugLv:
		level = zapcore.DebugLevel
	case InfoLv:
		level = zapcore.InfoLevel
	case WarnLv:
		level = zapcore.WarnLevel
	case ErrorLv:
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	core := zapcore.NewCore(
		encoder,
		zapcore.AddSync(os.Stderr),
		level,
	)

	zlogger := zap.New(
		core,
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	return &zapLogger{
		logger:  zlogger,
		service: config.Service,
	}, nil
}

// Methods implementing the Logger interface
func (l *zapLogger) Debug(msg string, fields ...interface{}) {
	l.log(zap.DebugLevel, msg, fields...)
}

func (l *zapLogger) Info(msg string, fields ...interface{}) {
	l.log(zap.InfoLevel, msg, fields...)
}

func (l *zapLogger) Warn(msg string, fields ...interface{}) {
	l.log(zap.WarnLevel, msg, fields...)
}

func (l *zapLogger) Error(msg string, fields ...interface{}) {
	l.log(zap.ErrorLevel, msg, fields...)
}

func (l *zapLogger) Fatal(msg string, fields ...interface{}) {
	l.log(zap.FatalLevel, msg, fields...)
}

func (l *zapLogger) With(fields ...interface{}) ILogger {
	zapFields := toZapFields(fields...)
	return &zapLogger{
		logger:  l.logger.With(zapFields...),
		service: l.service,
	}
}

func (l *zapLogger) log(level zapcore.Level, msg string, fields ...interface{}) {
	// Convert our Fields to zap.Fields
	zapFields := make([]zap.Field, 0, len(fields)+1)

	// Add service name
	zapFields = append(zapFields, zap.String("service", l.service))

	// Add custom fields
	zapFields = append(zapFields, toZapFields(fields)...)

	// Log with appropriate level
	switch level {
	case zap.DebugLevel:
		l.logger.Debug(msg, zapFields...)
	case zap.InfoLevel:
		l.logger.Info(msg, zapFields...)
	case zap.WarnLevel:
		l.logger.Warn(msg, zapFields...)
	case zap.ErrorLevel:
		zapFields = append(zapFields, captureStackTrace())
		l.logger.Error(msg, zapFields...)
	case zap.FatalLevel:
		zapFields = append(zapFields, captureStackTrace())
		l.logger.Fatal(msg, zapFields...)
	}
}

// sanitize cleanses sensitive data from log fields
func sanitize(value interface{}) interface{} {
	// Convert to string for analysis
	str, ok := value.(string)
	if !ok {
		// If it's not a string, try to marshal to JSON
		jsonBytes, err := json.Marshal(value)
		if err != nil {
			return value
		}
		str = string(jsonBytes)
	}

	// List of sensitive field names (case-insensitive)
	sensitiveFields := []string{
		"password",
		"token",
		"authorization",
		"api_key",
		"secret",
	}

	// Check if the value contains sensitive information
	strLower := strings.ToLower(str)
	for _, field := range sensitiveFields {
		if strings.Contains(strLower, field) {
			return "[REDACTED]"
		}
	}

	return value
}

func toZapFields(fields ...interface{}) []zap.Field {
	zapFields := make([]zap.Field, len(fields))
	for i, field := range fields {
		zapFields[i] = zap.Any(fmt.Sprintf("field_%d", i), field) // Dynamically assign field names
	}
	return zapFields
}

// Capture stack trace as Zap field
func captureStackTrace() zap.Field {
	pc := make([]uintptr, 10)
	runtime.Callers(3, pc) // Skip 3 frames
	frames := runtime.CallersFrames(pc)
	var stacktrace string
	for frame, more := frames.Next(); more; frame, more = frames.Next() {
		stacktrace += fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
	}
	return zap.String("stacktrace", stacktrace)
}
