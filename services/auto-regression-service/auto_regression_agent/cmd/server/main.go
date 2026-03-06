package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/redis/go-redis/v9"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/api/rest"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/config"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration/job"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration/spec"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/progress"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/storage/queue"
	workflowstore "gitlab.com/tekion/development/toc/poc/opentest/internal/storage/workflow"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

var (
	configPath = flag.String("config", "configs/config.yaml", "Path to configuration file")
)

// getMapKeys returns the keys of a map as a slice
func getMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func main() {
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Auto-Regression Server v1.0.0")
	log.Printf("Server will listen on %s:%d", cfg.Server.Host, cfg.Server.Port)

	ctx := context.Background()

	// Initialize Redis client (optional for development)
	var jobQueue queue.Queue
	redisClient := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		MaxRetries:   cfg.Redis.MaxRetries,
		PoolSize:     cfg.Redis.PoolSize,
		MinIdleConns: cfg.Redis.MinIdleConns,
	})

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️  Warning: Failed to connect to Redis: %v", err)
		log.Printf("⚠️  Running in DEVELOPMENT MODE without Redis")
		log.Printf("⚠️  Job queue will use in-memory storage (data will be lost on restart)")

		// Use in-memory queue for development
		jobQueue = queue.NewMemoryQueue()
		if err := jobQueue.Initialize(ctx); err != nil {
			log.Fatalf("Failed to initialize in-memory queue: %v", err)
		}
		log.Printf("Initialized in-memory job queue (development mode)")
	} else {
		log.Printf("Connected to Redis at %s:%d", cfg.Redis.Host, cfg.Redis.Port)

		// Initialize Redis queue
		jobQueue = queue.NewRedisQueue(
			redisClient,
			cfg.Queue.Redis.StreamName,
			cfg.Queue.Redis.ConsumerGroup,
			cfg.Queue.Redis.MaxLen,
		)

		if err := jobQueue.Initialize(ctx); err != nil {
			log.Fatalf("Failed to initialize queue: %v", err)
		}
		log.Printf("Initialized job queue: %s", cfg.Queue.Redis.StreamName)
	}

	// Initialize workflow store
	wfStore := workflowstore.NewStore()
	log.Printf("Workflow store initialized")

	// Initialize orchestrator components
	specParser := spec.NewParser()
	jobCreator := job.NewCreator(cfg.Orchestration.Workflow.MaxRetries)
	orch := orchestration.NewOrchestrator(specParser, jobCreator, jobQueue, wfStore)

	// Initialize event bus for autonomous agents (if Redis is available)
	if redisClient != nil {
		eventBus := events.NewBus(redisClient)
		orch.SetEventBus(eventBus)
		log.Printf("Event bus initialized for autonomous agents")
	}

	log.Printf("Orchestrator initialized")

	// Initialize REST API server
	server := rest.NewServer(cfg, orch)
	httpServer := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      server.Handler(),
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Start progress subscriber if Redis is available
	if redisClient != nil {
		progressBroadcaster := progress.NewBroadcaster(redisClient, "workflow_progress")
		wsHandler := server.WebSocketHandler()

		// Start progress subscription in goroutine
		go func() {
			log.Println("Starting progress subscriber...")
			err := progressBroadcaster.Subscribe(ctx, func(update progress.ProgressUpdate) {
				// Skip if workflow_id is empty
				if update.WorkflowID == "" {
					log.Printf("WARNING: Received progress update with empty workflow_id for job %s", update.JobID)
					return
				}

				// Update workflow job status and type in store
				if err := wfStore.UpdateJobInfo(ctx, update.WorkflowID, update.JobID, workflow.JobType(update.JobType), workflow.JobStatus(update.Status)); err != nil {
					log.Printf("Failed to update job info: %v", err)
					return
				}

				// Update job result if present
				if update.Result != nil && len(update.Result) > 0 {
					log.Printf("DEBUG: Storing job result for workflow=%s job=%s result_keys=%v", update.WorkflowID, update.JobID, getMapKeys(update.Result))
					if err := wfStore.UpdateJobResult(ctx, update.WorkflowID, update.JobID, update.Result); err != nil {
						log.Printf("Failed to update job result: %v", err)
					} else {
						log.Printf("DEBUG: Successfully stored job result for workflow=%s job=%s", update.WorkflowID, update.JobID)
					}
				} else {
					log.Printf("DEBUG: No result to store for workflow=%s job=%s (result is nil or empty)", update.WorkflowID, update.JobID)
				}

				// Update workflow state based on job type and status
				if update.Status == "running" {
					var newState workflow.WorkflowState
					switch update.JobType {
					case string(workflow.JobTypeSpecAnalysis):
						newState = workflow.WorkflowStateAnalyzing
					case string(workflow.JobTypeTestGeneration):
						newState = workflow.WorkflowStateGenerating
					case string(workflow.JobTypeTestExecution):
						newState = workflow.WorkflowStateExecuting
					case string(workflow.JobTypeResultAnalysis):
						newState = workflow.WorkflowStateReporting
					}
					if newState != "" {
						if err := wfStore.UpdateState(ctx, update.WorkflowID, newState); err != nil {
							log.Printf("Failed to update workflow state: %v", err)
						}
					}
				}

				// Calculate progress
				progressPercent, err := wfStore.CalculateProgress(ctx, update.WorkflowID)
				if err != nil {
					log.Printf("Failed to calculate progress: %v", err)
					progressPercent = 0
					return
				}

				// Check if workflow is complete
				if progressPercent >= 100 {
					if err := wfStore.UpdateState(ctx, update.WorkflowID, workflow.WorkflowStateCompleted); err != nil {
						log.Printf("Failed to mark workflow as completed: %v", err)
					}
				}

				// Broadcast to WebSocket clients
				wsHandler.BroadcastWorkflowStatus(update.WorkflowID, update.Phase, update.Status, progressPercent)

				// Broadcast agent activity
				wsHandler.BroadcastAgentActivity(
					update.WorkflowID,
					update.JobType,
					update.Status,
					update.Message,
					map[string]interface{}{
						"job_id":   update.JobID,
						"progress": progressPercent,
					},
				)

				log.Printf("Progress update: workflow=%s job=%s status=%s progress=%d%%",
					update.WorkflowID, update.JobID, update.Status, progressPercent)
			})
			if err != nil {
				log.Printf("Progress subscriber error: %v", err)
			}
		}()
		log.Println("Progress subscriber started")

		// Start log subscription in goroutine
		go func() {
			log.Println("Starting log subscriber...")
			logChannel := "workflow_logs"
			pubsub := redisClient.Subscribe(ctx, logChannel)
			defer pubsub.Close()

			// Wait for confirmation
			_, err := pubsub.Receive(ctx)
			if err != nil {
				log.Printf("Failed to subscribe to logs: %v", err)
				return
			}

			log.Printf("Subscribed to logs on channel: %s", logChannel)

			// Listen for log messages
			ch := pubsub.Channel()
			for {
				select {
				case <-ctx.Done():
					return
				case msg := <-ch:
					var logMsg progress.LogMessage
					if err := json.Unmarshal([]byte(msg.Payload), &logMsg); err != nil {
						log.Printf("Failed to unmarshal log message: %v", err)
						continue
					}

					// Broadcast to WebSocket clients
					wsHandler.BroadcastLog(
						logMsg.WorkflowID,
						logMsg.Level,
						logMsg.Message,
						logMsg.Agent,
						logMsg.Details,
					)

					log.Printf("Log: workflow=%s level=%s agent=%s message=%s",
						logMsg.WorkflowID, logMsg.Level, logMsg.Agent, logMsg.Message)
				}
			}
		}()
		log.Println("Log subscriber started")
	}

	// Start server in goroutine
	go func() {
		log.Printf("REST API server listening on %s", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Close Redis connection
	if err := redisClient.Close(); err != nil {
		log.Printf("Error closing Redis connection: %v", err)
	}

	log.Println("Server stopped")
}
