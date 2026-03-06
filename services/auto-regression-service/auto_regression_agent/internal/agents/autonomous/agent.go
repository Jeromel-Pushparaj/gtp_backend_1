package autonomous

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"gitlab.com/tekion/development/toc/poc/opentest/internal/events"
	"gitlab.com/tekion/development/toc/poc/opentest/pkg/llm"
)

// AgentType represents the type of autonomous agent
type AgentType string

const (
	AgentTypeDiscovery AgentType = "discovery"
	AgentTypeDesigner  AgentType = "designer"
	AgentTypePayload   AgentType = "payload"
	AgentTypeExecutor  AgentType = "executor"
	AgentTypeAnalyzer  AgentType = "analyzer"
	AgentTypeHealing   AgentType = "healing"
)

// AgentState represents the current state of an agent
type AgentState string

const (
	AgentStateIdle       AgentState = "idle"
	AgentStateProcessing AgentState = "processing"
	AgentStateWaiting    AgentState = "waiting"
	AgentStateError      AgentState = "error"
)

// Goal represents an agent's goal
type Goal struct {
	ID          string
	Description string
	Priority    int
	Status      string
	CreatedAt   time.Time
	CompletedAt *time.Time
}

// Memory represents agent's memory
type Memory struct {
	ShortTerm map[string]interface{} // Recent context
	LongTerm  map[string]interface{} // Learned patterns
	mu        sync.RWMutex
}

// NewMemory creates a new memory instance
func NewMemory() *Memory {
	return &Memory{
		ShortTerm: make(map[string]interface{}),
		LongTerm:  make(map[string]interface{}),
	}
}

// Store stores a value in short-term memory
func (m *Memory) Store(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.ShortTerm[key] = value
}

// Recall recalls a value from memory
func (m *Memory) Recall(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Check short-term first
	if val, ok := m.ShortTerm[key]; ok {
		return val, true
	}

	// Check long-term
	if val, ok := m.LongTerm[key]; ok {
		return val, true
	}

	return nil, false
}

// Learn stores a value in long-term memory
func (m *Memory) Learn(key string, value interface{}) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.LongTerm[key] = value
}

// Agent represents an autonomous AI agent
type Agent struct {
	ID           string
	Name         string
	Type         AgentType
	Capabilities []string

	// AI Components
	LLMClient *llm.Client
	Memory    *Memory
	Goals     []Goal

	// Communication
	EventBus   *events.Bus
	MessageBus *events.MessageBus
	Consensus  *events.ConsensusEngine

	// State Management
	State       AgentState
	CurrentTask *events.Task
	mu          sync.RWMutex

	// Context
	ctx    context.Context
	cancel context.CancelFunc
}

// NewAgent creates a new autonomous agent
func NewAgent(
	name string,
	agentType AgentType,
	capabilities []string,
	llmClient *llm.Client,
	eventBus *events.Bus,
	messageBus *events.MessageBus,
	consensus *events.ConsensusEngine,
) *Agent {
	ctx, cancel := context.WithCancel(context.Background())

	return &Agent{
		ID:           uuid.New().String(),
		Name:         name,
		Type:         agentType,
		Capabilities: capabilities,
		LLMClient:    llmClient,
		Memory:       NewMemory(),
		Goals:        make([]Goal, 0),
		EventBus:     eventBus,
		MessageBus:   messageBus,
		Consensus:    consensus,
		State:        AgentStateIdle,
		ctx:          ctx,
		cancel:       cancel,
	}
}

// Start starts the agent
func (a *Agent) Start(ctx context.Context) error {
	log.Printf("🤖 Starting autonomous agent: name=%s, type=%s", a.Name, a.Type)

	a.setState(AgentStateIdle)

	// Start message listener
	go a.listenToMessages(ctx)

	// Start broadcast listener
	go a.listenToBroadcasts(ctx)

	log.Printf("✅ Agent started: %s", a.Name)
	return nil
}

// Stop stops the agent
func (a *Agent) Stop() {
	log.Printf("🛑 Stopping agent: %s", a.Name)
	a.cancel()
}

// setState sets the agent's state
func (a *Agent) setState(state AgentState) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.State = state
	log.Printf("🔄 Agent %s state changed: %s", a.Name, state)
}

// GetState gets the agent's current state
func (a *Agent) GetState() AgentState {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.State
}

// PublishEvent publishes an event to the event bus
func (a *Agent) PublishEvent(ctx context.Context, eventType events.EventType, payload map[string]interface{}) error {
	event := &events.Event{
		Type:      eventType,
		Source:    a.Name,
		Payload:   payload,
		Priority:  5,
		Timestamp: time.Now(),
	}

	return a.EventBus.Publish(ctx, event)
}

