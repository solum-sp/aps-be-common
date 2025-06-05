package cron

import (
	"context"
	"fmt"
	"time"
)

// JobStatus represents the execution status of a cron job
type JobStatus string

const (
	JobStatusIdle    JobStatus = "idle"
	JobStatusRunning JobStatus = "running"
	JobStatusFailed  JobStatus = "failed"
	JobStatusSuccess JobStatus = "success"
)

// JobInfo contains metadata about a cron job
type JobInfo struct {
	Name         string         `json:"name"`
	Schedule     string         `json:"schedule"`
	Status       JobStatus      `json:"status"`
	LastRun      *time.Time     `json:"last_run,omitempty"`
	LastDuration *time.Duration `json:"last_duration,omitempty"`
	LastError    string         `json:"last_error,omitempty"`
	RunCount     int64          `json:"run_count"`
	FailCount    int64          `json:"fail_count"`
}

// JobWithMetadata wraps a Job with execution metadata
type JobWithMetadata struct {
	Job
	info JobInfo
}

// BaseJob provides a basic implementation that other jobs can embed
type BaseJob struct {
	name string
}

// NewBaseJob creates a new BaseJob with the given name
func NewBaseJob(name string) *BaseJob {
	return &BaseJob{name: name}
}

// GetName returns the job name
func (b *BaseJob) GetName() string {
	return b.name
}

// FuncJob wraps a function to implement the Job interface
type FuncJob struct {
	*BaseJob
	fn func(ctx context.Context) error
}

// NewFuncJob creates a new job from a function
func NewFuncJob(name string, fn func(ctx context.Context) error) *FuncJob {
	return &FuncJob{
		BaseJob: NewBaseJob(name),
		fn:      fn,
	}
}

// Run executes the wrapped function
func (f *FuncJob) Run(ctx context.Context) error {
	return f.fn(ctx)
}

// PeriodicJob represents a job that runs at specific intervals
type PeriodicJob struct {
	*BaseJob
	interval time.Duration
	fn       func(ctx context.Context) error
}

// NewPeriodicJob creates a new periodic job
func NewPeriodicJob(name string, interval time.Duration, fn func(ctx context.Context) error) *PeriodicJob {
	return &PeriodicJob{
		BaseJob:  NewBaseJob(name),
		interval: interval,
		fn:       fn,
	}
}

// Run executes the periodic job function
func (p *PeriodicJob) Run(ctx context.Context) error {
	return p.fn(ctx)
}

// GetSchedule returns the cron schedule for this periodic job
func (p *PeriodicJob) GetSchedule() string {
	// Convert duration to cron expression (6-field format with seconds)
	if p.interval >= time.Hour {
		hours := int(p.interval.Hours())
		return fmt.Sprintf("0 0 */%d * * *", hours)
	} else if p.interval >= time.Minute {
		minutes := int(p.interval.Minutes())
		return fmt.Sprintf("0 */%d * * * *", minutes)
	}
	// For sub-minute intervals, use seconds
	seconds := int(p.interval.Seconds())
	if seconds < 1 {
		seconds = 1 // Minimum 1 second
	}
	return fmt.Sprintf("*/%d * * * * *", seconds)
}
