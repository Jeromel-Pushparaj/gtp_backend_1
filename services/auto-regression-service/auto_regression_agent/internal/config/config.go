package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server            ServerConfig            `yaml:"server"`
	Database          DatabaseConfig          `yaml:"database"`
	Redis             RedisConfig             `yaml:"redis"`
	Queue             QueueConfig             `yaml:"queue"`
	Orchestration     OrchestrationConfig     `yaml:"orchestration"`
	Security          SecurityConfig          `yaml:"security"`
	Observability     ObservabilityConfig     `yaml:"observability"`
	LLM               LLMConfig               `yaml:"llm"`
	Agents            AgentsConfig            `yaml:"agents"`
	Knowledge         KnowledgeConfig         `yaml:"knowledge"`
	PayloadGeneration PayloadGenerationConfig `yaml:"payload_generation"`
	VectorDB          VectorDBConfig          `yaml:"vector_db"`
	Sessions          SessionsConfig          `yaml:"sessions"`
	TestHistory       TestHistoryConfig       `yaml:"test_history"`
}

// SessionsConfig represents chat session configuration
type SessionsConfig struct {
	StorageType    string        `yaml:"storage_type"`    // file, database
	StoragePath    string        `yaml:"storage_path"`    // For file storage
	MaxAge         time.Duration `yaml:"max_age"`         // Max session age before cleanup
	CleanupEnabled bool          `yaml:"cleanup_enabled"` // Enable automatic cleanup
	CleanupPeriod  time.Duration `yaml:"cleanup_period"`  // How often to run cleanup
}

// TestHistoryConfig represents test history configuration
type TestHistoryConfig struct {
	Enabled        bool          `yaml:"enabled"`
	RetentionDays  int           `yaml:"retention_days"`    // How long to keep test runs
	MaxRunsPerFlow int           `yaml:"max_runs_per_flow"` // Max runs to keep per workflow
	CleanupPeriod  time.Duration `yaml:"cleanup_period"`    // How often to run cleanup
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	GRPCPort        int           `yaml:"grpc_port"`
	ReadTimeout     time.Duration `yaml:"read_timeout"`
	WriteTimeout    time.Duration `yaml:"write_timeout"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

// DatabaseConfig represents database configuration
type DatabaseConfig struct {
	Host            string        `yaml:"host"`
	Port            int           `yaml:"port"`
	Name            string        `yaml:"name"`
	User            string        `yaml:"user"`
	Password        string        `yaml:"password"`
	SSLMode         string        `yaml:"ssl_mode"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
	LogQueries      bool          `yaml:"log_queries"`
}

// RedisConfig represents Redis configuration
type RedisConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	Password     string `yaml:"password"`
	DB           int    `yaml:"db"`
	MaxRetries   int    `yaml:"max_retries"`
	PoolSize     int    `yaml:"pool_size"`
	MinIdleConns int    `yaml:"min_idle_conns"`
}

// QueueConfig represents queue configuration
type QueueConfig struct {
	Provider string           `yaml:"provider"` // redis_streams, nats
	Redis    RedisQueueConfig `yaml:"redis"`
}

// RedisQueueConfig represents Redis queue configuration
type RedisQueueConfig struct {
	StreamName    string `yaml:"stream_name"`
	ConsumerGroup string `yaml:"consumer_group"`
	MaxLen        int64  `yaml:"max_len"`
}

// OrchestrationConfig represents orchestration configuration
type OrchestrationConfig struct {
	Workflow WorkflowConfig `yaml:"workflow"`
	Executor ExecutorConfig `yaml:"executor"`
}

// WorkflowConfig represents workflow configuration
type WorkflowConfig struct {
	StateTimeout time.Duration `yaml:"state_timeout"`
	MaxRetries   int           `yaml:"max_retries"`
}

// ExecutorConfig represents executor configuration
type ExecutorConfig struct {
	DefaultTimeout    time.Duration `yaml:"default_timeout"`
	MaxConcurrency    int           `yaml:"max_concurrency"`
	ParallelExecution bool          `yaml:"parallel_execution"`
}

// SecurityConfig represents security configuration
type SecurityConfig struct {
	Auth      AuthConfig      `yaml:"auth"`
	RateLimit RateLimitConfig `yaml:"rate_limit"`
}

