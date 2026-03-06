package orchestration

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration/job"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/orchestration/spec"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/storage/queue"
	workflowstore "gitlab.com/tekion/development/toc/poc/opentest/internal/storage/workflow"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/api"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// Orchestrator coordinates spec analysis and job creation
type Orchestrator struct {
	parser        *spec.Parser
	jobCreator    *job.Creator
	queue         queue.Queue
	workflowStore *workflowstore.Store
	eventBus      *events.Bus
}

// NewOrchestrator creates a new orchestrator
func NewOrchestrator(
	parser *spec.Parser,
	jobCreator *job.Creator,
	q queue.Queue,
	workflowStore *workflowstore.Store,
) *Orchestrator {
	return &Orchestrator{
		parser:        parser,
		jobCreator:    jobCreator,
		queue:         q,
		workflowStore: workflowStore,
		eventBus:      nil, // Will be set later if autonomous mode is enabled
	}
}

// SetEventBus sets the event bus for autonomous agent communication
func (o *Orchestrator) SetEventBus(eventBus *events.Bus) {
	o.eventBus = eventBus
}

// ProcessSpec processes an OpenAPI spec and creates jobs
func (o *Orchestrator) ProcessSpec(
	ctx context.Context,
	specID string,
	teamID string,
	content []byte,
	format string,
	runMode workflow.RunMode,
) (*ProcessResult, error) {
	result := &ProcessResult{
		SpecID:    specID,
		TeamID:    teamID,
		RunMode:   runMode,
		StartedAt: time.Now(),
	}

	// Step 1: Parse OpenAPI spec (this converts Swagger 2.0 to OpenAPI 3.0 if needed)
	doc, err := o.parser.Parse(content, format)
	if err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("failed to parse spec: %w", err)
	}
	result.SpecParsed = true

	// Step 2: Save the parsed (and potentially converted) spec to disk for workers to access
	// Always save as JSON since we've converted Swagger 2.0 to OpenAPI 3.0
	specPath := fmt.Sprintf("./output/specs/%s.json", specID)
	if err := o.saveOpenAPISpec(specPath, doc); err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("failed to save spec: %w", err)
	}

	// Step 3: Extract endpoints
	endpoints, err := o.parser.ExtractEndpoints(doc, specID)
	if err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("failed to extract endpoints: %w", err)
	}
	result.EndpointsExtracted = len(endpoints)

	// Step 3: Extract metadata
	metadata := o.parser.ExtractMetadata(doc, specID)
	result.Metadata = metadata

	// Step 4: Create jobs for endpoints
	jobs, err := o.jobCreator.CreateJobsForSpec(ctx, specID, teamID, endpoints, runMode, format)
	if err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("failed to create jobs: %w", err)
	}
	result.JobsCreated = len(jobs)

	// Step 5: Create workflow BEFORE enqueuing jobs
	jobIDs := make([]string, len(jobs))
	for i, j := range jobs {
		jobIDs[i] = j.ID
	}
	wf := o.jobCreator.CreateWorkflow(ctx, specID, teamID, runMode, jobIDs)

	// Add spec metadata to workflow
	if metadata != nil {
		wf.Metadata["spec_name"] = metadata.Title
		wf.Metadata["spec_version"] = metadata.Version
		wf.Metadata["endpoints_count"] = len(endpoints)
	}

	result.WorkflowID = wf.ID

	// Step 6: Save workflow to store
	if err := o.workflowStore.Save(ctx, wf); err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("failed to save workflow: %w", err)
	}

	// Step 6.5: Publish spec_uploaded event for autonomous agents (if event bus is available)
	if o.eventBus != nil {
		event := &events.Event{
			Type:       events.EventTypeSpecUploaded,
			Source:     "orchestrator",
			WorkflowID: wf.ID,
			Payload: map[string]interface{}{
				"spec_id":   specID,
				"team_id":   teamID,
				"spec_path": specPath,
				"format":    format,
				"run_mode":  string(runMode),
			},
			Priority:  10,
			Timestamp: time.Now(),
		}

		if err := o.eventBus.Publish(ctx, event); err != nil {
			// Log error but don't fail the whole process
			fmt.Printf("Warning: failed to publish spec_uploaded event: %v\n", err)
		}
	}

	// Step 7: Add workflow_id to all jobs
	for _, job := range jobs {
		if job.Payload == nil {
			job.Payload = make(map[string]interface{})
		}
		job.Payload["workflow_id"] = wf.ID
	}

	// Step 8: Enqueue jobs with workflow_id (idempotent - same job ID won't duplicate)
	if err := o.queue.EnqueueBatch(ctx, jobs); err != nil {
		result.Error = err.Error()
		result.CompletedAt = time.Now()
		return result, fmt.Errorf("failed to enqueue jobs: %w", err)
	}
	result.JobsEnqueued = len(jobs)

	result.CompletedAt = time.Now()
	result.Success = true

	return result, nil
}

