package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/knowledge"
	"github.com/google/uuid"
)

// Store manages the knowledge base with thread-safe operations
type Store struct {
	kb           *knowledge.KnowledgeBase
	mu           sync.RWMutex
	storagePath  string
	autoSave     bool
	saveInterval time.Duration
	stopChan     chan struct{}
}

// NewStore creates a new knowledge store
func NewStore(storagePath string, autoSave bool, saveInterval time.Duration) (*Store, error) {
	store := &Store{
		kb: &knowledge.KnowledgeBase{
			SuccessPatterns: make([]knowledge.SuccessPattern, 0),
			FailurePatterns: make([]knowledge.FailurePattern, 0),
			PerformanceData: make([]knowledge.PerformanceData, 0),
			AgentFeedback:   make([]knowledge.AgentFeedback, 0),
			CoverageData: knowledge.CoverageData{
				EndpointCoverage:   make(map[string]knowledge.EndpointCover),
				StatusCodeCoverage: make(map[int]int),
				MethodCoverage:     make(map[string]int),
				UncoveredEndpoints: make([]string, 0),
			},
			AdaptiveSettings: knowledge.AdaptiveSettings{
				FocusOnFailures:        true,
				SkipStableTests:        false,
				StabilityThreshold:     5,
				FailureFocusMultiplier: 2.0,
				RiskBasedPriority:      true,
				EndpointPriorities:     make(map[string]int),
			},
			LastUpdated: time.Now(),
		},
		storagePath:  storagePath,
		autoSave:     autoSave,
		saveInterval: saveInterval,
		stopChan:     make(chan struct{}),
	}

	// Load existing knowledge base if it exists
	if err := store.Load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load knowledge base: %w", err)
		}
	}

	// Start auto-save goroutine if enabled
	if autoSave {
		go store.autoSaveLoop()
	}

	return store, nil
}

// Load loads the knowledge base from disk
func (s *Store) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.storagePath)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, s.kb)
}

// Save saves the knowledge base to disk
func (s *Store) Save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Ensure directory exists
	dir := filepath.Dir(s.storagePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	s.kb.LastUpdated = time.Now()

	data, err := json.MarshalIndent(s.kb, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal knowledge base: %w", err)
	}

	return os.WriteFile(s.storagePath, data, 0644)
}

// autoSaveLoop periodically saves the knowledge base
func (s *Store) autoSaveLoop() {
	ticker := time.NewTicker(s.saveInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := s.Save(); err != nil {
				fmt.Printf("ERROR: Failed to auto-save knowledge base: %v\n", err)
			}
		case <-s.stopChan:
			return
		}
	}
}

// Close stops auto-save and saves final state
func (s *Store) Close() error {
	if s.autoSave {
		close(s.stopChan)
	}
	return s.Save()
}

// RecordLearningEvent processes a learning event and updates knowledge base
func (s *Store) RecordLearningEvent(ctx context.Context, event knowledge.LearningEvent) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Update coverage data
	s.updateCoverageData(event)

	// Update performance data
	s.updatePerformanceData(event)

	// Learn from success or failure
	if event.Success {
		s.learnFromSuccess(event)
	} else {
		s.learnFromFailure(event)
	}

	return nil
}

// updateCoverageData updates coverage metrics
func (s *Store) updateCoverageData(event knowledge.LearningEvent) {
	key := fmt.Sprintf("%s %s", event.Method, event.Endpoint)

	cover, exists := s.kb.CoverageData.EndpointCoverage[key]
	if !exists {
		cover = knowledge.EndpointCover{
			Endpoint:        event.Endpoint,
			Method:          event.Method,
			StatusCodesSeen: make([]int, 0),
		}
		s.kb.CoverageData.TestedEndpoints++
	}

	cover.TestCount++
	cover.LastTested = event.Timestamp

	// Add status code if not seen before
	statusSeen := false
	for _, code := range cover.StatusCodesSeen {
		if code == event.StatusCode {
			statusSeen = true
			break
		}
	}
	if !statusSeen {
		cover.StatusCodesSeen = append(cover.StatusCodesSeen, event.StatusCode)
	}

	s.kb.CoverageData.EndpointCoverage[key] = cover
	s.kb.CoverageData.StatusCodeCoverage[event.StatusCode]++
	s.kb.CoverageData.MethodCoverage[event.Method]++
	s.kb.CoverageData.LastUpdated = time.Now()
}

