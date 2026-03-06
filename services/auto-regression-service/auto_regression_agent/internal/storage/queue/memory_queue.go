package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// MemoryQueue implements an in-memory job queue for development
type MemoryQueue struct {
	jobs   []*workflow.Job
	mu     sync.RWMutex
	closed bool
}

// NewMemoryQueue creates a new in-memory queue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		jobs:   make([]*workflow.Job, 0),
		closed: false,
	}
}

// Initialize initializes the queue
func (q *MemoryQueue) Initialize(ctx context.Context) error {
	return nil
}

// Enqueue adds a job to the queue
func (q *MemoryQueue) Enqueue(ctx context.Context, job *workflow.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	// Update job status
	job.Status = workflow.JobStatusQueued
	now := time.Now()
	job.QueuedAt = &now
	job.UpdatedAt = now

	// Add to queue
	q.jobs = append(q.jobs, job)

	return nil
}

// EnqueueBatch adds multiple jobs to the queue
func (q *MemoryQueue) EnqueueBatch(ctx context.Context, jobs []*workflow.Job) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	now := time.Now()
	for _, job := range jobs {
		job.Status = workflow.JobStatusQueued
		job.QueuedAt = &now
		job.UpdatedAt = now
		q.jobs = append(q.jobs, job)
	}

	return nil
}

// Dequeue retrieves a job from the queue
func (q *MemoryQueue) Dequeue(ctx context.Context, consumerID string, count int, blockTime time.Duration) ([]*workflow.Job, error) {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return nil, fmt.Errorf("queue is closed")
	}

	// Return available jobs (up to count)
	if len(q.jobs) == 0 {
		return []*workflow.Job{}, nil
	}

	// Get jobs
	n := count
	if n > len(q.jobs) {
		n = len(q.jobs)
	}

	result := make([]*workflow.Job, n)
	copy(result, q.jobs[:n])

	// Remove from queue
	q.jobs = q.jobs[n:]

	return result, nil
}

// Acknowledge marks a job as completed
func (q *MemoryQueue) Acknowledge(ctx context.Context, job *workflow.Job) error {
	// In-memory queue doesn't need acknowledgment
	return nil
}

// GetQueueLength returns the number of jobs in the queue
func (q *MemoryQueue) GetQueueLength(ctx context.Context) (int64, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return int64(len(q.jobs)), nil
}

// GetPendingJobs returns the number of pending jobs
func (q *MemoryQueue) GetPendingJobs(ctx context.Context, consumerGroup string) (int64, error) {
	q.mu.RLock()
	defer q.mu.RUnlock()
	return int64(len(q.jobs)), nil
}

// Close closes the queue
func (q *MemoryQueue) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	q.closed = true
	q.jobs = nil
	return nil
}

// Health checks the queue health
func (q *MemoryQueue) Health(ctx context.Context) error {
	q.mu.RLock()
	defer q.mu.RUnlock()

	if q.closed {
		return fmt.Errorf("queue is closed")
	}

	return nil
}