// AuthConfig represents authentication configuration
type AuthConfig struct {
	JWT    JWTConfig    `yaml:"jwt"`
	APIKey APIKeyConfig `yaml:"api_key"`
}

// JWTConfig represents JWT configuration
type JWTConfig struct {
	Secret     string        `yaml:"secret"`
	Issuer     string        `yaml:"issuer"`
	Expiration time.Duration `yaml:"expiration"`
}

// APIKeyConfig represents API key configuration
type APIKeyConfig struct {
	Enabled    bool   `yaml:"enabled"`
	HeaderName string `yaml:"header_name"`
}

// RateLimitConfig represents rate limiting configuration
type RateLimitConfig struct {
	Enabled           bool `yaml:"enabled"`
	RequestsPerMinute int  `yaml:"requests_per_minute"`
	Burst             int  `yaml:"burst"`
}

// ObservabilityConfig represents observability configuration
type ObservabilityConfig struct {
	Logging LoggingConfig `yaml:"logging"`
	Metrics MetricsConfig `yaml:"metrics"`
	Health  HealthConfig  `yaml:"health"`
}

// LoggingConfig represents logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`  // debug, info, warn, error
	Format string `yaml:"format"` // json, text
	Output string `yaml:"output"` // stdout, stderr, file
}

// MetricsConfig represents metrics configuration
type MetricsConfig struct {
	Enabled    bool             `yaml:"enabled"`
	Prometheus PrometheusConfig `yaml:"prometheus"`
}

// PrometheusConfig represents Prometheus configuration
type PrometheusConfig struct {
	Port int    `yaml:"port"`
	Path string `yaml:"path"`
}

// HealthConfig represents health check configuration
type HealthConfig struct {
	Enabled       bool   `yaml:"enabled"`
	Path          string `yaml:"path"`
	ReadinessPath string `yaml:"readiness_path"`
}

// LLMConfig represents LLM configuration
type LLMConfig struct {
	Provider    string  `yaml:"provider"`    // groq (active), openai, anthropic, local
	APIKey      string  `yaml:"api_key"`     // Can be overridden by env var (GROQ_API_KEY or OPENAI_API_KEY)
	Model       string  `yaml:"model"`       // openai/gpt-oss-120b (Groq) or gpt-4-turbo-preview (OpenAI)
	Temperature float64 `yaml:"temperature"` // 0.0 - 1.0
	MaxTokens   int     `yaml:"max_tokens"`  // Max tokens per request
	Timeout     int     `yaml:"timeout"`     // Timeout in seconds
	BaseURL     string  `yaml:"base_url"`    // For custom endpoints (Groq or OpenAI)
}

// VectorDBConfig represents vector database configuration for learning/memory
type VectorDBConfig struct {
	Enabled        bool   `yaml:"enabled"`         // Enable vector database for learning
	Host           string `yaml:"host"`            // PostgreSQL host
	Port           int    `yaml:"port"`            // PostgreSQL port
	Database       string `yaml:"database"`        // Database name
	User           string `yaml:"user"`            // Database user
	Password       string `yaml:"password"`        // Database password
	SSLMode        string `yaml:"ssl_mode"`        // SSL mode (disable, require, verify-full)
	EmbeddingModel string `yaml:"embedding_model"` // Embedding model (text-embedding-004)
}

// AgentsConfig represents AI agents configuration
type AgentsConfig struct {
	Payload PayloadAgentConfig `yaml:"payload"`
	Healing HealingAgentConfig `yaml:"healing"`
}

// PayloadAgentConfig represents payload generation agent configuration
type PayloadAgentConfig struct {
	Enabled          bool    `yaml:"enabled"`
	Temperature      float64 `yaml:"temperature"`
	MaxTokens        int     `yaml:"max_tokens"`
	MaxRetries       int     `yaml:"max_retries"`
	StrictMode       bool    `yaml:"strict_mode"`       // Reject hallucinated fields
	GenerateNegative bool    `yaml:"generate_negative"` // Generate negative test cases
	GenerateBoundary bool    `yaml:"generate_boundary"` // Generate boundary test cases
}

