package session

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/jeromelp/gtp_backend_1/services/chat-agent-service/client"
	"github.com/redis/go-redis/v9"
)

const (
	// Session key prefix in Redis
	sessionKeyPrefix = "chat:session:"
	
	// Maximum messages to store per session (10 turns = 20 messages)
	maxMessagesPerSession = 20
	
	// Session expiration time (30 minutes of inactivity)
	sessionTTL = 30 * time.Minute
)

// RedisSessionManager manages conversation history in Redis
type RedisSessionManager struct {
	client *redis.Client
	ctx    context.Context
}

// NewRedisSessionManager creates a new Redis session manager
func NewRedisSessionManager(host, port, password string) (*RedisSessionManager, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", host, port),
		Password: password,
		DB:       0, // Use default DB
	})

	ctx := context.Background()

	// Test connection
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	log.Printf("✅ Redis session manager connected to %s:%s", host, port)

	return &RedisSessionManager{
		client: rdb,
		ctx:    ctx,
	}, nil
}

// GetHistory retrieves conversation history for a session
func (m *RedisSessionManager) GetHistory(sessionID string) ([]client.ChatMessage, error) {
	key := sessionKeyPrefix + sessionID

	// Get messages from Redis
	data, err := m.client.Get(m.ctx, key).Result()
	if err == redis.Nil {
		// Session doesn't exist yet, return empty history
		return []client.ChatMessage{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session from Redis: %w", err)
	}

	// Deserialize messages
	var messages []client.ChatMessage
	if err := json.Unmarshal([]byte(data), &messages); err != nil {
		return nil, fmt.Errorf("failed to unmarshal session data: %w", err)
	}

	log.Printf("[SESSION] Retrieved %d messages for session %s", len(messages), sessionID)
	return messages, nil
}

// AddMessage adds a message to the session history
func (m *RedisSessionManager) AddMessage(sessionID string, message client.ChatMessage) error {
	// Get current history
	history, err := m.GetHistory(sessionID)
	if err != nil {
		return err
	}

	// Append new message
	history = append(history, message)

	// Trim to max size (keep most recent messages)
	if len(history) > maxMessagesPerSession {
		history = history[len(history)-maxMessagesPerSession:]
	}

	// Serialize and store
	data, err := json.Marshal(history)
	if err != nil {
		return fmt.Errorf("failed to marshal session data: %w", err)
	}

	key := sessionKeyPrefix + sessionID
	if err := m.client.Set(m.ctx, key, data, sessionTTL).Err(); err != nil {
		return fmt.Errorf("failed to store session in Redis: %w", err)
	}

	log.Printf("[SESSION] Stored message for session %s (total: %d messages)", sessionID, len(history))
	return nil
}

// DeleteSession removes a session from Redis
func (m *RedisSessionManager) DeleteSession(sessionID string) error {
	key := sessionKeyPrefix + sessionID
	return m.client.Del(m.ctx, key).Err()
}

// Close closes the Redis connection
func (m *RedisSessionManager) Close() error {
	return m.client.Close()
}