// SendMessage sends a message to another agent
func (a *Agent) SendMessage(ctx context.Context, to string, subject string, body string, data map[string]interface{}) error {
	msg := &events.Message{
		From:      a.Name,
		To:        to,
		Subject:   subject,
		Body:      body,
		Data:      data,
		Type:      events.MessageTypeRequest,
		Priority:  5,
		Timestamp: time.Now(),
	}

	return a.MessageBus.SendMessage(ctx, msg)
}

// BroadcastMessage broadcasts a message to all agents
func (a *Agent) BroadcastMessage(ctx context.Context, subject string, body string, data map[string]interface{}) error {
	msg := &events.Message{
		From:      a.Name,
		Subject:   subject,
		Body:      body,
		Data:      data,
		Type:      events.MessageTypeBroadcast,
		Priority:  5,
		Timestamp: time.Now(),
	}

	return a.MessageBus.BroadcastMessage(ctx, msg)
}

// RequestConsensus requests consensus from other agents
func (a *Agent) RequestConsensus(ctx context.Context, question string, options []string, context map[string]interface{}) (string, error) {
	req := &events.ConsensusRequest{
		DecisionType:  "strategy",
		Question:      question,
		Options:       options,
		Context:       context,
		RequiredVotes: 1, // Reduced from 3 to avoid rate limits
		Timeout:       5 * time.Minute,
		CreatedBy:     a.Name,
	}

	if err := a.Consensus.RequestConsensus(ctx, req); err != nil {
		return "", err
	}

	// Poll for votes with timeout
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	timeout := time.After(req.Timeout)

	log.Printf("🗳️  Waiting for consensus votes (need %d votes, timeout: %v)...", req.RequiredVotes, req.Timeout)

	for {
		select {
		case <-timeout:
			// Timeout reached, try to make decision with whatever votes we have
			log.Printf("⏱️  Consensus timeout reached, making decision with available votes")
			decision, err := a.Consensus.MakeDecision(ctx, req.ID)
			if err != nil {
				return "", fmt.Errorf("consensus timeout and failed to make decision: %w", err)
			}
			return decision.Decision, nil

		case <-ticker.C:
			// Check if we have enough votes
			votes, err := a.Consensus.GetVotes(ctx, req.ID)
			if err != nil {
				log.Printf("Warning: failed to get votes: %v", err)
				continue
			}

			if len(votes) >= req.RequiredVotes {
				// Got enough votes, make decision
				log.Printf("✅ Got %d/%d votes, making consensus decision", len(votes), req.RequiredVotes)
				decision, err := a.Consensus.MakeDecision(ctx, req.ID)
				if err != nil {
					return "", fmt.Errorf("failed to make consensus decision: %w", err)
				}
				return decision.Decision, nil
			}

			// Log progress
			if len(votes) > 0 {
				log.Printf("🗳️  Consensus progress: %d/%d votes received", len(votes), req.RequiredVotes)
			}
		}
	}
}

// SubmitVote submits a vote for a consensus request
func (a *Agent) SubmitVote(ctx context.Context, requestID string, vote string, confidence float64, reasoning string) error {
	v := &events.ConsensusVote{
		RequestID:  requestID,
		AgentName:  a.Name,
		Vote:       vote,
		Confidence: confidence,
		Reasoning:  reasoning,
		Timestamp:  time.Now(),
	}

	return a.Consensus.SubmitVote(ctx, v)
}

// listenToMessages listens to direct messages
func (a *Agent) listenToMessages(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			messages, err := a.MessageBus.ReceiveMessages(ctx, a.Name, 10, 1*time.Second)
			if err != nil {
				log.Printf("Error receiving messages: %v", err)
				continue
			}

			for _, msg := range messages {
				a.handleMessage(ctx, msg)
			}
		}
	}
}

// listenToBroadcasts listens to broadcast messages
func (a *Agent) listenToBroadcasts(ctx context.Context) {
	err := a.MessageBus.SubscribeToBroadcast(ctx, func(msg *events.Message) error {
		// Don't process own broadcasts
		if msg.From == a.Name {
			return nil
		}

		return a.handleMessage(ctx, msg)
	})

	if err != nil && err != context.Canceled {
		log.Printf("Error subscribing to broadcasts: %v", err)
	}
}

// handleMessage handles incoming messages (to be overridden by specific agents)
func (a *Agent) handleMessage(ctx context.Context, msg *events.Message) error {
	log.Printf("📬 Agent %s received message: from=%s, subject=%s", a.Name, msg.From, msg.Subject)

	// Store in short-term memory
	a.Memory.Store(fmt.Sprintf("message:%s", msg.ID), msg)

	return nil
}