// HealingAgentConfig represents self-healing agent configuration
type HealingAgentConfig struct {
	Enabled             bool    `yaml:"enabled"`
	ApprovalRequired    bool    `yaml:"approval_required"`    // Always require human approval
	MaxFixes            int     `yaml:"max_fixes"`            // Max fixes per analysis
	ConfidenceThreshold float64 `yaml:"confidence_threshold"` // Min confidence to propose fix
}

// KnowledgeConfig represents knowledge base and learning configuration
type KnowledgeConfig struct {
	Enabled       bool                `yaml:"enabled"`
	StoragePath   string              `yaml:"storage_path"`
	AutoSave      bool                `yaml:"auto_save"`
	SaveInterval  time.Duration       `yaml:"save_interval"`
	Learning      LearningConfig      `yaml:"learning"`
	Collaboration CollaborationConfig `yaml:"collaboration"`
	Adaptive      AdaptiveConfig      `yaml:"adaptive"`
}

// LearningConfig represents learning engine configuration
type LearningConfig struct {
	Enabled              bool    `yaml:"enabled"`
	MinConfidence        float64 `yaml:"min_confidence"`
	PatternRetentionDays int     `yaml:"pattern_retention_days"`
}

// CollaborationConfig represents agent collaboration configuration
type CollaborationConfig struct {
	Enabled            bool    `yaml:"enabled"`
	ConsensusThreshold float64 `yaml:"consensus_threshold"`
	FeedbackBufferSize int     `yaml:"feedback_buffer_size"`
}

// AdaptiveConfig represents adaptive testing configuration
type AdaptiveConfig struct {
	Enabled                bool    `yaml:"enabled"`
	FocusOnFailures        bool    `yaml:"focus_on_failures"`
	SkipStableTests        bool    `yaml:"skip_stable_tests"`
	StabilityThreshold     int     `yaml:"stability_threshold"`
	FailureFocusMultiplier float64 `yaml:"failure_focus_multiplier"`
	RiskBasedPriority      bool    `yaml:"risk_based_priority"`
}

// PayloadGenerationConfig represents payload generation enhancements configuration
type PayloadGenerationConfig struct {
	Enabled            bool                  `yaml:"enabled"`
	SmartData          SmartDataConfig       `yaml:"smart_data"`
	Mutation           MutationTestConfig    `yaml:"mutation"`
	Security           SecurityTestingConfig `yaml:"security"`
	Performance        PerformanceTestConfig `yaml:"performance"`
	MaxPayloadsPerType int                   `yaml:"max_payloads_per_type"`
	EnableCombinations bool                  `yaml:"enable_combinations"`
	UseKnowledgeBase   bool                  `yaml:"use_knowledge_base"`
}

// SmartDataConfig represents smart data generation configuration
type SmartDataConfig struct {
	Enabled               bool   `yaml:"enabled"`
	UseRealisticData      bool   `yaml:"use_realistic_data"`
	LocalePreference      string `yaml:"locale_preference"`
	ApplyRelationships    bool   `yaml:"apply_relationships"`
	UseHistoricalPatterns bool   `yaml:"use_historical_patterns"`
}

// MutationTestConfig represents mutation-based testing configuration
type MutationTestConfig struct {
	Enabled              bool `yaml:"enabled"`
	EnableFieldMutation  bool `yaml:"enable_field_mutation"`
	EnableTypeMutation   bool `yaml:"enable_type_mutation"`
	EnableBoundaryTests  bool `yaml:"enable_boundary_tests"`
	CombinationDepth     int  `yaml:"combination_depth"`
	MaxMutationsPerField int  `yaml:"max_mutations_per_field"`
}

// SecurityTestingConfig represents security testing configuration
type SecurityTestingConfig struct {
	Enabled                bool     `yaml:"enabled"`
	EnableSQLInjection     bool     `yaml:"enable_sql_injection"`
	EnableXSS              bool     `yaml:"enable_xss"`
	EnableCommandInjection bool     `yaml:"enable_command_injection"`
	EnablePathTraversal    bool     `yaml:"enable_path_traversal"`
	EnableXMLInjection     bool     `yaml:"enable_xml_injection"`
	EnableLDAPInjection    bool     `yaml:"enable_ldap_injection"`
	EnableAuthBypass       bool     `yaml:"enable_auth_bypass"`
	OWASPCategories        []string `yaml:"owasp_categories"`
}

