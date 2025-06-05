package cron

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/solum-sp/aps-be-common/common/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// mockLogger implements the logger.ILogger interface for testing
type mockLogger struct {
	logs []string
	mu   sync.Mutex
}

func newMockLogger() *mockLogger {
	return &mockLogger{
		logs: make([]string, 0),
	}
}

func (m *mockLogger) Info(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Error(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Debug(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Warn(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) Fatal(msg string, keysAndValues ...interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.logs = append(m.logs, msg)
}

func (m *mockLogger) With(keysAndValues ...interface{}) logger.ILogger {
	return m
}

func (m *mockLogger) getLogs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.logs...)
}

// testJob is a simple job for testing
type testJob struct {
	name     string
	executed chan bool
	err      error
	delay    time.Duration
}

func newTestJob(name string) *testJob {
	return &testJob{
		name:     name,
		executed: make(chan bool, 1),
	}
}

func (t *testJob) Run(ctx context.Context) error {
	if t.delay > 0 {
		time.Sleep(t.delay)
	}
	t.executed <- true
	return t.err
}

func (t *testJob) GetName() string {
	return t.name
}

func (t *testJob) setError(err error) {
	t.err = err
}

func (t *testJob) setDelay(delay time.Duration) {
	t.delay = delay
}

func (t *testJob) waitExecution(timeout time.Duration) bool {
	select {
	case <-t.executed:
		return true
	case <-time.After(timeout):
		return false
	}
}

func TestCronManager_Basic(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	assert.False(t, manager.IsRunning())
	assert.Equal(t, 0, manager.GetJobCount())

	job := newTestJob("test-job")
	err := manager.AddJob("*/1 * * * * *", job) // Every second
	require.NoError(t, err)

	assert.Equal(t, 1, manager.GetJobCount())
	assert.Contains(t, manager.GetJobNames(), "test-job")

	manager.Start()
	assert.True(t, manager.IsRunning())

	// Wait for job execution
	assert.True(t, job.waitExecution(2*time.Second))

	manager.Stop()
	assert.False(t, manager.IsRunning())
}

func TestCronManager_JobInfo(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	job := newTestJob("info-test-job")
	err := manager.AddJob("*/1 * * * * *", job)
	require.NoError(t, err)

	// Check initial job info
	info, exists := manager.GetJobInfo("info-test-job")
	require.True(t, exists)
	assert.Equal(t, "info-test-job", info.Name)
	assert.Equal(t, "*/1 * * * * *", info.Schedule)
	assert.Equal(t, JobStatusIdle, info.Status)
	assert.Nil(t, info.LastRun)
	assert.Equal(t, int64(0), info.RunCount)

	manager.Start()
	defer manager.Stop()

	// Wait for job execution
	assert.True(t, job.waitExecution(2*time.Second))

	// Check updated job info
	time.Sleep(100 * time.Millisecond) // Allow metadata update
	info, exists = manager.GetJobInfo("info-test-job")
	require.True(t, exists)
	assert.Equal(t, JobStatusSuccess, info.Status)
	assert.NotNil(t, info.LastRun)
	assert.Equal(t, int64(1), info.RunCount)
	assert.Equal(t, int64(0), info.FailCount)
}

func TestCronManager_JobFailure(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	job := newTestJob("failing-job")
	job.setError(errors.New("test error"))

	err := manager.AddJob("*/1 * * * * *", job)
	require.NoError(t, err)

	manager.Start()
	defer manager.Stop()

	// Wait for job execution
	assert.True(t, job.waitExecution(2*time.Second))

	// Check job info shows failure
	time.Sleep(100 * time.Millisecond) // Allow metadata update
	info, exists := manager.GetJobInfo("failing-job")
	require.True(t, exists)
	assert.Equal(t, JobStatusFailed, info.Status)
	assert.Equal(t, "test error", info.LastError)
	assert.Equal(t, int64(1), info.FailCount)
}

func TestCronManager_RemoveJob(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	job := newTestJob("removable-job")
	err := manager.AddJob("*/1 * * * * *", job)
	require.NoError(t, err)

	assert.Equal(t, 1, manager.GetJobCount())

	manager.RemoveJob("removable-job")
	assert.Equal(t, 0, manager.GetJobCount())

	// Verify job info is also removed
	_, exists := manager.GetJobInfo("removable-job")
	assert.False(t, exists)
}

func TestCronManager_ReplaceJob(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	job1 := newTestJob("replaceable-job")
	err := manager.AddJob("*/1 * * * * *", job1)
	require.NoError(t, err)

	job2 := newTestJob("replaceable-job")       // Same name
	err = manager.AddJob("*/2 * * * * *", job2) // Different schedule
	require.NoError(t, err)

	assert.Equal(t, 1, manager.GetJobCount()) // Should still be 1

	info, exists := manager.GetJobInfo("replaceable-job")
	require.True(t, exists)
	assert.Equal(t, "*/2 * * * * *", info.Schedule) // Should have new schedule
}

func TestCronManager_GetAllJobsInfo(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	jobs := []string{"job1", "job2", "job3"}
	for _, jobName := range jobs {
		job := newTestJob(jobName)
		err := manager.AddJob("*/1 * * * * *", job)
		require.NoError(t, err)
	}

	allInfo := manager.GetAllJobsInfo()
	assert.Equal(t, 3, len(allInfo))

	for _, jobName := range jobs {
		info, exists := allInfo[jobName]
		assert.True(t, exists)
		assert.Equal(t, jobName, info.Name)
	}
}

func TestCronManager_PeriodicJob(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)

	job := newTestJob("periodic-job")
	err := manager.AddPeriodicJob(1*time.Second, job)
	require.NoError(t, err)

	manager.Start()
	defer manager.Stop()

	// Wait for job execution
	assert.True(t, job.waitExecution(2*time.Second))
}

