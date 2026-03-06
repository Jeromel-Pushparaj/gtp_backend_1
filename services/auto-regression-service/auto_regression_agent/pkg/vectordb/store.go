package vectordb

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
)

// Store provides vector database operations using PostgreSQL with pgvector
type Store struct {
	db              *sql.DB
	embeddingClient *EmbeddingClient
}

// StoreConfig configures the vector store
type StoreConfig struct {
	Host     string
	Port     int
	Database string
	User     string
	Password string
	SSLMode  string
}

// NewStore creates a new vector store
func NewStore(cfg StoreConfig, embClient *EmbeddingClient) (*Store, error) {
	if cfg.SSLMode == "" {
		cfg.SSLMode = "disable"
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.Database, cfg.User, cfg.Password, cfg.SSLMode,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Store{
		db:              db,
		embeddingClient: embClient,
	}, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// Ping checks database connectivity
func (s *Store) Ping(ctx context.Context) error {
	return s.db.PingContext(ctx)
}

// vectorToString converts a vector to PostgreSQL array format
func vectorToString(v Vector) string {
	if v == nil {
		return ""
	}
	parts := make([]string, len(v))
	for i, val := range v {
		parts[i] = fmt.Sprintf("%f", val)
	}
	return "[" + strings.Join(parts, ",") + "]"
}

// StoreLearning stores a new learning with auto-generated embedding
func (s *Store) StoreLearning(ctx context.Context, learning *Learning) error {
	// Generate embedding if not provided
	if learning.Embedding == nil && s.embeddingClient != nil {
		embedding, err := s.embeddingClient.Embed(ctx, learning.Content)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		learning.Embedding = embedding
	}

	// Generate UUID if not set
	if learning.ID == uuid.Nil {
		learning.ID = uuid.New()
	}

	// Marshal context to JSON
	contextJSON, err := json.Marshal(learning.Context)
	if err != nil {
		return fmt.Errorf("failed to marshal context: %w", err)
	}

	query := `
		INSERT INTO learnings (id, category, source_api, content, context, embedding, confidence)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id) DO UPDATE SET
			content = EXCLUDED.content,
			context = EXCLUDED.context,
			embedding = EXCLUDED.embedding,
			confidence = EXCLUDED.confidence,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = s.db.ExecContext(ctx, query,
		learning.ID,
		learning.Category,
		learning.SourceAPI,
		learning.Content,
		contextJSON,
		vectorToString(learning.Embedding),
		learning.Confidence,
	)
	if err != nil {
		return fmt.Errorf("failed to store learning: %w", err)
	}

	return nil
}

// StoreFailurePattern stores a failure pattern with auto-generated embedding
func (s *Store) StoreFailurePattern(ctx context.Context, pattern *FailurePattern) error {
	// Generate embedding from error signature + fix description
	if pattern.Embedding == nil && s.embeddingClient != nil {
		text := fmt.Sprintf("%s: %s. Fix: %s", pattern.FailureType, pattern.ErrorSignature, pattern.FixDescription)
		embedding, err := s.embeddingClient.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		pattern.Embedding = embedding
	}

	if pattern.ID == uuid.Nil {
		pattern.ID = uuid.New()
	}

	fixPayloadJSON, err := json.Marshal(pattern.FixPayload)
	if err != nil {
		return fmt.Errorf("failed to marshal fix payload: %w", err)
	}

	query := `
		INSERT INTO failure_patterns (
			id, failure_type, error_signature, error_code, endpoint_pattern,
			http_method, api_source, root_cause, fix_description, fix_payload, embedding
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (id) DO UPDATE SET
			times_encountered = failure_patterns.times_encountered + 1,
			last_seen_at = CURRENT_TIMESTAMP,
			updated_at = CURRENT_TIMESTAMP
	`

	_, err = s.db.ExecContext(ctx, query,
		pattern.ID,
		pattern.FailureType,
		pattern.ErrorSignature,
		pattern.ErrorCode,
		pattern.EndpointPattern,
		pattern.HTTPMethod,
		pattern.APISource,
		pattern.RootCause,
		pattern.FixDescription,
		fixPayloadJSON,
		vectorToString(pattern.Embedding),
	)
	if err != nil {
		return fmt.Errorf("failed to store failure pattern: %w", err)
	}

	return nil
}

// StoreStrategy stores a successful strategy with auto-generated embedding
func (s *Store) StoreStrategy(ctx context.Context, strategy *SuccessfulStrategy) error {
	if strategy.Embedding == nil && s.embeddingClient != nil {
		text := fmt.Sprintf("%s: %s. %s", strategy.StrategyType, strategy.StrategyName, strategy.Description)
		embedding, err := s.embeddingClient.Embed(ctx, text)
		if err != nil {
			return fmt.Errorf("failed to generate embedding: %w", err)
		}
		strategy.Embedding = embedding
	}

	if strategy.ID == uuid.Nil {
		strategy.ID = uuid.New()
	}

	contentJSON, _ := json.Marshal(strategy.StrategyContent)
	patternsJSON, _ := json.Marshal(strategy.ApplicablePatterns)

	query := `
		INSERT INTO successful_strategies (
			id, strategy_type, strategy_name, description, strategy_content,
			applicable_patterns, api_source, endpoint_patterns, embedding
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`

	_, err := s.db.ExecContext(ctx, query,
		strategy.ID,
		strategy.StrategyType,
		strategy.StrategyName,
		strategy.Description,
		contentJSON,
		patternsJSON,
		strategy.APISource,
		strategy.EndpointPatterns,
		vectorToString(strategy.Embedding),
	)
	if err != nil {
		return fmt.Errorf("failed to store strategy: %w", err)
	}

	return nil
}