// PerformanceTestConfig represents performance testing configuration
type PerformanceTestConfig struct {
	Enabled                  bool          `yaml:"enabled"`
	EnableConcurrentRequests bool          `yaml:"enable_concurrent_requests"`
	EnableLoadRampUp         bool          `yaml:"enable_load_ramp_up"`
	EnableSpikeTest          bool          `yaml:"enable_spike_test"`
	EnableEnduranceTest      bool          `yaml:"enable_endurance_test"`
	ConcurrencyLevels        []int         `yaml:"concurrency_levels"`
	RampUpDuration           time.Duration `yaml:"ramp_up_duration"`
	SpikeDuration            time.Duration `yaml:"spike_duration"`
	EnduranceDuration        time.Duration `yaml:"endurance_duration"`
}

// Load loads configuration from file and environment variables
func Load(configPath string) (*Config, error) {
	// Read config file
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config: %w", err)
	}

	// Override with environment variables
	cfg.applyEnvOverrides()

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

// applyEnvOverrides applies environment variable overrides
func (c *Config) applyEnvOverrides() {
	// Database overrides
	if v := os.Getenv("POSTGRES_HOST"); v != "" {
		c.Database.Host = v
	}
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Database.Port)
	}
	if v := os.Getenv("POSTGRES_DB"); v != "" {
		c.Database.Name = v
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		c.Database.User = v
	}
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		c.Database.Password = v
	}
	if v := os.Getenv("DB_PASSWORD"); v != "" {
		c.Database.Password = v
	}

	// Redis overrides
	if v := os.Getenv("REDIS_HOST"); v != "" {
		c.Redis.Host = v
	}
	if v := os.Getenv("REDIS_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.Redis.Port)
	}
	if v := os.Getenv("REDIS_PASSWORD"); v != "" {
		c.Redis.Password = v
	}
	if v := os.Getenv("REDIS_DB"); v != "" {
		fmt.Sscanf(v, "%d", &c.Redis.DB)
	}

	// Security overrides
	if v := os.Getenv("JWT_SECRET"); v != "" {
		c.Security.Auth.JWT.Secret = v
	}

	// LLM overrides
	// Priority: GROQ_API_KEY (active) > LLM_API_KEY > OPENAI_API_KEY
	if v := os.Getenv("LLM_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	if v := os.Getenv("GROQ_API_KEY"); v != "" {
		c.LLM.APIKey = v
	}
	// For OpenAI support
	if v := os.Getenv("OPENAI_API_KEY"); v != "" && os.Getenv("GROQ_API_KEY") == "" {
		// Only use OPENAI_API_KEY if GROQ_API_KEY is not set
		c.LLM.APIKey = v
	}
	if v := os.Getenv("LLM_PROVIDER"); v != "" {
		c.LLM.Provider = v
	}
	if v := os.Getenv("LLM_MODEL"); v != "" {
		c.LLM.Model = v
	}

	// VectorDB overrides (uses same Postgres env vars by default)
	if v := os.Getenv("POSTGRES_HOST"); v != "" {
		c.VectorDB.Host = v
	}
	if v := os.Getenv("POSTGRES_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.VectorDB.Port)
	}
	if v := os.Getenv("POSTGRES_DB"); v != "" {
		c.VectorDB.Database = v
	}
	if v := os.Getenv("POSTGRES_USER"); v != "" {
		c.VectorDB.User = v
	}
	if v := os.Getenv("POSTGRES_PASSWORD"); v != "" {
		c.VectorDB.Password = v
	}
	// VectorDB-specific overrides (take precedence)
	if v := os.Getenv("VECTORDB_HOST"); v != "" {
		c.VectorDB.Host = v
	}
	if v := os.Getenv("VECTORDB_PORT"); v != "" {
		fmt.Sscanf(v, "%d", &c.VectorDB.Port)
	}
	if v := os.Getenv("VECTORDB_DATABASE"); v != "" {
		c.VectorDB.Database = v
	}
	if v := os.Getenv("VECTORDB_ENABLED"); v != "" {
		c.VectorDB.Enabled = v == "true" || v == "1"
	}
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}
	if c.Database.Host == "" {
		return fmt.Errorf("database host is required")
	}
	if c.Redis.Host == "" {
		return fmt.Errorf("redis host is required")
	}
	return nil
}
