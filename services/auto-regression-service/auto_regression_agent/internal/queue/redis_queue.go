package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/queue"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	// Queue keys
	queueKeyPrefix      = "queue:"
	processingKeyPrefix = "processing:"
	deadLetterKeyPrefix = "dlq:"
	jobKeyPrefix        = "job:"

	// Stream names
	mainQueueStream  = "queue:main"
	deadLetterStream = "queue:dlq"

	// Consumer group
	consumerGroup = "workers"
)

// RedisQueue implements a Redis-based job queue
type RedisQueue struct {
	client            *redis.Client
	maxRetries        int
	retryDelay        time.Duration
	visibilityTimeout time.Duration
}

// Config represents queue configuration
type Config struct {
	MaxRetries        int
	RetryDelay        time.Duration
	VisibilityTimeout time.Duration
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(client *redis.Client, config Config) *RedisQueue {
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryDelay == 0 {
		config.RetryDelay = 5 * time.Second
	}
	if config.VisibilityTimeout == 0 {
		config.VisibilityTimeout = 5 * time.Minute
	}

	return &RedisQueue{
		client:            client,
		maxRetries:        config.MaxRetries,
		retryDelay:        config.RetryDelay,
		visibilityTimeout: config.VisibilityTimeout,
	}
}

// Initialize initializes the queue (creates consumer group)
func (q *RedisQueue) Initialize(ctx context.Context) error {
	// Create main queue stream and consumer group
	err := q.client.XGroupCreateMkStream(ctx, mainQueueStream, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}

	// Create dead letter queue stream
	err = q.client.XGroupCreateMkStream(ctx, deadLetterStream, consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create DLQ consumer group: %w", err)
	}

	return nil
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *queue.Job) error {
	// Set defaults
	if job.ID == "" {
		job.ID = uuid.New().String()
	}
	if job.TraceID == "" {
		job.TraceID = uuid.New().String()
	}
	if job.CreatedAt.IsZero() {
		job.CreatedAt = time.Now()
	}
	if job.MaxRetries == 0 {
		job.MaxRetries = q.maxRetries
	}
	job.Status = queue.JobStatusQueued
	now := time.Now()
	job.QueuedAt = &now

	// Serialize job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Store job data
	jobKey := jobKeyPrefix + job.ID
	if err := q.client.Set(ctx, jobKey, jobData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store job: %w", err)
	}

	// Add to stream with priority
	streamData := map[string]interface{}{
		"job_id":   job.ID,
		"type":     string(job.Type),
		"priority": job.Priority,
		"trace_id": job.TraceID,
	}

	if _, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: mainQueueStream,
		Values: streamData,
	}).Result(); err != nil {
		return fmt.Errorf("failed to add job to stream: %w", err)
	}

	return nil
}

// Dequeue retrieves a job from the queue for processing
func (q *RedisQueue) Dequeue(ctx context.Context, workerID string) (*queue.Job, error) {
	// Read from stream
	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    consumerGroup,
		Consumer: workerID,
		Streams:  []string{mainQueueStream, ">"},
		Count:    1,
		Block:    5 * time.Second,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, nil // No jobs available
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, nil // No jobs available
	}

	msg := streams[0].Messages[0]
	jobID := msg.Values["job_id"].(string)

	// Retrieve job data
	jobKey := jobKeyPrefix + jobID
	jobData, err := q.client.Get(ctx, jobKey).Result()
	if err != nil {
		// Acknowledge message even if job not found
		q.client.XAck(ctx, mainQueueStream, consumerGroup, msg.ID)
		return nil, fmt.Errorf("failed to get job data: %w", err)
	}

	var job queue.Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		q.client.XAck(ctx, mainQueueStream, consumerGroup, msg.ID)
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Update job status
	job.Status = queue.JobStatusProcessing
	now := time.Now()
	job.StartedAt = &now

	// Store message ID for acknowledgment
	job.Metadata.Tags = map[string]string{
		"message_id": msg.ID,
		"worker_id":  workerID,
	}

	return &job, nil
}