// ProcessResult contains the result of spec processing
type ProcessResult struct {
	SpecID             string            `json:"spec_id"`
	TeamID             string            `json:"team_id"`
	RunMode            workflow.RunMode  `json:"run_mode"`
	WorkflowID         string            `json:"workflow_id"`
	SpecParsed         bool              `json:"spec_parsed"`
	EndpointsExtracted int               `json:"endpoints_extracted"`
	JobsCreated        int               `json:"jobs_created"`
	JobsEnqueued       int               `json:"jobs_enqueued"`
	Metadata           *api.SpecMetadata `json:"metadata,omitempty"`
	Success            bool              `json:"success"`
	Error              string            `json:"error,omitempty"`
	StartedAt          time.Time         `json:"started_at"`
	CompletedAt        time.Time         `json:"completed_at"`
}

// GetQueueStats returns queue statistics
func (o *Orchestrator) GetQueueStats(ctx context.Context) (*QueueStats, error) {
	queueLen, err := o.queue.GetQueueLength(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get queue length: %w", err)
	}

	pendingCount, err := o.queue.GetPendingJobs(ctx, "")
	if err != nil {
		return nil, fmt.Errorf("failed to get pending count: %w", err)
	}

	return &QueueStats{
		QueueLength:  queueLen,
		PendingCount: pendingCount,
		Timestamp:    time.Now(),
	}, nil
}

// QueueStats contains queue statistics
type QueueStats struct {
	QueueLength  int64     `json:"queue_length"`
	PendingCount int64     `json:"pending_count"`
	Timestamp    time.Time `json:"timestamp"`
}

// saveSpecToDisk saves spec content to disk
func (o *Orchestrator) saveSpecToDisk(path string, content []byte) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Write spec to file
	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}

// saveOpenAPISpec saves an OpenAPI 3.0 document to disk as JSON
func (o *Orchestrator) saveOpenAPISpec(path string, doc *openapi3.T) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	// Marshal the OpenAPI document to JSON
	content, err := doc.MarshalJSON()
	if err != nil {
		return fmt.Errorf("failed to marshal OpenAPI spec: %w", err)
	}

	if err := os.WriteFile(path, content, 0644); err != nil {
		return fmt.Errorf("failed to write spec file: %w", err)
	}

	return nil
}

// ValidateSpec validates an OpenAPI spec without creating jobs
func (o *Orchestrator) ValidateSpec(
	ctx context.Context,
	content []byte,
	format string,
) (*ValidationResult, error) {
	result := &ValidationResult{
		Valid: false,
	}

	// Parse spec
	doc, err := o.parser.Parse(content, format)
	if err != nil {
		result.Errors = append(result.Errors, err.Error())
		return result, nil
	}

	// Extract basic info
	result.Valid = true
	result.Title = doc.Info.Title
	result.Version = doc.Info.Version
	result.EndpointCount = countEndpoints(doc)

	return result, nil
}

// ValidationResult contains spec validation result
type ValidationResult struct {
	Valid         bool     `json:"valid"`
	Title         string   `json:"title,omitempty"`
	Version       string   `json:"version,omitempty"`
	EndpointCount int      `json:"endpoint_count,omitempty"`
	Errors        []string `json:"errors,omitempty"`
}

// countEndpoints counts total endpoints in spec
func countEndpoints(doc *openapi3.T) int {
	count := 0
	if doc.Paths != nil {
		for _, pathItem := range doc.Paths {
			if pathItem != nil {
				count += len(pathItem.Operations())
			}
		}
	}
	return count
}

// GetWorkflowStore returns the workflow store
func (o *Orchestrator) GetWorkflowStore() *workflowstore.Store {
	return o.workflowStore
}
