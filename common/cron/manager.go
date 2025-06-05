package cron

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
	"github.com/solum-sp/aps-be-common/common/logger"
)

type CronManager struct {
	cron      *cron.Cron
	logger    logger.ILogger
	jobs      map[string]cron.EntryID
	jobInfo   map[string]*JobInfo
	jobsMu    sync.RWMutex
	isRunning bool
	mu        sync.RWMutex
}

type Job interface {
	Run(ctx context.Context) error
	GetName() string
}

// JobConfig represents configuration for a cron job
type JobConfig struct {
	Name     string `json:"name"`
	Schedule string `json:"schedule"`
	Enabled  bool   `json:"enabled"`
}

func NewCronManager(logger logger.ILogger) *CronManager {
	return &CronManager{
		cron:    cron.New(cron.WithSeconds()), // Enable seconds precision
		logger:  logger,
		jobs:    make(map[string]cron.EntryID),
		jobInfo: make(map[string]*JobInfo),
	}
}

func (cm *CronManager) AddJob(schedule string, job Job) error {
	cm.jobsMu.Lock()
	defer cm.jobsMu.Unlock()

	jobName := job.GetName()

	// Remove existing job if present
	if entryID, exists := cm.jobs[jobName]; exists {
		cm.cron.Remove(entryID)
		cm.logger.Info("Removed existing cron job", "job", jobName)
	}

	// Initialize job info if not exists
	if _, exists := cm.jobInfo[jobName]; !exists {
		cm.jobInfo[jobName] = &JobInfo{
			Name:     jobName,
			Schedule: schedule,
			Status:   JobStatusIdle,
		}
	} else {
		cm.jobInfo[jobName].Schedule = schedule
	}

	// Add new job with context and metadata tracking
	entryID, err := cm.cron.AddFunc(schedule, func() {
		cm.executeJob(job)
	})

	if err != nil {
		cm.logger.Error("Failed to add cron job", "job", jobName, "error", err.Error())
		return err
	}

	cm.jobs[jobName] = entryID
	cm.logger.Info("Added cron job", "job", jobName, "schedule", schedule)
	return nil
}

// executeJob wraps job execution with metadata tracking
func (cm *CronManager) executeJob(job Job) {
	jobName := job.GetName()
	ctx := context.Background()

	cm.jobsMu.Lock()
	if info, exists := cm.jobInfo[jobName]; exists {
		info.Status = JobStatusRunning
		now := time.Now()
		info.LastRun = &now
		info.RunCount++
	}
	cm.jobsMu.Unlock()

	cm.logger.Info("Starting cron job", "job", jobName)
	startTime := time.Now()

	err := job.Run(ctx)
	duration := time.Since(startTime)

	cm.jobsMu.Lock()
	if info, exists := cm.jobInfo[jobName]; exists {
		info.LastDuration = &duration
		if err != nil {
			info.Status = JobStatusFailed
			info.LastError = err.Error()
			info.FailCount++
			cm.logger.Error("Cron job failed", "job", jobName, "error", err.Error(), "duration", duration)
		} else {
			info.Status = JobStatusSuccess
			info.LastError = ""
			cm.logger.Info("Cron job completed successfully", "job", jobName, "duration", duration)
		}
	}
	cm.jobsMu.Unlock()
}

// AddPeriodicJob adds a job that runs at regular intervals
func (cm *CronManager) AddPeriodicJob(interval time.Duration, job Job) error {
	// Convert interval to cron expression
	schedule := cm.intervalToCron(interval)
	return cm.AddJob(schedule, job)
}

// intervalToCron converts a time.Duration to a cron expression
func (cm *CronManager) intervalToCron(interval time.Duration) string {
	if interval >= time.Hour {
		hours := int(interval.Hours())
		return fmt.Sprintf("0 0 */%d * * *", hours)
	} else if interval >= time.Minute {
		minutes := int(interval.Minutes())
		return fmt.Sprintf("0 */%d * * * *", minutes)
	}
	// For sub-minute intervals, use seconds
	seconds := int(interval.Seconds())
	return fmt.Sprintf("*/%d * * * * *", seconds)
}

func (cm *CronManager) Start() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.isRunning {
		cm.logger.Warn("Cron manager is already running")
		return
	}

	cm.cron.Start()
	cm.isRunning = true
	cm.logger.Info("Cron manager started", "jobs_count", len(cm.jobs))
}

func (cm *CronManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if !cm.isRunning {
		cm.logger.Warn("Cron manager is not running")
		return
	}

	ctx := cm.cron.Stop()
	select {
	case <-ctx.Done():
		cm.logger.Info("Cron manager stopped gracefully")
	default:
		cm.logger.Info("Cron manager stopped")
	}

	cm.isRunning = false
}

func (cm *CronManager) IsRunning() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.isRunning
}

func (cm *CronManager) GetJobCount() int {
	cm.jobsMu.RLock()
	defer cm.jobsMu.RUnlock()
	return len(cm.jobs)
}

func (cm *CronManager) RemoveJob(jobName string) {
	cm.jobsMu.Lock()
	defer cm.jobsMu.Unlock()

	if entryID, exists := cm.jobs[jobName]; exists {
		cm.cron.Remove(entryID)
		delete(cm.jobs, jobName)
		delete(cm.jobInfo, jobName)
		cm.logger.Info("Removed cron job", "job", jobName)
	}
}

// GetJobInfo returns information about a specific job
func (cm *CronManager) GetJobInfo(jobName string) (*JobInfo, bool) {
	cm.jobsMu.RLock()
	defer cm.jobsMu.RUnlock()

	info, exists := cm.jobInfo[jobName]
	if !exists {
		return nil, false
	}

	// Return a copy to avoid race conditions
	infoCopy := *info
	return &infoCopy, true
}

// GetAllJobsInfo returns information about all jobs
func (cm *CronManager) GetAllJobsInfo() map[string]JobInfo {
	cm.jobsMu.RLock()
	defer cm.jobsMu.RUnlock()

	result := make(map[string]JobInfo)
	for name, info := range cm.jobInfo {
		result[name] = *info
	}
	return result
}

// GetRunningJobs returns a list of currently running job names
func (cm *CronManager) GetRunningJobs() []string {
	cm.jobsMu.RLock()
	defer cm.jobsMu.RUnlock()

	var running []string
	for name, info := range cm.jobInfo {
		if info.Status == JobStatusRunning {
			running = append(running, name)
		}
	}
	return running
}

// GetJobNames returns all registered job names
func (cm *CronManager) GetJobNames() []string {
	cm.jobsMu.RLock()
	defer cm.jobsMu.RUnlock()

	names := make([]string, 0, len(cm.jobs))
	for name := range cm.jobs {
		names = append(names, name)
	}
	return names
}
