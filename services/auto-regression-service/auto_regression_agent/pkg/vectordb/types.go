package vectordb

import (
	"time"

	"github.com/google/uuid"
)

// EmbeddingDimension is the dimension of embeddings
// Default: all-MiniLM-L6-v2 (local) = 384 dimensions
// OpenAI text-embedding-3-small = 1536 dimensions
// all-mpnet-base-v2 (local) = 768 dimensions
// Adjust this constant based on your embedding model
const EmbeddingDimension = 384

// Vector represents a float32 vector for embeddings
type Vector []float32

// Learning represents a learned pattern stored in the database
type Learning struct {
	ID           uuid.UUID              `json:"id"`
	Category     string                 `json:"category"`     // auth_pattern, edge_case, business_rule, api_pattern
	SourceAPI    string                 `json:"source_api"`   // API/spec this was learned from
	Content      string                 `json:"content"`      // The actual learning description
	Context      map[string]interface{} `json:"context"`      // Additional context
	Embedding    Vector                 `json:"embedding"`    // Vector embedding
	Confidence   float64                `json:"confidence"`   // Confidence score 0-1
	TimesApplied int                    `json:"times_applied"`
	LastApplied  *time.Time             `json:"last_applied_at"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
}

// FailurePattern represents a stored failure pattern with its fix
type FailurePattern struct {
	ID              uuid.UUID              `json:"id"`
	FailureType     string                 `json:"failure_type"`     // auth_failure, validation_error, timeout, schema_mismatch
	ErrorSignature  string                 `json:"error_signature"`  // Error message/pattern
	ErrorCode       string                 `json:"error_code"`       // HTTP status or error code
	EndpointPattern string                 `json:"endpoint_pattern"` // e.g., /users/{id}
	HTTPMethod      string                 `json:"http_method"`      // GET, POST, PUT, DELETE
	APISource       string                 `json:"api_source"`       // Which API/spec
	RootCause       string                 `json:"root_cause"`       // Identified root cause
	FixDescription  string                 `json:"fix_description"`  // How to fix
	FixPayload      map[string]interface{} `json:"fix_payload"`      // Example fix
	Embedding       Vector                 `json:"embedding"`        // Vector embedding
	TimesEncountered int                   `json:"times_encountered"`
	TimesFixed      int                    `json:"times_fixed"`
	FixSuccessRate  float64                `json:"fix_success_rate"`
	FirstSeenAt     time.Time              `json:"first_seen_at"`
	LastSeenAt      time.Time              `json:"last_seen_at"`
	CreatedAt       time.Time              `json:"created_at"`
	UpdatedAt       time.Time              `json:"updated_at"`
}

// SuccessfulStrategy represents a test strategy that worked well
type SuccessfulStrategy struct {
	ID                 uuid.UUID              `json:"id"`
	StrategyType       string                 `json:"strategy_type"`       // auth_flow, data_generation, test_sequence
	StrategyName       string                 `json:"strategy_name"`
	Description        string                 `json:"description"`
	StrategyContent    map[string]interface{} `json:"strategy_content"`    // The strategy details
	ApplicablePatterns map[string]interface{} `json:"applicable_patterns"` // When to use
	APISource          string                 `json:"api_source"`
	EndpointPatterns   []string               `json:"endpoint_patterns"`
	Embedding          Vector                 `json:"embedding"`
	SuccessRate        float64                `json:"success_rate"`
	TimesUsed          int                    `json:"times_used"`
	AvgCoverage        float64                `json:"avg_coverage"`
	CreatedAt          time.Time              `json:"created_at"`
	UpdatedAt          time.Time              `json:"updated_at"`
}

// SimilarityResult represents a similarity search result
type SimilarityResult struct {
	ID         uuid.UUID `json:"id"`
	Score      float64   `json:"score"`      // Cosine similarity score (higher is more similar)
	Distance   float64   `json:"distance"`   // Cosine distance (lower is more similar)
	ResultType string    `json:"result_type"` // learning, failure_pattern, strategy
}

// LearningResult wraps Learning with similarity score
type LearningResult struct {
	Learning
	Score float64 `json:"score"`
}

// FailurePatternResult wraps FailurePattern with similarity score
type FailurePatternResult struct {
	FailurePattern
	Score float64 `json:"score"`
}

// StrategyResult wraps SuccessfulStrategy with similarity score
type StrategyResult struct {
	SuccessfulStrategy
	Score float64 `json:"score"`
}

// SearchOptions configures similarity search
type SearchOptions struct {
	Limit      int     `json:"limit"`       // Max results to return
	MinScore   float64 `json:"min_score"`   // Minimum similarity score (0-1)
	Category   string  `json:"category"`    // Filter by category (optional)
	APISource  string  `json:"api_source"`  // Filter by API source (optional)
}

// DefaultSearchOptions returns default search options
func DefaultSearchOptions() SearchOptions {
	return SearchOptions{
		Limit:    10,
		MinScore: 0.5,
	}
}

