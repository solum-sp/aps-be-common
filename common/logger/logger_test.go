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

	tests := []struct {
		level   zapcore.Level
		logFunc func(msg string, fields ...interface{})
	}{
		{zapcore.DebugLevel, func(msg string, fields ...interface{}) { testLogger.Debug(msg, fields) }},
		{zapcore.InfoLevel, func(msg string, fields ...interface{}) { testLogger.Info(msg, fields) }},
		{zapcore.WarnLevel, func(msg string, fields ...interface{}) { testLogger.Warn(msg, fields) }},
		{zapcore.ErrorLevel, func(msg string, fields ...interface{}) { testLogger.Error(msg, fields) }},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			recorded.TakeAll() // Clear logs
			msg := "test message"
			tt.logFunc(msg, "test_key", "test_value")

			logs := recorded.All()
			assert.Equal(t, 1, len(logs))
			assert.Equal(t, tt.level, logs[0].Level)
			assert.Equal(t, msg, logs[0].Message)
			ctxMap := logs[0].ContextMap()
			assert.Contains(t, ctxMap, "service")
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

// mockLogger is a mock implementation of ILogger for testing opentelemetry decorator
type MockLogger struct {
	mock.Mock
}

func (m *MockLogger) Debug(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Info(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Warn(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Error(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) Fatal(msg string, fields ...interface{}) {
	m.Called(msg, fields)
}

func (m *MockLogger) With(fields ...interface{}) ILogger {
	args := m.Called(fields)
	return args.Get(0).(ILogger)
}

func TestOpenTelemetryDecorator(t *testing.T) {
	tests := []struct {
		name      string
		level     string
		msg       string
		fields    []interface{}
		withTrace bool
	}{
		{
			name:      "Debug with trace",
			level:     "debug",
			msg:       "debug message",
			fields:    []interface{}{"key", "value"},
			withTrace: true,
		},
		{
			name:      "Info without trace",
			level:     "info",
			msg:       "info message",
			fields:    []interface{}{"key", "value"},
			withTrace: false,
		},
		{
			name:      "Error with fields",
			level:     "error",
			msg:       "error message",
			fields:    []interface{}{"error", "test error"},
			withTrace: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			mockLogger := new(MockLogger)
			decorator := NewOpenTelemetryDecorator()
			decoratedLogger := decorator.Decorate(mockLogger).(*openTelemetryLogger)

			// Create trace context if needed
			ctx := context.Background()
			if tt.withTrace {
				sc := trace.NewSpanContext(trace.SpanContextConfig{
					TraceID: trace.TraceID{0x01},
					SpanID:  trace.SpanID{0x01},
				})
				ctx = trace.ContextWithSpanContext(ctx, sc)
			}

			// Set expectations
			expectedFields := tt.fields
			if tt.withTrace {
				expectedFields = append(expectedFields,
					"trace_id", mock.Anything,
					"span_id", mock.Anything,
				)
			}

			mockLogger.On(tt.level, tt.msg, expectedFields).Return()

			// Execute
			switch tt.level {
			case "debug":
				decoratedLogger.Debug(tt.msg, tt.fields...)
			case "info":
				decoratedLogger.Info(tt.msg, tt.fields...)
			case "warn":
				decoratedLogger.Warn(tt.msg, tt.fields...)
			case "error":
				decoratedLogger.Error(tt.msg, tt.fields...)
			}

			// Assert
			mockLogger.AssertExpectations(t)
		})
	}

	t.Run("With method", func(t *testing.T) {
		mockLogger := new(MockLogger)
		decorator := NewOpenTelemetryDecorator()
		decoratedLogger := decorator.Decorate(mockLogger)

		fields := []interface{}{"key", "value"}
		mockLogger.On("With", fields).Return(mockLogger)

		newLogger := decoratedLogger.With(fields...)
		assert.NotNil(t, newLogger)
		mockLogger.AssertExpectations(t)
	})
}
