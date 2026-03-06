package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/workflow"
)

// Store provides in-memory storage for workflows
// TODO: Replace with persistent database storage
type Store struct {
	workflows map[string]*workflow.Workflow
	mu        sync.RWMutex
}

// NewStore creates a new workflow store
func NewStore() *Store {
	return &Store{
		workflows: make(map[string]*workflow.Workflow),
	}
}

// Save stores a workflow
func (s *Store) Save(ctx context.Context, wf *workflow.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if wf.ID == "" {
		return fmt.Errorf("workflow ID is required")
	}

	s.workflows[wf.ID] = wf
	return nil
}

// Get retrieves a workflow by ID
func (s *Store) Get(ctx context.Context, id string) (*workflow.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wf, exists := s.workflows[id]
	if !exists {
		return nil, fmt.Errorf("workflow not found: %s", id)
	}

	return wf, nil
}

// List retrieves all workflows with optional filtering
func (s *Store) List(ctx context.Context, teamID string, limit, offset int) ([]*workflow.Workflow, int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var filtered []*workflow.Workflow
	for _, wf := range s.workflows {
		if teamID == "" || wf.TeamID == teamID {
			filtered = append(filtered, wf)
		}
	}

	// Sort by created_at descending (newest first)
	for i := 0; i < len(filtered)-1; i++ {
		for j := i + 1; j < len(filtered); j++ {
			if filtered[i].CreatedAt.Before(filtered[j].CreatedAt) {
				filtered[i], filtered[j] = filtered[j], filtered[i]
			}
		}
	}

	total := len(filtered)

	// Apply pagination
	if offset >= len(filtered) {
		return []*workflow.Workflow{}, total, nil
	}

	end := offset + limit
	if end > len(filtered) {
		end = len(filtered)
	}

	return filtered[offset:end], total, nil
}

// Update updates a workflow
func (s *Store) Update(ctx context.Context, wf *workflow.Workflow) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[wf.ID]; !exists {
		return fmt.Errorf("workflow not found: %s", wf.ID)
	}

	wf.UpdatedAt = time.Now()
	s.workflows[wf.ID] = wf
	return nil
}

// Delete removes a workflow
func (s *Store) Delete(ctx context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.workflows[id]; !exists {
		return fmt.Errorf("workflow not found: %s", id)
	}

	delete(s.workflows, id)
	return nil
}

// UpdateState updates workflow state
func (s *Store) UpdateState(ctx context.Context, id string, state workflow.WorkflowState) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wf, exists := s.workflows[id]
	if !exists {
		return fmt.Errorf("workflow not found: %s", id)
	}

	wf.State = state
	wf.UpdatedAt = time.Now()

	// Set completed_at if workflow is completed or failed
	if state == workflow.WorkflowStateCompleted || state == workflow.WorkflowStateFailed {
		now := time.Now()
		wf.CompletedAt = &now
	}

	return nil
}

// GetBySpecID retrieves workflows by spec ID
func (s *Store) GetBySpecID(ctx context.Context, specID string) ([]*workflow.Workflow, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*workflow.Workflow
	for _, wf := range s.workflows {
		if wf.SpecID == specID {
			result = append(result, wf)
		}
	}

	return result, nil
}

// UpdateJobStatus updates a job's status in the workflow's metadata
func (s *Store) UpdateJobStatus(ctx context.Context, workflowID string, jobID string, status workflow.JobStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wf, exists := s.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Initialize job_statuses map if it doesn't exist
	if wf.Metadata == nil {
		wf.Metadata = make(map[string]interface{})
	}

	jobStatuses, ok := wf.Metadata["job_statuses"].(map[string]string)
	if !ok {
		jobStatuses = make(map[string]string)
		wf.Metadata["job_statuses"] = jobStatuses
	}

	// Update job status
	jobStatuses[jobID] = string(status)
	wf.UpdatedAt = time.Now()

	return nil
}

// UpdateJobInfo updates a job's status and type in the workflow's metadata
func (s *Store) UpdateJobInfo(ctx context.Context, workflowID string, jobID string, jobType workflow.JobType, status workflow.JobStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wf, exists := s.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Initialize metadata maps if they don't exist
	if wf.Metadata == nil {
		wf.Metadata = make(map[string]interface{})
	}

	// Update job status
	jobStatuses, ok := wf.Metadata["job_statuses"].(map[string]string)
	if !ok {
		jobStatuses = make(map[string]string)
		wf.Metadata["job_statuses"] = jobStatuses
	}
	jobStatuses[jobID] = string(status)

	// Update job type
	jobTypes, ok := wf.Metadata["job_types"].(map[string]string)
	if !ok {
		jobTypes = make(map[string]string)
		wf.Metadata["job_types"] = jobTypes
	}
	jobTypes[jobID] = string(jobType)

	wf.UpdatedAt = time.Now()

	return nil
}

// UpdateJobResult updates a job's result in the workflow's metadata
func (s *Store) UpdateJobResult(ctx context.Context, workflowID string, jobID string, result map[string]interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	wf, exists := s.workflows[workflowID]
	if !exists {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Initialize metadata maps if they don't exist
	if wf.Metadata == nil {
		wf.Metadata = make(map[string]interface{})
	}

	// Update job results
	jobResults, ok := wf.Metadata["job_results"].(map[string]interface{})
	if !ok {
		jobResults = make(map[string]interface{})
		wf.Metadata["job_results"] = jobResults
	}
	jobResults[jobID] = result

	wf.UpdatedAt = time.Now()

	return nil
}

// CalculateProgress calculates workflow progress based on job completion
func (s *Store) CalculateProgress(ctx context.Context, workflowID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	wf, exists := s.workflows[workflowID]
	if !exists {
		return 0, fmt.Errorf("workflow not found: %s", workflowID)
	}

	// Get job statuses from metadata
	jobStatuses, ok := wf.Metadata["job_statuses"].(map[string]string)
	if !ok || len(jobStatuses) == 0 {
		// No jobs tracked yet, return 0%
		return 0, nil
	}

	// Count total jobs and completed jobs from ACTUAL tracked jobs
	// This includes dynamically created jobs (test execution jobs)
	totalJobs := len(jobStatuses)
	completedJobs := 0

	for _, status := range jobStatuses {
		if status == string(workflow.JobStatusCompleted) || status == string(workflow.JobStatusFailed) {
			completedJobs++
		}
	}

	if totalJobs == 0 {
		return 0, nil
	}

	// Calculate progress percentage
	progress := (completedJobs * 100) / totalJobs

	// Cap at 100%
	if progress > 100 {
		progress = 100
	}

	return progress, nil
}
