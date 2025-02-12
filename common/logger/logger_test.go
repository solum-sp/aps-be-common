package logger

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"go.uber.org/zap/zaptest/observer"
)

func TestNewLogger(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		wantErr     bool
		wantService string
	}{
		{
			name: "valid config",
			config: Config{
				Service: "test-service",
				Level:   InfoLv,
			},
			wantErr:     false,
			wantService: "test-service",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger, err := NewLogger(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.wantService, logger.service)
		})
	}
}

func TestLogLevels(t *testing.T) {
	// Create an observer to capture logs
	core, recorded := observer.New(zapcore.DebugLevel)
	testLogger := &zapLogger{
		logger:  zap.New(core),
		service: "test-service",
	}

	ctx := context.Background()
	tests := []struct {
		level   zapcore.Level
		logFunc func(msg string, fields ...Field)
	}{
		{zapcore.DebugLevel, func(msg string, fields ...Field) { testLogger.Debug(ctx, msg, fields...) }},
		{zapcore.InfoLevel, func(msg string, fields ...Field) { testLogger.Info(ctx, msg, fields...) }},
		{zapcore.WarnLevel, func(msg string, fields ...Field) { testLogger.Warn(ctx, msg, fields...) }},
		{zapcore.ErrorLevel, func(msg string, fields ...Field) { testLogger.Error(ctx, msg, fields...) }},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			recorded.TakeAll() // Clear logs
			msg := "test message"
			fields := []Field{{Key: "test_key", Val: "test_value"}}

			tt.logFunc(msg, fields...)

			logs := recorded.All()
			assert.Equal(t, 1, len(logs))
			assert.Equal(t, tt.level, logs[0].Level)
			assert.Equal(t, msg, logs[0].Message)
			assert.Contains(t, logs[0].ContextMap(), "test_key")
		})
	}
}

func TestSanitize(t *testing.T) {
	tests := []struct {
		name  string
		input interface{}
		want  interface{}
	}{
		{
			name:  "password field",
			input: map[string]string{"password": "secret123"},
			want:  "[REDACTED]",
		},
		{
			name:  "regular field",
			input: map[string]string{"username": "john"},
			want:  map[string]string{"username": "john"},
		},
		{
			name:  "nested sensitive field",
			input: map[string]interface{}{"data": map[string]string{"api_key": "xyz123"}},
			want:  "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := sanitize(tt.input)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestCaptureStackTrace(t *testing.T) {
	field := captureStackTrace()
	assert.Equal(t, "stacktrace", field.Key)
	assert.NotEmpty(t, field.String)
	assert.Contains(t, field.String, "logger_test.go") // Should contain this file name
}

func TestLoggerWith(t *testing.T) {
	core, recorded := observer.New(zapcore.DebugLevel)
	logger := &zapLogger{
		logger:  zap.New(core),
		service: "test-service",
	}

	fields := []Field{
		{Key: "field1", Val: "value1"},
		{Key: "field2", Val: 123},
	}

	newLogger := logger.With(fields...)
	newZapLogger, ok := newLogger.(*zapLogger)
	assert.True(t, ok)
	assert.NotNil(t, newZapLogger)

	// Test logging with the new logger
	ctx := context.Background()
	newLogger.Info(ctx, "test message")

	logs := recorded.All()
	assert.Equal(t, 1, len(logs))
	assert.Contains(t, logs[0].ContextMap(), "field1")
	assert.Contains(t, logs[0].ContextMap(), "field2")
}

// mockLogger is a mock implementation of ILogger for testing opentelemetry decorator
type mockLogger struct {
	mock.Mock
}

func (m *mockLogger) Debug(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Info(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Warn(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Error(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) Fatal(ctx context.Context, msg string, fields ...Field) {
	m.Called(ctx, msg, fields)
}

func (m *mockLogger) With(fields ...Field) ILogger {
	args := m.Called(fields)
	return args.Get(0).(ILogger)
}

func TestOpenTelemetryDecoratorLogLevels(t *testing.T) {
	mockLogger := new(mockLogger)
	decorator := NewOpenTelemetryDecorator()
	decoratedLogger := decorator.Decorate(mockLogger)

	// Create context with span
	ctx := context.Background()
	sc := trace.NewSpanContext(trace.SpanContextConfig{
		TraceID:    trace.TraceID{0x1},
		SpanID:     trace.SpanID{0x1},
		TraceFlags: trace.FlagsSampled,
	})
	ctxWithSpan := trace.ContextWithSpanContext(ctx, sc)

	tests := []struct {
		name    string
		ctx     context.Context
		level   string
		logFunc func(ctx context.Context, msg string, fields ...Field)
	}{
		{
			name:    "debug with span",
			ctx:     ctxWithSpan,
			level:   "Debug",
			logFunc: decoratedLogger.Debug,
		},
		{
			name:    "info with span",
			ctx:     ctxWithSpan,
			level:   "Info",
			logFunc: decoratedLogger.Info,
		},
		{
			name:    "warn with span",
			ctx:     ctxWithSpan,
			level:   "Warn",
			logFunc: decoratedLogger.Warn,
		},
		{
			name:    "error with span",
			ctx:     ctxWithSpan,
			level:   "Error",
			logFunc: decoratedLogger.Error,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock expectation
			mockLogger.On(tt.level, tt.ctx, "test message", mock.Anything).Once().Return()

			// Execute test
			tt.logFunc(tt.ctx, "test message")

			// Verify expectations
			mockLogger.AssertExpectations(t)

			// Verify trace fields were added
			calls := mockLogger.Calls
			lastCall := calls[len(calls)-1]
			fields := lastCall.Arguments.Get(2).([]Field)

			var hasTraceID, hasSpanID bool
			for _, f := range fields {
				if f.Key == "trace_id" {
					hasTraceID = true
				}
				if f.Key == "span_id" {
					hasSpanID = true
				}
			}
			assert.True(t, hasTraceID, "trace_id field should be present")
			assert.True(t, hasSpanID, "span_id field should be present")
		})
	}
}

func TestOpenTelemetryDecorator_With(t *testing.T) {
	mockLogger := new(mockLogger)
	decorator := NewOpenTelemetryDecorator()
	decoratedLogger := decorator.Decorate(mockLogger)

	// Setup mock expectation
	fields := []Field{{Key: "test", Val: "value"}}
	mockLogger.On("With", fields).Return(mockLogger)

	// Execute test
	result := decoratedLogger.With(fields...)

	// Verify expectations
	assert.NotNil(t, result)
	mockLogger.AssertExpectations(t)
}
