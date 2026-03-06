package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// Bus represents an event bus for agent communication
type Bus struct {
	client      *redis.Client
	subscribers map[EventType][]chan *Event
	mu          sync.RWMutex
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewBus creates a new event bus
func NewBus(client *redis.Client) *Bus {
	ctx, cancel := context.WithCancel(context.Background())
	return &Bus{
		client:      client,
		subscribers: make(map[EventType][]chan *Event),
		ctx:         ctx,
		cancel:      cancel,
	}
}

// Publish publishes an event to the bus
func (b *Bus) Publish(ctx context.Context, event *Event) error {
	// Set event ID if not set
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	
	// Set timestamp if not set
	if event.Timestamp.IsZero() {
		event.Timestamp = time.Now()
	}

	// Serialize event
	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	// Publish to Redis channel
	channel := fmt.Sprintf("events:%s", event.Type)
	if err := b.client.Publish(ctx, channel, data).Err(); err != nil {
		return fmt.Errorf("failed to publish event: %w", err)
	}

	// Also publish to general events channel for monitoring
	if err := b.client.Publish(ctx, "events:all", data).Err(); err != nil {
		log.Printf("Warning: failed to publish to events:all: %v", err)
	}

	log.Printf("📢 Event published: type=%s, source=%s, target=%s, workflow=%s", 
		event.Type, event.Source, event.Target, event.WorkflowID)

	return nil
}

// Subscribe subscribes to events of a specific type
func (b *Bus) Subscribe(ctx context.Context, eventType EventType, handler func(*Event) error) error {
	channel := fmt.Sprintf("events:%s", eventType)
	
	pubsub := b.client.Subscribe(ctx, channel)
	defer pubsub.Close()

	log.Printf("📡 Subscribed to events: type=%s, channel=%s", eventType, channel)

	// Wait for confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Start listening
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Printf("📡 Unsubscribed from events: type=%s", eventType)
			return ctx.Err()
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Parse event
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Error parsing event: %v", err)
				continue
			}

			// Handle event
			if err := handler(&event); err != nil {
				log.Printf("Error handling event %s: %v", event.ID, err)
			}
		}
	}
}

// SubscribeMultiple subscribes to multiple event types
func (b *Bus) SubscribeMultiple(ctx context.Context, eventTypes []EventType, handler func(*Event) error) error {
	channels := make([]string, len(eventTypes))
	for i, et := range eventTypes {
		channels[i] = fmt.Sprintf("events:%s", et)
	}

	pubsub := b.client.Subscribe(ctx, channels...)
	defer pubsub.Close()

	log.Printf("📡 Subscribed to multiple events: types=%v", eventTypes)

	// Wait for confirmation
	_, err := pubsub.Receive(ctx)
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	// Start listening
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			log.Printf("📡 Unsubscribed from multiple events")
			return ctx.Err()
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Parse event
			var event Event
			if err := json.Unmarshal([]byte(msg.Payload), &event); err != nil {
				log.Printf("Error parsing event: %v", err)
				continue
			}

			// Handle event
			if err := handler(&event); err != nil {
				log.Printf("Error handling event %s: %v", event.ID, err)
			}
		}
	}
}

// Close closes the event bus
func (b *Bus) Close() error {
	b.cancel()
	return nil
}

