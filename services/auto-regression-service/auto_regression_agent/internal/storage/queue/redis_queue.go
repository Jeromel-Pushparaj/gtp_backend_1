package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// RedisQueue implements job queue using Redis Streams
type RedisQueue struct {
	client        *redis.Client
	streamName    string
	consumerGroup string
	maxLen        int64
}

// NewRedisQueue creates a new Redis queue
func NewRedisQueue(client *redis.Client, streamName, consumerGroup string, maxLen int64) *RedisQueue {
	return &RedisQueue{
		client:        client,
		streamName:    streamName,
		consumerGroup: consumerGroup,
		maxLen:        maxLen,
	}
}

// Initialize initializes the queue (creates consumer group)
func (q *RedisQueue) Initialize(ctx context.Context) error {
	// Create consumer group if it doesn't exist
	err := q.client.XGroupCreateMkStream(ctx, q.streamName, q.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to create consumer group: %w", err)
	}
	return nil
}

// Enqueue adds a job to the queue
func (q *RedisQueue) Enqueue(ctx context.Context, job *workflow.Job) error {
	// Ensure consumer group exists (in case Redis was flushed)
	err := q.client.XGroupCreateMkStream(ctx, q.streamName, q.consumerGroup, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return fmt.Errorf("failed to ensure consumer group: %w", err)
	}

	// Serialize job to JSON
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to stream - let Redis auto-generate ID for monotonic ordering
	// Priority is stored in the job data and can be used for processing order
	args := &redis.XAddArgs{
		Stream: q.streamName,
		MaxLen: q.maxLen,
		Approx: true,
		Values: map[string]interface{}{
			"job_id":   job.ID,
			"type":     job.Type,
			"priority": job.Priority,
			"team_id":  job.TeamID,
			"data":     string(data),
		},
	}

	if err := q.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("failed to add job to queue: %w", err)
	}

	// Update job status
	job.Status = workflow.JobStatusQueued
	now := time.Now()
	job.QueuedAt = &now
	job.UpdatedAt = now

	return nil
}

// EnqueueBatch adds multiple jobs to the queue
func (q *RedisQueue) EnqueueBatch(ctx context.Context, jobs []*workflow.Job) error {
	// Use pipeline for batch operations
	pipe := q.client.Pipeline()

	for _, job := range jobs {
		data, err := json.Marshal(job)
		if err != nil {
			return fmt.Errorf("failed to marshal job %s: %w", job.ID, err)
		}

		// Let Redis auto-generate IDs for monotonic ordering
		// Priority is stored in the job data
		args := &redis.XAddArgs{
			Stream: q.streamName,
			MaxLen: q.maxLen,
			Approx: true,
			Values: map[string]interface{}{
				"job_id":   job.ID,
				"type":     job.Type,
				"priority": job.Priority,
				"team_id":  job.TeamID,
				"data":     string(data),
			},
		}

		pipe.XAdd(ctx, args)

		// Update job status
		job.Status = workflow.JobStatusQueued
		now := time.Now()
		job.QueuedAt = &now
		job.UpdatedAt = now
	}

	// Execute pipeline
	if _, err := pipe.Exec(ctx); err != nil {
		return fmt.Errorf("failed to enqueue batch: %w", err)
	}

	return nil
}

// Dequeue retrieves jobs from the queue
func (q *RedisQueue) Dequeue(ctx context.Context, consumerID string, count int, blockTime time.Duration) ([]*workflow.Job, error) {
	// Read from stream
	streams, err := q.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    q.consumerGroup,
		Consumer: consumerID,
		Streams:  []string{q.streamName, ">"},
		Count:    int64(count),
		Block:    blockTime,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return []*workflow.Job{}, nil // No messages available
		}
		// If consumer group doesn't exist, try to create it
		if err.Error() == "NOGROUP No such key '"+q.streamName+"' or consumer group '"+q.consumerGroup+"' in XREADGROUP with GROUP option" {
			if initErr := q.Initialize(ctx); initErr != nil {
				return nil, fmt.Errorf("failed to initialize consumer group: %w", initErr)
			}
			return []*workflow.Job{}, nil // Return empty, will retry on next dequeue
		}
		return nil, fmt.Errorf("failed to read from queue: %w", err)
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return []*workflow.Job{}, nil
	}

	// Parse all messages
	jobs := make([]*workflow.Job, 0, len(streams[0].Messages))
	for _, msg := range streams[0].Messages {
		// Parse job data
		jobData, ok := msg.Values["data"].(string)
		if !ok {
			continue
		}

		var job workflow.Job
		if err := json.Unmarshal([]byte(jobData), &job); err != nil {
			continue
		}

		// Store message ID for acknowledgment
		if job.Payload == nil {
			job.Payload = make(map[string]interface{})
		}
		job.Payload["_stream_id"] = msg.ID

		jobs = append(jobs, &job)
	}

	return jobs, nil
}

// Acknowledge acknowledges a job (removes from pending)
func (q *RedisQueue) Acknowledge(ctx context.Context, job *workflow.Job) error {
	streamID, ok := job.Payload["_stream_id"].(string)
	if !ok {
		return fmt.Errorf("stream ID not found in job payload")
	}

	if err := q.client.XAck(ctx, q.streamName, q.consumerGroup, streamID).Err(); err != nil {
		return fmt.Errorf("failed to acknowledge job: %w", err)
	}

	return nil
}

// GetQueueLength returns the current queue length
func (q *RedisQueue) GetQueueLength(ctx context.Context) (int64, error) {
	length, err := q.client.XLen(ctx, q.streamName).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get queue length: %w", err)
	}
	return length, nil
}

// GetPendingCount returns the count of pending jobs
func (q *RedisQueue) GetPendingCount(ctx context.Context) (int64, error) {
	pending, err := q.client.XPending(ctx, q.streamName, q.consumerGroup).Result()
	if err != nil {
		return 0, fmt.Errorf("failed to get pending count: %w", err)
	}
	return pending.Count, nil
}

// GetPendingJobs returns the number of pending jobs (alias for GetPendingCount)
func (q *RedisQueue) GetPendingJobs(ctx context.Context, consumerGroup string) (int64, error) {
	return q.GetPendingCount(ctx)
}

// Close closes the Redis client connection
func (q *RedisQueue) Close() error {
	if q.client != nil {
		return q.client.Close()
	}
	return nil
}

// Health checks the Redis connection health
func (q *RedisQueue) Health(ctx context.Context) error {
	return q.client.Ping(ctx).Err()
}
