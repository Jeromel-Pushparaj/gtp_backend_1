package progress

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// ProgressUpdate represents a progress update message
type ProgressUpdate struct {
	WorkflowID string                 `json:"workflow_id"`
	JobID      string                 `json:"job_id"`
	JobType    string                 `json:"job_type"`
	Status     string                 `json:"status"`
	Progress   int                    `json:"progress"`
	Phase      string                 `json:"phase"`
	Message    string                 `json:"message,omitempty"`
	Result     map[string]interface{} `json:"result,omitempty"` // Job result data
}

// LogMessage represents a log message
type LogMessage struct {
	WorkflowID string                 `json:"workflow_id"`
	JobID      string                 `json:"job_id"`
	Level      string                 `json:"level"` // "info", "warn", "error"
	Message    string                 `json:"message"`
	Agent      string                 `json:"agent,omitempty"`
	Details    map[string]interface{} `json:"details,omitempty"`
}

// Broadcaster publishes progress updates and logs via Redis Pub/Sub
type Broadcaster struct {
	client     *redis.Client
	channel    string
	logChannel string
}

// NewBroadcaster creates a new progress broadcaster
func NewBroadcaster(client *redis.Client, channel string) *Broadcaster {
	if channel == "" {
		channel = "workflow_progress"
	}
	return &Broadcaster{
		client:     client,
		channel:    channel,
		logChannel: "workflow_logs",
	}
}

// Publish publishes a progress update
func (b *Broadcaster) Publish(ctx context.Context, update ProgressUpdate) error {
	data, err := json.Marshal(update)
	if err != nil {
		return fmt.Errorf("failed to marshal progress update: %w", err)
	}

	if err := b.client.Publish(ctx, b.channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish progress update: %w", err)
	}

	return nil
}

// PublishLog publishes a log message
func (b *Broadcaster) PublishLog(ctx context.Context, logMsg LogMessage) error {
	data, err := json.Marshal(logMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal log message: %w", err)
	}

	if err := b.client.Publish(ctx, b.logChannel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish log message: %w", err)
	}

	return nil
}

// Subscribe subscribes to progress updates
func (b *Broadcaster) Subscribe(ctx context.Context, handler func(ProgressUpdate)) error {
	pubsub := b.client.Subscribe(ctx, b.channel)
	defer pubsub.Close()

	// Wait for confirmation that subscription is created
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	log.Printf("Subscribed to progress updates on channel: %s", b.channel)

	// Listen for messages
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg := <-ch:
			var update ProgressUpdate
			if err := json.Unmarshal([]byte(msg.Payload), &update); err != nil {
				log.Printf("Failed to unmarshal progress update: %v", err)
				continue
			}
			handler(update)
		}
	}
}