// updatePerformanceData updates performance metrics
func (s *Store) updatePerformanceData(event knowledge.LearningEvent) {
	key := fmt.Sprintf("%s %s", event.Method, event.Endpoint)

	// Find existing performance data
	var perfData *knowledge.PerformanceData
	for i := range s.kb.PerformanceData {
		if s.kb.PerformanceData[i].EndpointPattern == key {
			perfData = &s.kb.PerformanceData[i]
			break
		}
	}

	if perfData == nil {
		// Create new performance data
		newPerfData := knowledge.PerformanceData{
			ID:              uuid.New().String(),
			EndpointPattern: key,
			Method:          event.Method,
			MinResponseTime: event.ResponseTime,
			MaxResponseTime: event.ResponseTime,
			AvgResponseTime: event.ResponseTime,
			ExecutionCount:  1,
			LastUpdated:     time.Now(),
		}
		s.kb.PerformanceData = append(s.kb.PerformanceData, newPerfData)
		return
	}

	// Update existing performance data
	perfData.ExecutionCount++

	// Update min/max
	if event.ResponseTime < perfData.MinResponseTime {
		perfData.MinResponseTime = event.ResponseTime
	}
	if event.ResponseTime > perfData.MaxResponseTime {
		perfData.MaxResponseTime = event.ResponseTime
	}

	// Update average (running average)
	perfData.AvgResponseTime = time.Duration(
		(int64(perfData.AvgResponseTime)*int64(perfData.ExecutionCount-1) + int64(event.ResponseTime)) /
			int64(perfData.ExecutionCount),
	)

	perfData.LastUpdated = time.Now()

	// Flag as slow if avg response time > 2 seconds
	perfData.IsSlow = perfData.AvgResponseTime > 2*time.Second
}

// learnFromSuccess records successful patterns
func (s *Store) learnFromSuccess(event knowledge.LearningEvent) {
	// Find existing success pattern
	var pattern *knowledge.SuccessPattern
	for i := range s.kb.SuccessPatterns {
		if s.kb.SuccessPatterns[i].EndpointPattern == event.Endpoint &&
			s.kb.SuccessPatterns[i].Method == event.Method &&
			s.kb.SuccessPatterns[i].StatusCode == event.StatusCode {
			pattern = &s.kb.SuccessPatterns[i]
			break
		}
	}

	if pattern == nil {
		// Create new success pattern
		newPattern := knowledge.SuccessPattern{
			ID:              uuid.New().String(),
			EndpointPattern: event.Endpoint,
			Method:          event.Method,
			PayloadTemplate: event.Payload,
			StatusCode:      event.StatusCode,
			ResponseSchema:  event.Response,
			SuccessCount:    1,
			Confidence:      0.5,
			FirstSeen:       event.Timestamp,
			LastSeen:        event.Timestamp,
			Tags:            []string{"learned"},
		}
		s.kb.SuccessPatterns = append(s.kb.SuccessPatterns, newPattern)
		return
	}

	// Update existing pattern
	pattern.SuccessCount++
	pattern.LastSeen = event.Timestamp

	// Increase confidence (max 1.0)
	pattern.Confidence = min(1.0, pattern.Confidence+0.05)
}

// learnFromFailure records failure patterns
func (s *Store) learnFromFailure(event knowledge.LearningEvent) {
	// Find existing failure pattern
	var pattern *knowledge.FailurePattern
	for i := range s.kb.FailurePatterns {
		if s.kb.FailurePatterns[i].EndpointPattern == event.Endpoint &&
			s.kb.FailurePatterns[i].Method == event.Method &&
			s.kb.FailurePatterns[i].ErrorMessage == event.ErrorMessage {
			pattern = &s.kb.FailurePatterns[i]
			break
		}
	}

	if pattern == nil {
		// Create new failure pattern
		newPattern := knowledge.FailurePattern{
			ID:              uuid.New().String(),
			EndpointPattern: event.Endpoint,
			Method:          event.Method,
			PayloadPattern:  event.Payload,
			ErrorMessage:    event.ErrorMessage,
			StatusCode:      event.StatusCode,
			FailureCount:    1,
			Confidence:      0.5,
			FirstSeen:       event.Timestamp,
			LastSeen:        event.Timestamp,
			Tags:            []string{"learned"},
		}
		s.kb.FailurePatterns = append(s.kb.FailurePatterns, newPattern)
		return
	}

	// Update existing pattern
	pattern.FailureCount++
	pattern.LastSeen = event.Timestamp

	// Increase confidence (max 1.0)
	pattern.Confidence = min(1.0, pattern.Confidence+0.05)
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}
