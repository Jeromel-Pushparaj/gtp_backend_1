package queue

import (
	"time"
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusQueued     JobStatus = "queued"
	JobStatusProcessing JobStatus = "processing"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
	JobStatusDeadLetter JobStatus = "dead_letter"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeDiscovery        JobType = "discovery"
	JobTypePayloadGen       JobType = "payload_generation"
	JobTypeDataSeeding      JobType = "data_seeding"
	JobTypeManifestCreation JobType = "manifest_creation"
	JobTypeExecution        JobType = "execution"
	JobTypeHealing          JobType = "healing"
)

// Job represents a queue job
type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	Priority    int                    `json:"priority"`
	TraceID     string                 `json:"trace_id"`
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	RetryCount  int                    `json:"retry_count"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	QueuedAt    *time.Time             `json:"queued_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	FailedAt    *time.Time             `json:"failed_at,omitempty"`
	Metadata    JobMetadata            `json:"metadata"`
}

// JobMetadata contains additional job metadata
type JobMetadata struct {
	SpecID      string            `json:"spec_id,omitempty"`
	EndpointID  string            `json:"endpoint_id,omitempty"`
	WorkflowID  string            `json:"workflow_id,omitempty"`
	TeamID      string            `json:"team_id,omitempty"`
	ServiceName string            `json:"service_name,omitempty"`
	RunMode     string            `json:"run_mode,omitempty"`
	Tags        map[string]string `json:"tags,omitempty"`
}

// JobResult represents the result of job processing
type JobResult struct {
	JobID        string                 `json:"job_id"`
	Success      bool                   `json:"success"`
	Output       map[string]interface{} `json:"output,omitempty"`
	Error        string                 `json:"error,omitempty"`
	ManifestPath string                 `json:"manifest_path,omitempty"`
	Duration     time.Duration          `json:"duration"`
	ProcessedAt  time.Time              `json:"processed_at"`
}

// WorkerStats represents worker statistics
type WorkerStats struct {
	WorkerID       string        `json:"worker_id"`
	Status         string        `json:"status"` // idle, busy, stopped
	JobsProcessed  int64         `json:"jobs_processed"`
	JobsSucceeded  int64         `json:"jobs_succeeded"`
	JobsFailed     int64         `json:"jobs_failed"`
	CurrentJob     *Job          `json:"current_job,omitempty"`
	Uptime         time.Duration `json:"uptime"`
	LastHeartbeat  time.Time     `json:"last_heartbeat"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	PendingJobs     int64              `json:"pending_jobs"`
	ProcessingJobs  int64              `json:"processing_jobs"`
	CompletedJobs   int64              `json:"completed_jobs"`
	FailedJobs      int64              `json:"failed_jobs"`
	DeadLetterJobs  int64              `json:"dead_letter_jobs"`
	ActiveWorkers   int                `json:"active_workers"`
	WorkerStats     []WorkerStats      `json:"worker_stats,omitempty"`
	JobsByType      map[JobType]int64  `json:"jobs_by_type"`
	JobsByPriority  map[int]int64      `json:"jobs_by_priority"`
	AverageWaitTime time.Duration      `json:"average_wait_time"`
	Timestamp       time.Time          `json:"timestamp"`
}

// DeadLetterEntry represents a dead letter queue entry
type DeadLetterEntry struct {
	Job           Job       `json:"job"`
	FailureReason string    `json:"failure_reason"`
	FailureCount  int       `json:"failure_count"`
	LastError     string    `json:"last_error"`
	MovedAt       time.Time `json:"moved_at"`
}

