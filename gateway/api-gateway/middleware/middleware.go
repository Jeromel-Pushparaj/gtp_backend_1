package middleware

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jeromelp/gtp_backend_1/gateway/api-gateway/config"
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

// RateLimiter - simple in-memory rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
	limit    int
	duration time.Duration
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, duration time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		duration: duration,
	}
}

// RateLimit middleware - limits requests per IP
func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get client IP
		clientIP := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Clean up old requests
		now := time.Now()
		if requests, exists := rl.requests[clientIP]; exists {
			validRequests := []time.Time{}
			for _, reqTime := range requests {
				if now.Sub(reqTime) < rl.duration {
					validRequests = append(validRequests, reqTime)
				}
			}
			rl.requests[clientIP] = validRequests
		}

		// Check rate limit
		if len(rl.requests[clientIP]) >= rl.limit {
			log.Printf("RATE_LIMIT: Rate limit exceeded for IP: %s", clientIP)
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Maximum %d requests per %v allowed", rl.limit, rl.duration),
			})
			c.Abort()
			return
		}

		// Add current request
		rl.requests[clientIP] = append(rl.requests[clientIP], now)
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