func TestJobRegistry_Basic(t *testing.T) {
	logger := newMockLogger()
	registry := NewJobRegistry(logger)

	// Register a job factory
	registry.RegisterJob("test-job", func() Job {
		return newTestJob("test-job")
	})

	// Check registered jobs
	jobs := registry.GetRegisteredJobs()
	assert.Contains(t, jobs, "test-job")

	// Create job instance
	job, err := registry.CreateJob("test-job")
	require.NoError(t, err)
	assert.Equal(t, "test-job", job.GetName())

	// Try to create non-existent job
	_, err = registry.CreateJob("non-existent")
	assert.Error(t, err)
}

func TestJobRegistry_WithManager(t *testing.T) {
	logger := newMockLogger()
	manager := NewCronManager(logger)
	registry := NewJobRegistry(logger)

	// Register job and config
	registry.RegisterJob("registry-job", func() Job {
		return newTestJob("registry-job")
	})
	registry.SetJobConfig("registry-job", JobConfig{
		Name:     "registry-job",
		Schedule: "*/1 * * * * *",
		Enabled:  true,
	})

	// Register jobs with manager
	err := registry.RegisterJobsWithManager(manager)
	require.NoError(t, err)

	assert.Equal(t, 1, manager.GetJobCount())
}

func TestJobRegistryBuilder(t *testing.T) {
	logger := newMockLogger()

	registry := NewJobRegistryBuilder(logger).
		WithJob("builder-job", func() Job {
			return newTestJob("builder-job")
		}).
		WithConfig("builder-job", "*/1 * * * * *", true).
		Build()

	jobs := registry.GetRegisteredJobs()
	assert.Contains(t, jobs, "builder-job")

	config, exists := registry.GetJobConfig("builder-job")
	assert.True(t, exists)
	assert.Equal(t, "*/1 * * * * *", config.Schedule)
	assert.True(t, config.Enabled)
}

func TestBaseJob(t *testing.T) {
	job := NewBaseJob("base-test")
	assert.Equal(t, "base-test", job.GetName())
}

