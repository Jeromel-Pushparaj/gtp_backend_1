package job

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/api"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// Creator creates jobs from endpoints
type Creator struct {
	maxRetries int
}

// NewCreator creates a new job creator
func NewCreator(maxRetries int) *Creator {
	return &Creator{
		maxRetries: maxRetries,
	}
}

// CreateJobsForSpec creates a complete workflow of jobs for a spec
// Phase 1: Single spec analysis job
// Phase 2: Test generation jobs (one per endpoint)
// Phase 3: Test execution jobs (created after Phase 2 completes)
// Phase 4: Result analysis job (created after Phase 3 completes)
func (c *Creator) CreateJobsForSpec(
	ctx context.Context,
	specID string,
	teamID string,
	endpoints []*api.Endpoint,
	runMode workflow.RunMode,
	format string,
) ([]*workflow.Job, error) {
	var jobs []*workflow.Job

	// Phase 1: Create a single spec analysis job
	// This job will run the Discovery Agent to analyze the entire spec
	specAnalysisJob := c.createSpecAnalysisJob(specID, teamID, runMode, format)
	jobs = append(jobs, specAnalysisJob)

	// Phase 2: Create test generation jobs for each endpoint
	// These jobs will run after spec analysis and generate test manifests
	for _, endpoint := range endpoints {
		// Skip deprecated endpoints in smoke mode
		if runMode == workflow.RunModeSmoke && endpoint.Deprecated {
			continue
		}

		// Determine priority based on endpoint characteristics
		priority := c.determinePriority(endpoint, runMode)

		// Skip low priority endpoints in smoke mode
		if runMode == workflow.RunModeSmoke && priority == workflow.JobPriorityLow {
			continue
		}

		testGenJob := c.createTestGenerationJob(specID, teamID, endpoint, runMode, priority, specAnalysisJob.ID)
		jobs = append(jobs, testGenJob)
	}

	return jobs, nil
}

// createSpecAnalysisJob creates a spec analysis job (Phase 1)
func (c *Creator) createSpecAnalysisJob(
	specID string,
	teamID string,
	runMode workflow.RunMode,
	format string,
) *workflow.Job {
	now := time.Now()
	jobID := fmt.Sprintf("spec-analysis-%s", specID)

	return &workflow.Job{
		ID:       jobID,
		Type:     workflow.JobTypeSpecAnalysis,
		Status:   workflow.JobStatusPending,
		Priority: workflow.JobPriorityHigh, // Spec analysis is always high priority
		TeamID:   teamID,
		SpecID:   specID,
		RunMode:  runMode,
		Payload: map[string]interface{}{
			"spec_id": specID,
			// Always use .json extension since orchestrator converts all specs to OpenAPI 3.0 JSON
			"spec_path":      fmt.Sprintf("./output/specs/%s.json", specID),
			"discovery_path": fmt.Sprintf("./output/discovery/%s-discovery.json", specID),
			"run_mode":       runMode,
		},
		Retries:    0,
		MaxRetries: c.maxRetries,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// createTestGenerationJob creates a test generation job (Phase 2)
func (c *Creator) createTestGenerationJob(
	specID string,
	teamID string,
	endpoint *api.Endpoint,
	runMode workflow.RunMode,
	priority workflow.JobPriority,
	dependsOnJobID string,
) *workflow.Job {
	now := time.Now()
	jobID := c.generateJobID(specID, endpoint.Path, endpoint.Method, runMode)

	return &workflow.Job{
		ID:         jobID,
		Type:       workflow.JobTypeTestGeneration,
		Status:     workflow.JobStatusPending,
		Priority:   priority,
		TeamID:     teamID,
		SpecID:     specID,
		EndpointID: endpoint.ID,
		RunMode:    runMode,
		Payload: map[string]interface{}{
			"spec_id":        specID,
			"endpoint":       endpoint,
			"path":           endpoint.Path,
			"method":         endpoint.Method,
			"discovery_path": fmt.Sprintf("./output/discovery/%s-discovery.json", specID),
			"manifest_path":  fmt.Sprintf("./output/manifests/%s-manifest.json", endpoint.ID),
			"depends_on":     dependsOnJobID,
		},
		Retries:    0,
		MaxRetries: c.maxRetries,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// generateJobID generates a deterministic job ID
// This ensures idempotent job creation - same spec/endpoint/mode = same job ID
func (c *Creator) generateJobID(specID, path, method string, runMode workflow.RunMode) string {
	// Create a deterministic hash
	data := fmt.Sprintf("%s:%s:%s:%s", specID, path, method, runMode)
	hash := sha256.Sum256([]byte(data))
	hashStr := hex.EncodeToString(hash[:])

	// Use first 32 chars of hash as UUID-like ID
	return fmt.Sprintf("%s-%s-%s-%s-%s",
		hashStr[0:8],
		hashStr[8:12],
		hashStr[12:16],
		hashStr[16:20],
		hashStr[20:32],
	)
}

// determinePriority determines job priority based on endpoint characteristics
func (c *Creator) determinePriority(endpoint *api.Endpoint, runMode workflow.RunMode) workflow.JobPriority {
	score := 0

	// Critical methods get higher priority
	switch endpoint.Method {
	case "POST", "PUT", "DELETE":
		score += 3
	case "GET":
		score += 1
	}

	// Endpoints with security requirements are more critical
	if len(endpoint.Security) > 0 {
		score += 2
	}

	// Endpoints with request bodies are more complex
	if endpoint.RequestBody != nil {
		score += 2
	}

	// Count response codes (more responses = more complex)
	if len(endpoint.Responses) > 3 {
		score += 2
	}

	// Deprecated endpoints get lower priority
	if endpoint.Deprecated {
		score -= 3
	}

	// Adjust based on run mode
	switch runMode {
	case workflow.RunModeSmoke:
		// In smoke mode, only critical endpoints
		if score >= 7 {
			return workflow.JobPriorityCritical
		}
		return workflow.JobPriorityLow
	case workflow.RunModeFull:
		// In full mode, balanced priorities
		if score >= 7 {
			return workflow.JobPriorityHigh
		} else if score >= 4 {
			return workflow.JobPriorityMedium
		}
		return workflow.JobPriorityLow
	case workflow.RunModeNightly:
		// In nightly mode, all endpoints are important
		if score >= 7 {
			return workflow.JobPriorityHigh
		}
		return workflow.JobPriorityMedium
	}

	return workflow.JobPriorityMedium
}

// CreateWorkflow creates a workflow for a spec
func (c *Creator) CreateWorkflow(
	ctx context.Context,
	specID string,
	teamID string,
	runMode workflow.RunMode,
	jobIDs []string,
) *workflow.Workflow {
	now := time.Now()

	return &workflow.Workflow{
		ID:      uuid.New().String(),
		SpecID:  specID,
		TeamID:  teamID,
		RunMode: runMode,
		State:   workflow.WorkflowStateCreated,
		Jobs:    jobIDs,
		Metadata: map[string]interface{}{
			"total_jobs": len(jobIDs),
			"run_mode":   runMode,
		},
		CreatedAt: now,
		UpdatedAt: now,
	}
}
