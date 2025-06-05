package cron

import (
	"context"
	"fmt"
	"time"

	"github.com/solum-sp/aps-be-common/common/logger"
)

// DatabaseCleanupJob is an example job that cleans up old database records
type DatabaseCleanupJob struct {
	*BaseJob
	retentionDays int
	logger        logger.ILogger
}

// NewDatabaseCleanupJob creates a new database cleanup job
func NewDatabaseCleanupJob(retentionDays int, logger logger.ILogger) *DatabaseCleanupJob {
	return &DatabaseCleanupJob{
		BaseJob:       NewBaseJob("database_cleanup"),
		retentionDays: retentionDays,
		logger:        logger,
	}
}

// Run executes the database cleanup job
func (d *DatabaseCleanupJob) Run(ctx context.Context) error {
	d.logger.Info("Starting database cleanup", "retention_days", d.retentionDays)

	// Simulate database cleanup work
	time.Sleep(100 * time.Millisecond)

	// Example: Delete records older than retention period
	cutoffDate := time.Now().AddDate(0, 0, -d.retentionDays)
	d.logger.Info("Cleaning up records", "cutoff_date", cutoffDate.Format("2006-01-02"))

	// In a real implementation, you would:
	// 1. Connect to your database
	// 2. Execute cleanup queries
	// 3. Log the number of records cleaned up

	d.logger.Info("Database cleanup completed successfully")
	return nil
}

// CacheWarmupJob is an example job that warms up application caches
type CacheWarmupJob struct {
	*BaseJob
	cacheKeys []string
	logger    logger.ILogger
}

// NewCacheWarmupJob creates a new cache warmup job
func NewCacheWarmupJob(cacheKeys []string, logger logger.ILogger) *CacheWarmupJob {
	return &CacheWarmupJob{
		BaseJob:   NewBaseJob("cache_warmup"),
		cacheKeys: cacheKeys,
		logger:    logger,
	}
}

// Run executes the cache warmup job
func (c *CacheWarmupJob) Run(ctx context.Context) error {
	c.logger.Info("Starting cache warmup", "keys_count", len(c.cacheKeys))

	for _, key := range c.cacheKeys {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Simulate cache warmup for each key
			c.logger.Debug("Warming up cache key", "key", key)
			time.Sleep(10 * time.Millisecond)
		}
	}

	c.logger.Info("Cache warmup completed successfully")
	return nil
}

// HealthCheckJob is an example job that performs system health checks
type HealthCheckJob struct {
	*BaseJob
	endpoints []string
	logger    logger.ILogger
}

// NewHealthCheckJob creates a new health check job
func NewHealthCheckJob(endpoints []string, logger logger.ILogger) *HealthCheckJob {
	return &HealthCheckJob{
		BaseJob:   NewBaseJob("health_check"),
		endpoints: endpoints,
		logger:    logger,
	}
}

// Run executes the health check job
func (h *HealthCheckJob) Run(ctx context.Context) error {
	h.logger.Info("Starting health check", "endpoints_count", len(h.endpoints))

	var failures []string

	for _, endpoint := range h.endpoints {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// Simulate health check
			h.logger.Debug("Checking endpoint", "endpoint", endpoint)

			// In a real implementation, you would:
			// 1. Make HTTP request to the endpoint
			// 2. Check response status and content
			// 3. Record metrics

			// Simulate occasional failure
			if endpoint == "http://example.com/unhealthy" {
				failures = append(failures, endpoint)
			}

			time.Sleep(50 * time.Millisecond)
		}
	}

	if len(failures) > 0 {
		return fmt.Errorf("health check failed for endpoints: %v", failures)
	}

	h.logger.Info("Health check completed successfully")
	return nil
}

// ReportGenerationJob is an example job that generates periodic reports
type ReportGenerationJob struct {
	*BaseJob
	reportType string
	outputPath string
	logger     logger.ILogger
}

