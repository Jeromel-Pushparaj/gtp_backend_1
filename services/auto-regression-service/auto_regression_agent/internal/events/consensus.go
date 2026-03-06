package events

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

// ConsensusEngine manages consensus decisions among agents
type ConsensusEngine struct {
	client *redis.Client
	bus    *Bus
}

// NewConsensusEngine creates a new consensus engine
func NewConsensusEngine(client *redis.Client, bus *Bus) *ConsensusEngine {
	return &ConsensusEngine{
		client: client,
		bus:    bus,
	}
}

// RequestConsensus requests consensus from multiple agents
func (ce *ConsensusEngine) RequestConsensus(ctx context.Context, req *ConsensusRequest) error {
	// Set request ID if not set
	if req.ID == "" {
		req.ID = uuid.New().String()
	}

	// Set timestamp if not set
	if req.CreatedAt.IsZero() {
		req.CreatedAt = time.Now()
	}

	// Store request in Redis
	reqKey := fmt.Sprintf("consensus:%s:request", req.ID)
	data, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	if err := ce.client.Set(ctx, reqKey, data, req.Timeout).Err(); err != nil {
		return fmt.Errorf("failed to store request: %w", err)
	}

	// Publish consensus request event
	event := &Event{
		Type:   EventTypeConsensusRequest,
		Source: req.CreatedBy,
		Payload: map[string]interface{}{
			"request_id":    req.ID,
			"decision_type": req.DecisionType,
			"question":      req.Question,
			"options":       req.Options,
			"context":       req.Context,
		},
		Priority:  8,
		Timestamp: time.Now(),
	}

	if err := ce.bus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish consensus request: %w", err)
	}

	log.Printf("🗳️  Consensus requested: id=%s, type=%s, question=%s",
		req.ID, req.DecisionType, req.Question)

	return nil
}

// SubmitVote submits a vote for a consensus request
func (ce *ConsensusEngine) SubmitVote(ctx context.Context, vote *ConsensusVote) error {
	// Set timestamp if not set
	if vote.Timestamp.IsZero() {
		vote.Timestamp = time.Now()
	}

	// Store vote in Redis hash
	voteKey := fmt.Sprintf("consensus:%s:votes", vote.RequestID)
	data, err := json.Marshal(vote)
	if err != nil {
		return fmt.Errorf("failed to marshal vote: %w", err)
	}

	if err := ce.client.HSet(ctx, voteKey, vote.AgentName, data).Err(); err != nil {
		return fmt.Errorf("failed to store vote: %w", err)
	}

	// Set expiry (7 days)
	ce.client.Expire(ctx, voteKey, 7*24*time.Hour)

	// Publish vote event
	event := &Event{
		Type:   EventTypeConsensusVote,
		Source: vote.AgentName,
		Payload: map[string]interface{}{
			"request_id": vote.RequestID,
			"vote":       vote.Vote,
			"confidence": vote.Confidence,
			"reasoning":  vote.Reasoning,
		},
		Priority:  7,
		Timestamp: time.Now(),
	}

	if err := ce.bus.Publish(ctx, event); err != nil {
		return fmt.Errorf("failed to publish vote event: %w", err)
	}

	log.Printf("🗳️  Vote submitted: request=%s, agent=%s, vote=%s, confidence=%.2f",
		vote.RequestID, vote.AgentName, vote.Vote, vote.Confidence)

	return nil
}

// GetVotes retrieves all votes for a consensus request
func (ce *ConsensusEngine) GetVotes(ctx context.Context, requestID string) ([]ConsensusVote, error) {
	voteKey := fmt.Sprintf("consensus:%s:votes", requestID)

	result, err := ce.client.HGetAll(ctx, voteKey).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get votes: %w", err)
	}

	votes := make([]ConsensusVote, 0, len(result))
	for _, data := range result {
		var vote ConsensusVote
		if err := json.Unmarshal([]byte(data), &vote); err != nil {
			log.Printf("Error parsing vote: %v", err)
			continue
		}
		votes = append(votes, vote)
	}

	return votes, nil
}

// MakeDecision makes a consensus decision based on votes
func (ce *ConsensusEngine) MakeDecision(ctx context.Context, requestID string) (*ConsensusDecision, error) {
	votes, err := ce.GetVotes(ctx, requestID)
	if err != nil {
		return nil, err
	}

	if len(votes) == 0 {
		return nil, fmt.Errorf("no votes found for request %s", requestID)
	}

	// Calculate consensus (simple majority with weighted confidence)
	voteCounts := make(map[string]float64)
	for _, vote := range votes {
		voteCounts[vote.Vote] += vote.Confidence
	}

	// Find decision with highest weighted score
	var decision string
	var maxScore float64
	for vote, score := range voteCounts {
		if score > maxScore {
			maxScore = score
			decision = vote
		}
	}

	// Calculate average confidence
	avgConfidence := maxScore / float64(len(votes))

	consensusDecision := &ConsensusDecision{
		RequestID:  requestID,
		Decision:   decision,
		Confidence: avgConfidence,
		Votes:      votes,
		DecidedAt:  time.Now(),
	}

	// Store decision
	decisionKey := fmt.Sprintf("consensus:%s:decision", requestID)
	data, err := json.Marshal(consensusDecision)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal decision: %w", err)
	}

	if err := ce.client.Set(ctx, decisionKey, data, 7*24*time.Hour).Err(); err != nil {
		return nil, fmt.Errorf("failed to store decision: %w", err)
	}

	// Publish decision event
	event := &Event{
		Type:   EventTypeConsensusDecision,
		Source: "consensus_engine",
		Payload: map[string]interface{}{
			"request_id": requestID,
			"decision":   decision,
			"confidence": avgConfidence,
			"vote_count": len(votes),
		},
		Priority:  9,
		Timestamp: time.Now(),
	}

	if err := ce.bus.Publish(ctx, event); err != nil {
		log.Printf("Warning: failed to publish decision event: %v", err)
	}

	log.Printf("✅ Consensus reached: request=%s, decision=%s, confidence=%.2f, votes=%d",
		requestID, decision, avgConfidence, len(votes))

	return consensusDecision, nil
}
