package events

import (
	"time"
)

// EventType represents the type of event
type EventType string

const (
	// Workflow Events
	EventTypeSpecUploaded     EventType = "spec_uploaded"
	EventTypeSpecAnalyzed     EventType = "spec_analyzed"
	EventTypeStrategyProposed EventType = "strategy_proposed"
	EventTypeStrategyApproved EventType = "strategy_approved"
	EventTypePayloadsReady    EventType = "payloads_ready"
	EventTypeTestsComplete    EventType = "tests_complete"
	EventTypeAnalysisComplete EventType = "analysis_complete"

	// Agent Communication Events
	EventTypeAgentMessage  EventType = "agent_message"
	EventTypeAgentRequest  EventType = "agent_request"
	EventTypeAgentResponse EventType = "agent_response"
	EventTypeAgentFeedback EventType = "agent_feedback"

	// Consensus Events
	EventTypeConsensusRequest  EventType = "consensus_request"
	EventTypeConsensusVote     EventType = "consensus_vote"
	EventTypeConsensusDecision EventType = "consensus_decision"

	// Task Events
	EventTypeTaskCreated   EventType = "task_created"
	EventTypeTaskStarted   EventType = "task_started"
	EventTypeTaskCompleted EventType = "task_completed"
	EventTypeTaskFailed    EventType = "task_failed"

	// Learning Events
	EventTypeLearningUpdate  EventType = "learning_update"
	EventTypePatternDetected EventType = "pattern_detected"

	// User Feedback Events
	EventTypeUserFeedback EventType = "user_feedback"
)

// Event represents a system event
type Event struct {
	ID            string                 `json:"id"`
	Type          EventType              `json:"type"`
	Source        string                 `json:"source"`      // Agent or component that created the event
	Target        string                 `json:"target"`      // Target agent (empty for broadcast)
	WorkflowID    string                 `json:"workflow_id"` // Associated workflow
	Payload       map[string]interface{} `json:"payload"`
	Metadata      map[string]interface{} `json:"metadata"`
	Priority      int                    `json:"priority"` // 1-10, higher is more urgent
	Timestamp     time.Time              `json:"timestamp"`
	CorrelationID string                 `json:"correlation_id"` // For tracking related events
}

// Message represents an agent-to-agent message
type Message struct {
	ID            string                 `json:"id"`
	Type          MessageType            `json:"type"`
	From          string                 `json:"from"`
	To            string                 `json:"to"`
	Subject       string                 `json:"subject"`
	Body          string                 `json:"body"`
	Data          map[string]interface{} `json:"data"`
	Priority      int                    `json:"priority"`
	Timestamp     time.Time              `json:"timestamp"`
	ReplyTo       string                 `json:"reply_to"`
	CorrelationID string                 `json:"correlation_id"`
}

// MessageType represents the type of message
type MessageType string

const (
	MessageTypeRequest      MessageType = "request"
	MessageTypeResponse     MessageType = "response"
	MessageTypeNotification MessageType = "notification"
	MessageTypeBroadcast    MessageType = "broadcast"
	MessageTypeFeedback     MessageType = "feedback"
)

// ConsensusRequest represents a request for consensus
type ConsensusRequest struct {
	ID            string                 `json:"id"`
	DecisionType  string                 `json:"decision_type"`
	Question      string                 `json:"question"`
	Options       []string               `json:"options"`
	Context       map[string]interface{} `json:"context"`
	RequiredVotes int                    `json:"required_votes"`
	Timeout       time.Duration          `json:"timeout"`
	CreatedBy     string                 `json:"created_by"`
	CreatedAt     time.Time              `json:"created_at"`
}

// ConsensusVote represents a vote in a consensus decision
type ConsensusVote struct {
	RequestID  string                 `json:"request_id"`
	AgentName  string                 `json:"agent_name"`
	Vote       string                 `json:"vote"`
	Confidence float64                `json:"confidence"`
	Reasoning  string                 `json:"reasoning"`
	Data       map[string]interface{} `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
}

// ConsensusDecision represents the final consensus decision
type ConsensusDecision struct {
	RequestID  string                 `json:"request_id"`
	Decision   string                 `json:"decision"`
	Confidence float64                `json:"confidence"`
	Votes      []ConsensusVote        `json:"votes"`
	Metadata   map[string]interface{} `json:"metadata"`
	DecidedAt  time.Time              `json:"decided_at"`
}

// Task represents a task for an agent
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`
	Description string                 `json:"description"`
	AssignedTo  string                 `json:"assigned_to"`
	Priority    int                    `json:"priority"`
	Status      TaskStatus             `json:"status"`
	Input       map[string]interface{} `json:"input"`
	Output      map[string]interface{} `json:"output"`
	Error       string                 `json:"error,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending   TaskStatus = "pending"
	TaskStatusRunning   TaskStatus = "running"
	TaskStatusCompleted TaskStatus = "completed"
	TaskStatusFailed    TaskStatus = "failed"
	TaskStatusCancelled TaskStatus = "cancelled"
)
