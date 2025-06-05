package cron

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"

	"github.com/solum-sp/aps-be-common/common/logger"
)

// JobFactory is a function that creates a new job instance
type JobFactory func() Job

// JobRegistry manages job factories and configurations
type JobRegistry struct {
	factories map[string]JobFactory
	configs   map[string]JobConfig
	mu        sync.RWMutex
	logger    logger.ILogger
}

// NewJobRegistry creates a new job registry
func NewJobRegistry(logger logger.ILogger) *JobRegistry {
	return &JobRegistry{
		factories: make(map[string]JobFactory),
		configs:   make(map[string]JobConfig),
		logger:    logger,
	}
}

// RegisterJob registers a job factory with the registry
func (jr *JobRegistry) RegisterJob(name string, factory JobFactory) {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	jr.factories[name] = factory
	jr.logger.Info("Registered job factory", "job", name)
}

// LoadConfigFromFile loads job configurations from a JSON file
func (jr *JobRegistry) LoadConfigFromFile(configPath string) error {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var configs []JobConfig
	if err := json.Unmarshal(data, &configs); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	jr.mu.Lock()
	defer jr.mu.Unlock()

	for _, config := range configs {
		jr.configs[config.Name] = config
		jr.logger.Info("Loaded job config", "job", config.Name, "schedule", config.Schedule, "enabled", config.Enabled)
	}

	return nil
}

// SetJobConfig sets configuration for a specific job
func (jr *JobRegistry) SetJobConfig(name string, config JobConfig) {
	jr.mu.Lock()
	defer jr.mu.Unlock()

	jr.configs[name] = config
}

// GetJobConfig returns the configuration for a specific job
func (jr *JobRegistry) GetJobConfig(name string) (JobConfig, bool) {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	config, exists := jr.configs[name]
	return config, exists
}

// CreateJob creates a job instance using the registered factory
func (jr *JobRegistry) CreateJob(name string) (Job, error) {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	factory, exists := jr.factories[name]
	if !exists {
		return nil, fmt.Errorf("no factory registered for job: %s", name)
	}

	return factory(), nil
}

// GetRegisteredJobs returns all registered job names
func (jr *JobRegistry) GetRegisteredJobs() []string {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	names := make([]string, 0, len(jr.factories))
	for name := range jr.factories {
		names = append(names, name)
	}
	return names
}

// RegisterJobsWithManager registers all configured jobs with the cron manager
func (jr *JobRegistry) RegisterJobsWithManager(manager *CronManager) error {
	jr.mu.RLock()
	defer jr.mu.RUnlock()

	for name, config := range jr.configs {
		if !config.Enabled {
			jr.logger.Info("Skipping disabled job", "job", name)
			continue
		}

		job, err := jr.CreateJob(name)
		if err != nil {
			jr.logger.Error("Failed to create job", "job", name, "error", err.Error())
			continue
		}

		if err := manager.AddJob(config.Schedule, job); err != nil {
			jr.logger.Error("Failed to add job to manager", "job", name, "error", err.Error())
			continue
		}

		jr.logger.Info("Successfully registered job with manager", "job", name)
	}

	return nil
}

// JobRegistryBuilder provides a fluent interface for building job registries
type JobRegistryBuilder struct {
	registry *JobRegistry
}

// NewJobRegistryBuilder creates a new job registry builder
func NewJobRegistryBuilder(logger logger.ILogger) *JobRegistryBuilder {
	return &JobRegistryBuilder{
		registry: NewJobRegistry(logger),
	}
}

// WithJob adds a job factory to the registry
func (jrb *JobRegistryBuilder) WithJob(name string, factory JobFactory) *JobRegistryBuilder {
	jrb.registry.RegisterJob(name, factory)
	return jrb
}

// WithConfig adds a job configuration
func (jrb *JobRegistryBuilder) WithConfig(name, schedule string, enabled bool) *JobRegistryBuilder {
	jrb.registry.SetJobConfig(name, JobConfig{
		Name:     name,
		Schedule: schedule,
		Enabled:  enabled,
	})
	return jrb
}

// WithConfigFile loads configurations from a file
func (jrb *JobRegistryBuilder) WithConfigFile(configPath string) *JobRegistryBuilder {
	if err := jrb.registry.LoadConfigFromFile(configPath); err != nil {
		jrb.registry.logger.Error("Failed to load config file", "path", configPath, "error", err.Error())
	}
	return jrb
}

// Build returns the configured job registry
func (jrb *JobRegistryBuilder) Build() *JobRegistry {
	return jrb.registry
}

// DefaultJobRegistry provides common job implementations
type DefaultJobRegistry struct {
	*JobRegistry
}

// NewDefaultJobRegistry creates a registry with some common jobs pre-registered
func NewDefaultJobRegistry(logger logger.ILogger) *DefaultJobRegistry {
	registry := NewJobRegistry(logger)

	// Register some common job types
	registry.RegisterJob("heartbeat", func() Job {
		return NewFuncJob("heartbeat", func(ctx context.Context) error {
			logger.Info("Heartbeat job executed")
			return nil
		})
	})

	registry.RegisterJob("cleanup", func() Job {
		return NewFuncJob("cleanup", func(ctx context.Context) error {
			logger.Info("Cleanup job executed")
			// Add your cleanup logic here
			return nil
		})
	})

	return &DefaultJobRegistry{JobRegistry: registry}
}
