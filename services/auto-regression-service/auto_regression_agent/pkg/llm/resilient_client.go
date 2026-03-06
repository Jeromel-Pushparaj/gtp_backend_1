package llm

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/resilience"
)

// ResilientClient wraps an LLM client with retry and circuit breaker
type ResilientClient struct {
	client         *Client
	retryer        *resilience.Retryer
	circuitBreaker *resilience.CircuitBreaker
}

// ResilientConfig configures the resilient client
type ResilientConfig struct {
	Client         Config
	Retry          resilience.RetryConfig
	CircuitBreaker resilience.CircuitBreakerConfig
}

// NewResilientClient creates a new resilient LLM client
func NewResilientClient(config ResilientConfig) *ResilientClient {
	return &ResilientClient{
		client:         NewClient(config.Client),
		retryer:        resilience.NewRetryer(config.Retry),
		circuitBreaker: resilience.NewCircuitBreaker(config.CircuitBreaker),
	}
}

// NewResilientClientWithDefaults creates a resilient client with default retry/circuit breaker settings
func NewResilientClientWithDefaults(clientConfig Config) *ResilientClient {
	return NewResilientClient(ResilientConfig{
		Client:         clientConfig,
		Retry:          resilience.DefaultRetryConfig(),
		CircuitBreaker: resilience.DefaultCircuitBreakerConfig(),
	})
}

// GenerateCompletion generates a completion with retry and circuit breaker protection
func (rc *ResilientClient) GenerateCompletion(ctx context.Context, prompt string, opts CompletionOptions) (string, error) {
	var result string

	err := rc.circuitBreaker.Execute(ctx, func() error {
		return rc.retryer.Do(ctx, func() error {
			var err error
			result, err = rc.client.GenerateCompletion(ctx, prompt, opts)
			if err != nil {
				return rc.wrapError(err)
			}
			return nil
		})
	})

	if err != nil {
		return "", err
	}
	return result, nil
}

// GenerateAgentCompletion generates an agent completion with retry and circuit breaker protection
func (rc *ResilientClient) GenerateAgentCompletion(ctx context.Context, messages []AgentMessage, opts AgentCompletionOptions) (*AgentResponse, error) {
	var result *AgentResponse

	err := rc.circuitBreaker.Execute(ctx, func() error {
		return rc.retryer.Do(ctx, func() error {
			var err error
			result, err = rc.client.GenerateAgentCompletion(ctx, messages, opts)
			if err != nil {
				return rc.wrapError(err)
			}
			return nil
		})
	})

	if err != nil {
		return nil, err
	}
	return result, nil
}

// wrapError wraps an error with retry information based on the error message
func (rc *ResilientClient) wrapError(err error) error {
	if err == nil {
		return nil
	}

	errStr := err.Error()

	// Check for rate limiting
	if strings.Contains(errStr, "429") || strings.Contains(errStr, "Too Many Requests") {
		return &resilience.RetryableError{
			Err:        err,
			StatusCode: http.StatusTooManyRequests,
			Retryable:  true,
		}
	}

	// Check for server errors
	if strings.Contains(errStr, "500") || strings.Contains(errStr, "Internal Server Error") {
		return &resilience.RetryableError{
			Err:        err,
			StatusCode: http.StatusInternalServerError,
			Retryable:  true,
		}
	}

	if strings.Contains(errStr, "502") || strings.Contains(errStr, "Bad Gateway") {
		return &resilience.RetryableError{
			Err:        err,
			StatusCode: http.StatusBadGateway,
			Retryable:  true,
		}
	}

	if strings.Contains(errStr, "503") || strings.Contains(errStr, "Service Unavailable") {
		return &resilience.RetryableError{
			Err:        err,
			StatusCode: http.StatusServiceUnavailable,
			Retryable:  true,
		}
	}

	if strings.Contains(errStr, "504") || strings.Contains(errStr, "Gateway Timeout") {
		return &resilience.RetryableError{
			Err:        err,
			StatusCode: http.StatusGatewayTimeout,
			Retryable:  true,
		}
	}

	// Non-retryable error
	return err
}

// Provider returns the underlying provider name
func (rc *ResilientClient) Provider() string {
	return rc.client.Provider()
}

// CircuitBreakerState returns the current circuit breaker state
func (rc *ResilientClient) CircuitBreakerState() resilience.CircuitState {
	return rc.circuitBreaker.State()
}

// CircuitBreakerStats returns circuit breaker statistics
func (rc *ResilientClient) CircuitBreakerStats() resilience.CircuitBreakerStats {
	return rc.circuitBreaker.Stats()
}

// OnCircuitStateChange sets a callback for circuit breaker state changes
func (rc *ResilientClient) OnCircuitStateChange(fn func(from, to resilience.CircuitState)) {
	rc.circuitBreaker.OnStateChange(fn)
}

// ResetCircuitBreaker resets the circuit breaker to closed state
func (rc *ResilientClient) ResetCircuitBreaker() {
	rc.circuitBreaker.Reset()
}

// IsHealthy returns true if the circuit breaker is not open
func (rc *ResilientClient) IsHealthy() bool {
	return rc.circuitBreaker.State() != resilience.StateOpen
}

// HealthStatus returns a human-readable health status
func (rc *ResilientClient) HealthStatus() string {
	state := rc.circuitBreaker.State()
	switch state {
	case resilience.StateClosed:
		return "healthy"
	case resilience.StateHalfOpen:
		return "recovering"
	case resilience.StateOpen:
		return "unhealthy"
	default:
		return fmt.Sprintf("unknown (%s)", state)
	}
}