// Acknowledge marks a job as successfully completed
func (q *RedisQueue) Acknowledge(ctx context.Context, job *queue.Job) error {
	// Update job status
	job.Status = queue.JobStatusCompleted
	now := time.Now()
	job.CompletedAt = &now

	// Store updated job
	jobData, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	jobKey := jobKeyPrefix + job.ID
	if err := q.client.Set(ctx, jobKey, jobData, 24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	// Acknowledge message in stream
	messageID := job.Metadata.Tags["message_id"]
	if messageID != "" {
		if err := q.client.XAck(ctx, mainQueueStream, consumerGroup, messageID).Err(); err != nil {
			return fmt.Errorf("failed to ack message: %w", err)
		}
	}

	return nil
}

// Fail marks a job as failed and handles retry logic
func (q *RedisQueue) Fail(ctx context.Context, job *queue.Job, errMsg string) error {
	job.Error = errMsg
	job.RetryCount++

	// Check if we should retry
	if job.RetryCount < job.MaxRetries {
		job.Status = queue.JobStatusRetrying

		// Update job
		jobData, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal job: %w", err)
		}

		jobKey := jobKeyPrefix + job.ID
		if err := q.client.Set(ctx, jobKey, jobData, 24*time.Hour).Err(); err != nil {
			return fmt.Errorf("failed to update job: %w", err)
		}

		// Re-enqueue with delay
		time.Sleep(q.retryDelay)

		streamData := map[string]interface{}{
			"job_id":   job.ID,
			"type":     string(job.Type),
			"priority": job.Priority,
			"trace_id": job.TraceID,
			"retry":    job.RetryCount,
		}

		if _, err := q.client.XAdd(ctx, &redis.XAddArgs{
			Stream: mainQueueStream,
			Values: streamData,
		}).Result(); err != nil {
			return fmt.Errorf("failed to re-enqueue job: %w", err)
		}

		// Acknowledge original message
		messageID := job.Metadata.Tags["message_id"]
		if messageID != "" {
			q.client.XAck(ctx, mainQueueStream, consumerGroup, messageID)
		}

		return nil
	}

	// Move to dead letter queue
	return q.moveToDeadLetter(ctx, job, errMsg)
}

// moveToDeadLetter moves a job to the dead letter queue
func (q *RedisQueue) moveToDeadLetter(ctx context.Context, job *queue.Job, reason string) error {
	job.Status = queue.JobStatusDeadLetter
	now := time.Now()
	job.FailedAt = &now

	// Create dead letter entry
	dlEntry := queue.DeadLetterEntry{
		Job:           *job,
		FailureReason: reason,
		FailureCount:  job.RetryCount,
		LastError:     job.Error,
		MovedAt:       now,
	}

	// Serialize entry
	entryData, err := json.Marshal(dlEntry)
	if err != nil {
		return fmt.Errorf("failed to marshal DL entry: %w", err)
	}

	// Store in dead letter queue
	dlKey := deadLetterKeyPrefix + job.ID
	if err := q.client.Set(ctx, dlKey, entryData, 7*24*time.Hour).Err(); err != nil {
		return fmt.Errorf("failed to store DL entry: %w", err)
	}

	// Add to dead letter stream
	streamData := map[string]interface{}{
		"job_id":   job.ID,
		"type":     string(job.Type),
		"reason":   reason,
		"trace_id": job.TraceID,
	}

	if _, err := q.client.XAdd(ctx, &redis.XAddArgs{
		Stream: deadLetterStream,
		Values: streamData,
	}).Result(); err != nil {
		return fmt.Errorf("failed to add to DL stream: %w", err)
	}

	// Acknowledge original message
	messageID := job.Metadata.Tags["message_id"]
	if messageID != "" {
		q.client.XAck(ctx, mainQueueStream, consumerGroup, messageID)
	}

	return nil
}

// GetStats returns queue statistics
func (q *RedisQueue) GetStats(ctx context.Context) (*queue.QueueStats, error) {
	stats := &queue.QueueStats{
		JobsByType:     make(map[queue.JobType]int64),
		JobsByPriority: make(map[int]int64),
		Timestamp:      time.Now(),
	}

	// Get stream info
	info, err := q.client.XInfoStream(ctx, mainQueueStream).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get stream info: %w", err)
	}

	stats.PendingJobs = info.Length

	// Get consumer group info
	groups, err := q.client.XInfoGroups(ctx, mainQueueStream).Result()
	if err == nil && len(groups) > 0 {
		stats.ProcessingJobs = groups[0].Pending
	}

	// Get dead letter queue size
	dlInfo, err := q.client.XInfoStream(ctx, deadLetterStream).Result()
	if err == nil {
		stats.DeadLetterJobs = dlInfo.Length
	}

	return stats, nil
}

// GetJob retrieves a job by ID
func (q *RedisQueue) GetJob(ctx context.Context, jobID string) (*queue.Job, error) {
	jobKey := jobKeyPrefix + jobID
	jobData, err := q.client.Get(ctx, jobKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("job not found: %s", jobID)
		}
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	var job queue.Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	return &job, nil
}
