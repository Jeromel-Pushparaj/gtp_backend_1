// Package shutdown provides graceful shutdown handling
package shutdown

import (
	"context"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// Handler manages graceful shutdown of components
type Handler struct {
	timeout    time.Duration
	components []Component
	mu         sync.Mutex
	done       chan struct{}
	logger     Logger
}

// Component represents a component that can be shut down
type Component interface {
	Name() string
	Shutdown(ctx context.Context) error
}

// Logger interface for shutdown logging
type Logger interface {
	Info(msg string, fields ...interface{})
	Error(msg string, fields ...interface{})
}

// defaultLogger is a simple logger that prints to stdout
type defaultLogger struct{}

func (l *defaultLogger) Info(msg string, fields ...interface{}) {
	// Simple implementation - in production use structured logging
}

func (l *defaultLogger) Error(msg string, fields ...interface{}) {
	// Simple implementation - in production use structured logging
}

// Config configures the shutdown handler
type Config struct {
	Timeout time.Duration
	Logger  Logger
}

// NewHandler creates a new shutdown handler
func NewHandler(cfg Config) *Handler {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Logger == nil {
		cfg.Logger = &defaultLogger{}
	}

	return &Handler{
		timeout:    cfg.Timeout,
		components: make([]Component, 0),
		done:       make(chan struct{}),
		logger:     cfg.Logger,
	}
}

// Register registers a component for graceful shutdown
func (h *Handler) Register(component Component) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.components = append(h.components, component)
}

// RegisterFunc registers a shutdown function with a name
func (h *Handler) RegisterFunc(name string, fn func(ctx context.Context) error) {
	h.Register(&funcComponent{name: name, fn: fn})
}

// funcComponent wraps a function as a Component
type funcComponent struct {
	name string
	fn   func(ctx context.Context) error
}

func (fc *funcComponent) Name() string {
	return fc.name
}

func (fc *funcComponent) Shutdown(ctx context.Context) error {
	return fc.fn(ctx)
}

// Wait blocks until a shutdown signal is received
func (h *Handler) Wait() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	select {
	case sig := <-sigChan:
		h.logger.Info("Received shutdown signal", "signal", sig)
	case <-h.done:
		h.logger.Info("Shutdown initiated programmatically")
	}
}

// Shutdown initiates graceful shutdown
func (h *Handler) Shutdown() {
	close(h.done)
}

// Execute performs graceful shutdown of all registered components
func (h *Handler) Execute() error {
	ctx, cancel := context.WithTimeout(context.Background(), h.timeout)
	defer cancel()

	h.mu.Lock()
	components := make([]Component, len(h.components))
	copy(components, h.components)
	h.mu.Unlock()

	// Shutdown in reverse order (LIFO)
	var lastErr error
	for i := len(components) - 1; i >= 0; i-- {
		component := components[i]
		h.logger.Info("Shutting down component", "name", component.Name())

		if err := component.Shutdown(ctx); err != nil {
			h.logger.Error("Failed to shutdown component", "name", component.Name(), "error", err)
			lastErr = err
		} else {
			h.logger.Info("Component shutdown complete", "name", component.Name())
		}
	}

	return lastErr
}

// WaitAndShutdown waits for a signal and then executes shutdown
func (h *Handler) WaitAndShutdown() error {
	h.Wait()
	return h.Execute()
}

// HTTPServerComponent wraps an HTTP server for graceful shutdown
type HTTPServerComponent struct {
	name   string
	server interface {
		Shutdown(ctx context.Context) error
	}
}

// NewHTTPServerComponent creates a new HTTP server component
func NewHTTPServerComponent(name string, server interface{ Shutdown(ctx context.Context) error }) *HTTPServerComponent {
	return &HTTPServerComponent{name: name, server: server}
}

func (c *HTTPServerComponent) Name() string {
	return c.name
}

func (c *HTTPServerComponent) Shutdown(ctx context.Context) error {
	return c.server.Shutdown(ctx)
}