// NewReportGenerationJob creates a new report generation job
func NewReportGenerationJob(reportType, outputPath string, logger logger.ILogger) *ReportGenerationJob {
	return &ReportGenerationJob{
		BaseJob:    NewBaseJob("report_generation"),
		reportType: reportType,
		outputPath: outputPath,
		logger:     logger,
	}
}

// Run executes the report generation job
func (r *ReportGenerationJob) Run(ctx context.Context) error {
	r.logger.Info("Starting report generation", "type", r.reportType, "output", r.outputPath)

	// Simulate report generation work
	steps := []string{"Collecting data", "Processing records", "Generating charts", "Creating PDF"}

	for i, step := range steps {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			r.logger.Info("Report generation step", "step", i+1, "description", step)
			time.Sleep(200 * time.Millisecond)
		}
	}

	// In a real implementation, you would:
	// 1. Query database for report data
	// 2. Process and aggregate the data
	// 3. Generate charts/graphs
	// 4. Create PDF or other output format
	// 5. Save to file system or send via email

	r.logger.Info("Report generation completed successfully", "output_file", r.outputPath)
	return nil
}

// ExampleJobFactory demonstrates how to create job factories for different job types
type ExampleJobFactory struct {
	logger logger.ILogger
}

// NewExampleJobFactory creates a new example job factory
func NewExampleJobFactory(logger logger.ILogger) *ExampleJobFactory {
	return &ExampleJobFactory{logger: logger}
}

// CreateDatabaseCleanupJob creates a database cleanup job factory
func (f *ExampleJobFactory) CreateDatabaseCleanupJob(retentionDays int) JobFactory {
	return func() Job {
		return NewDatabaseCleanupJob(retentionDays, f.logger)
	}
}

// CreateCacheWarmupJob creates a cache warmup job factory
func (f *ExampleJobFactory) CreateCacheWarmupJob(cacheKeys []string) JobFactory {
	return func() Job {
		return NewCacheWarmupJob(cacheKeys, f.logger)
	}
}

// CreateHealthCheckJob creates a health check job factory
func (f *ExampleJobFactory) CreateHealthCheckJob(endpoints []string) JobFactory {
	return func() Job {
		return NewHealthCheckJob(endpoints, f.logger)
	}
}

// CreateReportGenerationJob creates a report generation job factory
func (f *ExampleJobFactory) CreateReportGenerationJob(reportType, outputPath string) JobFactory {
	return func() Job {
		return NewReportGenerationJob(reportType, outputPath, f.logger)
	}
}

// SetupExampleJobs demonstrates how to set up a complete cron system with example jobs
func SetupExampleJobs(logger logger.ILogger) (*CronManager, *JobRegistry) {
	// Create cron manager
	manager := NewCronManager(logger)

	// Create job factory
	factory := NewExampleJobFactory(logger)

	// Create and configure job registry
	registry := NewJobRegistryBuilder(logger).
		WithJob("database_cleanup", factory.CreateDatabaseCleanupJob(30)).
		WithJob("cache_warmup", factory.CreateCacheWarmupJob([]string{"users", "products", "categories"})).
		WithJob("health_check", factory.CreateHealthCheckJob([]string{"http://api.example.com/health", "http://db.example.com/ping"})).
		WithJob("daily_report", factory.CreateReportGenerationJob("daily", "/tmp/daily_report.pdf")).
		WithConfig("database_cleanup", "0 2 * * *", true). // Daily at 2 AM
		WithConfig("cache_warmup", "*/15 * * * *", true).  // Every 15 minutes
		WithConfig("health_check", "*/5 * * * *", true).   // Every 5 minutes
		WithConfig("daily_report", "0 9 * * 1-5", true).   // Weekdays at 9 AM
		Build()

	// Register jobs with manager
	if err := registry.RegisterJobsWithManager(manager); err != nil {
		logger.Error("Failed to register jobs with manager", "error", err.Error())
	}

	return manager, registry
}