func TestFuncJob(t *testing.T) {
	executed := false
	job := NewFuncJob("func-test", func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.Equal(t, "func-test", job.GetName())

	err := job.Run(context.Background())
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestPeriodicJob(t *testing.T) {
	executed := false
	job := NewPeriodicJob("periodic-test", 5*time.Minute, func(ctx context.Context) error {
		executed = true
		return nil
	})

	assert.Equal(t, "periodic-test", job.GetName())
	assert.Equal(t, "0 */5 * * * *", job.GetSchedule())

	err := job.Run(context.Background())
	assert.NoError(t, err)
	assert.True(t, executed)
}

func TestExampleJobs(t *testing.T) {
	// Create a real logger for the examples
	zapLogger, _ := zap.NewDevelopment()
	logger := &zapLoggerWrapper{zapLogger}

	// Test database cleanup job
	dbJob := NewDatabaseCleanupJob(30, logger)
	assert.Equal(t, "database_cleanup", dbJob.GetName())
	err := dbJob.Run(context.Background())
	assert.NoError(t, err)

	// Test cache warmup job
	cacheJob := NewCacheWarmupJob([]string{"key1", "key2"}, logger)
	assert.Equal(t, "cache_warmup", cacheJob.GetName())
	err = cacheJob.Run(context.Background())
	assert.NoError(t, err)

	// Test health check job (with success)
	healthJob := NewHealthCheckJob([]string{"http://api.example.com/health"}, logger)
	assert.Equal(t, "health_check", healthJob.GetName())
	err = healthJob.Run(context.Background())
	assert.NoError(t, err)

	// Test health check job (with failure)
	failingHealthJob := NewHealthCheckJob([]string{"http://example.com/unhealthy"}, logger)
	err = failingHealthJob.Run(context.Background())
	assert.Error(t, err)

	// Test report generation job
	reportJob := NewReportGenerationJob("daily", "/tmp/test.pdf", logger)
	assert.Equal(t, "report_generation", reportJob.GetName())
	err = reportJob.Run(context.Background())
	assert.NoError(t, err)
}

func TestSetupExampleJobs(t *testing.T) {
	// Create a real logger for the examples
	zapLogger, _ := zap.NewDevelopment()
	logger := &zapLoggerWrapper{zapLogger}

	manager, registry := SetupExampleJobs(logger)

	assert.NotNil(t, manager)
	assert.NotNil(t, registry)
	assert.Equal(t, 4, manager.GetJobCount())

	expectedJobs := []string{"database_cleanup", "cache_warmup", "health_check", "daily_report"}
	for _, jobName := range expectedJobs {
		info, exists := manager.GetJobInfo(jobName)
		assert.True(t, exists, "Job %s should exist", jobName)
		assert.Equal(t, jobName, info.Name)
	}
}

// zapLoggerWrapper wraps zap.Logger to implement our ILogger interface
type zapLoggerWrapper struct {
	*zap.Logger
}

func (z *zapLoggerWrapper) Info(msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	z.Logger.Info(msg, fields...)
}

func (z *zapLoggerWrapper) Error(msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	z.Logger.Error(msg, fields...)
}

func (z *zapLoggerWrapper) Debug(msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	z.Logger.Debug(msg, fields...)
}

func (z *zapLoggerWrapper) Warn(msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	z.Logger.Warn(msg, fields...)
}

func (z *zapLoggerWrapper) Fatal(msg string, keysAndValues ...interface{}) {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	z.Logger.Fatal(msg, fields...)
}

func (z *zapLoggerWrapper) With(keysAndValues ...interface{}) logger.ILogger {
	fields := make([]zap.Field, 0, len(keysAndValues)/2)
	for i := 0; i < len(keysAndValues); i += 2 {
		if i+1 < len(keysAndValues) {
			key := keysAndValues[i].(string)
			value := keysAndValues[i+1]
			fields = append(fields, zap.Any(key, value))
		}
	}
	return &zapLoggerWrapper{z.Logger.With(fields...)}
}
