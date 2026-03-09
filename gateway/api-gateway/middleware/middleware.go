package middleware

import (
	"fmt"
	"hash/fnv"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/gateway/api-gateway/config"
)

const (
	// Number of shards for distributed locking
	numShards = 256
)

// Logger middleware - logs all incoming requests
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// Choose status label based on status code
		var status string
		switch {
		case statusCode >= 200 && statusCode < 300:
			status = "SUCCESS"
		case statusCode >= 300 && statusCode < 400:
			status = "REDIRECT"
		case statusCode >= 400 && statusCode < 500:
			status = "CLIENT_ERROR"
		case statusCode >= 500:
			status = "SERVER_ERROR"
		default:
			status = "UNKNOWN"
		}

		// Log the request
		log.Printf("[%s] %s %s %d %v", status, method, path, statusCode, latency)
	}
}

// CORS middleware - handles Cross-Origin Resource Sharing
func CORS(cfg *config.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Set CORS headers
		if cfg.CORSAllowedOrigins == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else {
			// Check if origin is allowed
			allowedOrigins := strings.Split(cfg.CORSAllowedOrigins, ",")
			for _, allowedOrigin := range allowedOrigins {
				if strings.TrimSpace(allowedOrigin) == origin {
					c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE, PATCH")

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// rateLimiterShard represents a single shard with its own lock
type rateLimiterShard struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
}

// RateLimiter - optimized in-memory rate limiting with sharded locks
type RateLimiter struct {
	shards   [numShards]*rateLimiterShard
	limit    int
	duration time.Duration
	stopChan chan struct{}
}

// NewRateLimiter creates a new rate limiter with sharded locks
func NewRateLimiter(limit int, duration time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limit:    limit,
		duration: duration,
		stopChan: make(chan struct{}),
	}

	// Initialize all shards
	for i := 0; i < numShards; i++ {
		rl.shards[i] = &rateLimiterShard{
			requests: make(map[string][]time.Time),
		}
	}

	// Start background cleanup goroutine
	go rl.cleanupRoutine()

	log.Printf("INFO: Rate limiter initialized with %d shards", numShards)

	return rl
}

// getShard returns the shard for a given IP address using FNV hash
func (rl *RateLimiter) getShard(ip string) *rateLimiterShard {
	h := fnv.New32a()
	h.Write([]byte(ip))
	return rl.shards[h.Sum32()%numShards]
}

// cleanupRoutine runs periodically to clean up old entries
func (rl *RateLimiter) cleanupRoutine() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanup()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanup removes expired entries from all shards
func (rl *RateLimiter) cleanup() {
	now := time.Now()
	cleaned := 0

	for _, shard := range rl.shards {
		shard.mu.Lock()
		for ip, requests := range shard.requests {
			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.duration {
					validRequests = append(validRequests, reqTime)
				}
			}

			if len(validRequests) == 0 {
				delete(shard.requests, ip)
				cleaned++
			} else {
				shard.requests[ip] = validRequests
			}
		}
		shard.mu.Unlock()
	}

	if cleaned > 0 {
		log.Printf("DEBUG: Rate limiter cleanup removed %d expired IP entries", cleaned)
	}
}

// Stop stops the background cleanup routine
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// RateLimit middleware - limits requests per IP using sharded locks
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		shard := rl.getShard(clientIP)
		now := time.Now()

		// Use read lock first to check if rate limit is exceeded
		shard.mu.RLock()
		requests, exists := shard.requests[clientIP]

		if exists {
			// Count valid requests within the time window
			validCount := 0
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.duration {
					validCount++
				}
			}

			// Check if rate limit exceeded
			if validCount >= rl.limit {
				shard.mu.RUnlock()
				log.Printf("RATE_LIMIT: Rate limit exceeded for IP: %s", clientIP)
				c.JSON(http.StatusTooManyRequests, gin.H{
					"error":   "Rate limit exceeded",
					"message": fmt.Sprintf("Maximum %d requests per %v allowed", rl.limit, rl.duration),
				})
				c.Abort()
				return
			}
		}
		shard.mu.RUnlock()

		// Acquire write lock to update requests
		shard.mu.Lock()

		// Clean up old requests and add new one
		validRequests := []time.Time{}
		if requests, exists := shard.requests[clientIP]; exists {
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.duration {
					validRequests = append(validRequests, reqTime)
				}
			}
		}
		validRequests = append(validRequests, now)
		shard.requests[clientIP] = validRequests

		shard.mu.Unlock()

		c.Next()
	}
}

// Recovery middleware - recovers from panics
func Recovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("PANIC: Panic recovered: %v", err)
				c.JSON(http.StatusInternalServerError, gin.H{
					"error":   "Internal server error",
					"message": "An unexpected error occurred",
				})
				c.Abort()
			}
		}()
		c.Next()
	}
}
