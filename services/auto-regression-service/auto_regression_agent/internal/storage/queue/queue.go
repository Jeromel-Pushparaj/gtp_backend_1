package queue

import (
	"context"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// Queue defines the interface for job queue operations
type Queue interface {
	// Initialize initializes the queue
	Initialize(ctx context.Context) error

	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *workflow.Job) error

	// EnqueueBatch adds multiple jobs to the queue
	EnqueueBatch(ctx context.Context, jobs []*workflow.Job) error

	// Dequeue retrieves jobs from the queue
	Dequeue(ctx context.Context, consumerID string, count int, blockTime time.Duration) ([]*workflow.Job, error)

	// Acknowledge marks a job as completed
	Acknowledge(ctx context.Context, job *workflow.Job) error

	// GetQueueLength returns the number of jobs in the queue
	GetQueueLength(ctx context.Context) (int64, error)

	// GetPendingJobs returns the number of pending jobs
	GetPendingJobs(ctx context.Context, consumerGroup string) (int64, error)

	// Close closes the queue
	Close() error

	// Health checks the queue health
	Health(ctx context.Context) error
}
