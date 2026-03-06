// Package health provides health checking functionality for OpenTest
package health

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/vectordb"
)

// Status represents the health status of a component
type Status string

const (
	StatusHealthy   Status = "healthy"
	StatusDegraded  Status = "degraded"
	StatusUnhealthy Status = "unhealthy"
)

// ComponentHealth represents the health of a single component
type ComponentHealth struct {
	Status      Status    `json:"status"`
	Message     string    `json:"message,omitempty"`
	LastChecked time.Time `json:"last_checked"`
	Latency     string    `json:"latency,omitempty"`
}

// HealthStatus represents the overall system health
type HealthStatus struct {
	Status     Status                     `json:"status"`
	Version    string                     `json:"version"`
	Uptime     string                     `json:"uptime"`
	Timestamp  time.Time                  `json:"timestamp"`
	Components map[string]ComponentHealth `json:"components"`
}

// HealthChecker provides health checking functionality
type HealthChecker struct {
	db          *sql.DB
	vectorStore *vectordb.Store
	startTime   time.Time
	version     string
	mu          sync.RWMutex
	lastStatus  *HealthStatus
}

// HealthCheckerConfig configures the health checker
type HealthCheckerConfig struct {
	DB          *sql.DB
	VectorStore *vectordb.Store
	Version     string
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(cfg HealthCheckerConfig) *HealthChecker {
	version := cfg.Version
	if version == "" {
		version = "dev"
	}

	return &HealthChecker{
		db:          cfg.DB,
		vectorStore: cfg.VectorStore,
		startTime:   time.Now(),
		version:     version,
	}
}

// Check performs a comprehensive health check
func (hc *HealthChecker) Check(ctx context.Context) *HealthStatus {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	status := &HealthStatus{
		Status:     StatusHealthy,
		Version:    hc.version,
		Uptime:     time.Since(hc.startTime).Round(time.Second).String(),
		Timestamp:  time.Now(),
		Components: make(map[string]ComponentHealth),
	}

	// Check database
	if hc.db != nil {
		dbHealth := hc.checkDatabase(ctx)
		status.Components["database"] = dbHealth
		if dbHealth.Status != StatusHealthy {
			status.Status = StatusDegraded
		}
	}

	// Check vector store
	if hc.vectorStore != nil {
		vecHealth := hc.checkVectorStore(ctx)
		status.Components["vector_store"] = vecHealth
		if vecHealth.Status != StatusHealthy && status.Status == StatusHealthy {
			status.Status = StatusDegraded
		}
	}

	// If any component is unhealthy, mark system as degraded
	for _, comp := range status.Components {
		if comp.Status == StatusUnhealthy {
			status.Status = StatusDegraded
			break
		}
	}

	hc.lastStatus = status
	return status
}

// checkDatabase checks database connectivity
func (hc *HealthChecker) checkDatabase(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		LastChecked: time.Now(),
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hc.db.PingContext(ctx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = err.Error()
	} else {
		health.Status = StatusHealthy
		health.Message = "Connected"
	}

	health.Latency = time.Since(start).Round(time.Millisecond).String()
	return health
}

// checkVectorStore checks vector store connectivity
func (hc *HealthChecker) checkVectorStore(ctx context.Context) ComponentHealth {
	start := time.Now()
	health := ComponentHealth{
		LastChecked: time.Now(),
	}

	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := hc.vectorStore.Ping(ctx); err != nil {
		health.Status = StatusUnhealthy
		health.Message = err.Error()
	} else {
		health.Status = StatusHealthy
		health.Message = "Connected"
	}

	health.Latency = time.Since(start).Round(time.Millisecond).String()
	return health
}

// HTTPHandler returns an HTTP handler for health checks
func (hc *HealthChecker) HTTPHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := hc.Check(r.Context())

		w.Header().Set("Content-Type", "application/json")

		switch status.Status {
		case StatusHealthy:
			w.WriteHeader(http.StatusOK)
		case StatusDegraded:
			w.WriteHeader(http.StatusOK) // Still OK but with warnings
		case StatusUnhealthy:
			w.WriteHeader(http.StatusServiceUnavailable)
		}

		json.NewEncoder(w).Encode(status)
	})
}

// ReadinessHandler returns an HTTP handler for readiness checks
func (hc *HealthChecker) ReadinessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		status := hc.Check(r.Context())

		w.Header().Set("Content-Type", "application/json")

		// Ready if not unhealthy
		if status.Status == StatusUnhealthy {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"ready":  status.Status != StatusUnhealthy,
			"status": status.Status,
		})
	})
}

// LivenessHandler returns an HTTP handler for liveness checks
func (hc *HealthChecker) LivenessHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"alive":   true,
			"uptime":  time.Since(hc.startTime).Round(time.Second).String(),
			"version": hc.version,
		})
	})
}

// GetLastStatus returns the last cached health status
func (hc *HealthChecker) GetLastStatus() *HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.lastStatus
}
