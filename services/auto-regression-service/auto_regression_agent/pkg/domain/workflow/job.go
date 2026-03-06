package workflow

import "time"

// JobType represents the type of job
type JobType string

const (
	JobTypeSpecAnalysis   JobType = "spec_analysis"
	JobTypeTestGeneration JobType = "test_generation"
	JobTypeTestExecution  JobType = "test_execution"
	JobTypeResultAnalysis JobType = "result_analysis"
)

// MarshalBinary implements encoding.BinaryMarshaler
func (jt JobType) MarshalBinary() ([]byte, error) {
	return []byte(jt), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (jt *JobType) UnmarshalBinary(data []byte) error {
	*jt = JobType(data)
	return nil
}

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending   JobStatus = "pending"
	JobStatusQueued    JobStatus = "queued"
	JobStatusRunning   JobStatus = "running"
	JobStatusCompleted JobStatus = "completed"
	JobStatusFailed    JobStatus = "failed"
	JobStatusCancelled JobStatus = "cancelled"
)

// MarshalBinary implements encoding.BinaryMarshaler
func (js JobStatus) MarshalBinary() ([]byte, error) {
	return []byte(js), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (js *JobStatus) UnmarshalBinary(data []byte) error {
	*js = JobStatus(data)
	return nil
}

// JobPriority represents job priority
type JobPriority int

const (
	JobPriorityLow      JobPriority = 1
	JobPriorityMedium   JobPriority = 5
	JobPriorityHigh     JobPriority = 10
	JobPriorityCritical JobPriority = 20
)

// MarshalBinary implements encoding.BinaryMarshaler
func (jp JobPriority) MarshalBinary() ([]byte, error) {
	return []byte{byte(jp)}, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (jp *JobPriority) UnmarshalBinary(data []byte) error {
	if len(data) > 0 {
		*jp = JobPriority(data[0])
	}
	return nil
}

// Job represents a unit of work in the system
type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Status      JobStatus              `json:"status"`
	Priority    JobPriority            `json:"priority"`
	TeamID      string                 `json:"team_id"`
	SpecID      string                 `json:"spec_id,omitempty"`
	EndpointID  string                 `json:"endpoint_id,omitempty"`
	RunMode     RunMode                `json:"run_mode"`
	Payload     map[string]interface{} `json:"payload"`
	Result      map[string]interface{} `json:"result,omitempty"`
	Error       string                 `json:"error,omitempty"`
	Retries     int                    `json:"retries"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	QueuedAt    *time.Time             `json:"queued_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	UpdatedAt   time.Time              `json:"updated_at"`
}

// RunMode represents the test run mode
type RunMode string

const (
	RunModeSmoke   RunMode = "smoke"   // Quick validation, critical endpoints only
	RunModeFull    RunMode = "full"    // Complete test suite
	RunModeNightly RunMode = "nightly" // Extended tests, performance tests
)

// MarshalBinary implements encoding.BinaryMarshaler
func (rm RunMode) MarshalBinary() ([]byte, error) {
	return []byte(rm), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (rm *RunMode) UnmarshalBinary(data []byte) error {
	*rm = RunMode(data)
	return nil
}

// JobFilter represents filter criteria for jobs
type JobFilter struct {
	TeamID        string
	SpecID        string
	Type          JobType
	Status        JobStatus
	RunMode       RunMode
	CreatedAfter  *time.Time
	CreatedBefore *time.Time
	Limit         int
	Offset        int
}

// WorkflowState represents the state of a workflow
type WorkflowState string

const (
	WorkflowStateCreated    WorkflowState = "created"
	WorkflowStateQueued     WorkflowState = "queued"
	WorkflowStateAnalyzing  WorkflowState = "analyzing"
	WorkflowStateGenerating WorkflowState = "generating"
	WorkflowStateValidating WorkflowState = "validating"
	WorkflowStateReady      WorkflowState = "ready"
	WorkflowStateExecuting  WorkflowState = "executing"
	WorkflowStateComparing  WorkflowState = "comparing"
	WorkflowStateReporting  WorkflowState = "reporting"
	WorkflowStateCompleted  WorkflowState = "completed"
	WorkflowStateFailed     WorkflowState = "failed"
	WorkflowStateRetrying   WorkflowState = "retrying"
)

// MarshalBinary implements encoding.BinaryMarshaler
func (ws WorkflowState) MarshalBinary() ([]byte, error) {
	return []byte(ws), nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (ws *WorkflowState) UnmarshalBinary(data []byte) error {
	*ws = WorkflowState(data)
	return nil
}

// Workflow represents a complete test workflow
type Workflow struct {
	ID          string                 `json:"id"`
	SpecID      string                 `json:"spec_id"`
	TeamID      string                 `json:"team_id"`
	RunMode     RunMode                `json:"run_mode"`
	State       WorkflowState          `json:"state"`
	Jobs        []string               `json:"jobs"` // Job IDs
	Metadata    map[string]interface{} `json:"metadata"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// IsTerminal returns true if the job is in a terminal state
func (j *Job) IsTerminal() bool {
	return j.Status == JobStatusCompleted ||
		j.Status == JobStatusFailed ||
		j.Status == JobStatusCancelled
}

// CanRetry returns true if the job can be retried
func (j *Job) CanRetry() bool {
	return j.Status == JobStatusFailed && j.Retries < j.MaxRetries
}

// IsTerminal returns true if the workflow is in a terminal state
func (w *Workflow) IsTerminal() bool {
	return w.State == WorkflowStateCompleted ||
		w.State == WorkflowStateFailed
}
