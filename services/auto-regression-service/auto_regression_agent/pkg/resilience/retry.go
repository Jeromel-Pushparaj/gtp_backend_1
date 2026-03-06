// Package resilience provides retry and circuit breaker patterns
package resilience

import (
	"context"
	"errors"
	"math"
	"math/rand"
	"net/http"
	"time"
)

// RetryConfig configures retry behavior
type RetryConfig struct {
	MaxAttempts     int           // Maximum number of attempts (including initial)
	InitialDelay    time.Duration // Initial delay between retries
	MaxDelay        time.Duration // Maximum delay between retries
	Multiplier      float64       // Backoff multiplier
	Jitter          float64       // Jitter factor (0-1)
	RetryableErrors []error       // Specific errors to retry on
	RetryableStatus []int         // HTTP status codes to retry on
}

// DefaultRetryConfig returns sensible defaults for LLM API calls
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxAttempts:  3,
		InitialDelay: 1 * time.Second,
		MaxDelay:     30 * time.Second,
		Multiplier:   2.0,
		Jitter:       0.1,
		RetryableStatus: []int{
			http.StatusTooManyRequests,     // 429
			http.StatusInternalServerError, // 500
			http.StatusBadGateway,          // 502
			http.StatusServiceUnavailable,  // 503
			http.StatusGatewayTimeout,      // 504
		},
	}
}

// Retryer handles retry logic with exponential backoff
type Retryer struct {
	config RetryConfig
}

// NewRetryer creates a new retryer with the given configuration
func NewRetryer(config RetryConfig) *Retryer {
	if config.MaxAttempts <= 0 {
		config.MaxAttempts = 3
	}
	if config.InitialDelay <= 0 {
		config.InitialDelay = time.Second
	}
	if config.MaxDelay <= 0 {
		config.MaxDelay = 30 * time.Second
	}
	if config.Multiplier <= 0 {
		config.Multiplier = 2.0
	}
	return &Retryer{config: config}
}

// RetryableError wraps an error with retry information
type RetryableError struct {
	Err        error
	StatusCode int
	Retryable  bool
}

func (e *RetryableError) Error() string {
	return e.Err.Error()
}

func (e *RetryableError) Unwrap() error {
	return e.Err
}

// Do executes the function with retry logic
func (r *Retryer) Do(ctx context.Context, fn func() error) error {
	var lastErr error

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		// Check context before attempting
		if ctx.Err() != nil {
			return ctx.Err()
		}

		err := fn()
		if err == nil {
			return nil
		}

		lastErr = err

		// Check if error is retryable
		if !r.isRetryable(err) {
			return err
		}

		// Don't sleep after the last attempt
		if attempt < r.config.MaxAttempts-1 {
			delay := r.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return lastErr
}

// DoWithResult executes a function that returns a result with retry logic
func (r *Retryer) DoWithResult(ctx context.Context, fn func() (interface{}, error)) (interface{}, error) {
	var lastErr error
	var result interface{}

	for attempt := 0; attempt < r.config.MaxAttempts; attempt++ {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		result, lastErr = fn()
		if lastErr == nil {
			return result, nil
		}

		if !r.isRetryable(lastErr) {
			return nil, lastErr
		}

		if attempt < r.config.MaxAttempts-1 {
			delay := r.calculateDelay(attempt)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}
	}

	return nil, lastErr
}

// isRetryable checks if an error should be retried
func (r *Retryer) isRetryable(err error) bool {
	// Check for RetryableError
	var retryableErr *RetryableError
	if errors.As(err, &retryableErr) {
		// Check status code
		for _, status := range r.config.RetryableStatus {
			if retryableErr.StatusCode == status {
				return true
			}
		}
		return retryableErr.Retryable
	}

	// Check for specific retryable errors
	for _, retryable := range r.config.RetryableErrors {
		if errors.Is(err, retryable) {
			return true
		}
	}

	// Default: retry on context deadline exceeded (timeout)
	return errors.Is(err, context.DeadlineExceeded)
}

// calculateDelay calculates the delay for the given attempt
func (r *Retryer) calculateDelay(attempt int) time.Duration {
	// Exponential backoff
	delay := float64(r.config.InitialDelay) * math.Pow(r.config.Multiplier, float64(attempt))

	// Apply jitter
	if r.config.Jitter > 0 {
		jitter := delay * r.config.Jitter * (rand.Float64()*2 - 1)
		delay += jitter
	}

	// Cap at max delay
	if delay > float64(r.config.MaxDelay) {
		delay = float64(r.config.MaxDelay)
	}

	return time.Duration(delay)
}
