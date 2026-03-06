package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// MessageBus handles agent-to-agent messaging
type MessageBus struct {
	client *redis.Client
}

// NewMessageBus creates a new message bus
func NewMessageBus(client *redis.Client) *MessageBus {
	return &MessageBus{
		client: client,
	}
}

// SendMessage sends a message to a specific agent
func (mb *MessageBus) SendMessage(ctx context.Context, msg *Message) error {
	// Set message ID if not set
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	
	// Set timestamp if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Push to agent's inbox (Redis list)
	inboxKey := fmt.Sprintf("agent:%s:inbox", msg.To)
	if err := mb.client.LPush(ctx, inboxKey, data).Err(); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	// Set expiry on inbox (7 days)
	mb.client.Expire(ctx, inboxKey, 7*24*time.Hour)

	log.Printf("📨 Message sent: from=%s, to=%s, type=%s, subject=%s", 
		msg.From, msg.To, msg.Type, msg.Subject)

	return nil
}

// BroadcastMessage broadcasts a message to all agents
func (mb *MessageBus) BroadcastMessage(ctx context.Context, msg *Message) error {
	// Set message ID if not set
	if msg.ID == "" {
		msg.ID = uuid.New().String()
	}
	
	// Set timestamp if not set
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}

	msg.Type = MessageTypeBroadcast

	// Serialize message
	data, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Publish to broadcast channel
	if err := mb.client.Publish(ctx, "agent:broadcast", data).Err(); err != nil {
		return fmt.Errorf("failed to broadcast message: %w", err)
	}

	log.Printf("📢 Message broadcast: from=%s, subject=%s", msg.From, msg.Subject)

	return nil
}

// ReceiveMessages receives messages from agent's inbox
func (mb *MessageBus) ReceiveMessages(ctx context.Context, agentName string, count int, timeout time.Duration) ([]*Message, error) {
	inboxKey := fmt.Sprintf("agent:%s:inbox", agentName)

	// Pop messages from inbox
	messages := make([]*Message, 0, count)
	for i := 0; i < count; i++ {
		result, err := mb.client.BRPop(ctx, timeout, inboxKey).Result()
		if err != nil {
			if err == redis.Nil {
				break // No more messages
			}
			return nil, fmt.Errorf("failed to receive message: %w", err)
		}

		if len(result) < 2 {
			continue
		}

		// Parse message
		var msg Message
		if err := json.Unmarshal([]byte(result[1]), &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			continue
		}

		messages = append(messages, &msg)
	}

	return messages, nil
}

// SubscribeToBroadcast subscribes to broadcast messages
func (mb *MessageBus) SubscribeToBroadcast(ctx context.Context, handler func(*Message) error) error {
	pubsub := mb.client.Subscribe(ctx, "agent:broadcast")
	defer pubsub.Close()

	log.Printf("📡 Subscribed to broadcast messages")

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
			log.Printf("📡 Unsubscribed from broadcast messages")
			return ctx.Err()
		case msg := <-ch:
			if msg == nil {
				continue
			}

			// Parse message
			var message Message
			if err := json.Unmarshal([]byte(msg.Payload), &message); err != nil {
				log.Printf("Error parsing broadcast message: %v", err)
				continue
			}

			// Handle message
			if err := handler(&message); err != nil {
				log.Printf("Error handling broadcast message %s: %v", message.ID, err)
			}
		}
	}
}

