package knowledge

import (
	"fmt"
	"strings"

	"gitlab.com/tekion/development/toc/poc/opentest/pkg/domain/knowledge"
)

// GetSuccessPatterns returns success patterns for an endpoint
func (s *Store) GetSuccessPatterns(endpoint, method string) []knowledge.SuccessPattern {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patterns := make([]knowledge.SuccessPattern, 0)
	for _, pattern := range s.kb.SuccessPatterns {
		if matchesEndpoint(pattern.EndpointPattern, endpoint) && pattern.Method == method {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// GetFailurePatterns returns failure patterns for an endpoint
func (s *Store) GetFailurePatterns(endpoint, method string) []knowledge.FailurePattern {
	s.mu.RLock()
	defer s.mu.RUnlock()

	patterns := make([]knowledge.FailurePattern, 0)
	for _, pattern := range s.kb.FailurePatterns {
		if matchesEndpoint(pattern.EndpointPattern, endpoint) && pattern.Method == method {
			patterns = append(patterns, pattern)
		}
	}
	return patterns
}

// GetPerformanceData returns performance data for an endpoint
func (s *Store) GetPerformanceData(endpoint, method string) *knowledge.PerformanceData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s %s", method, endpoint)
	for _, perfData := range s.kb.PerformanceData {
		if perfData.EndpointPattern == key {
			return &perfData
		}
	}
	return nil
}

// GetCoverageData returns the current coverage data
func (s *Store) GetCoverageData() knowledge.CoverageData {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.kb.CoverageData
}

// GetAdaptiveSettings returns the current adaptive settings
func (s *Store) GetAdaptiveSettings() knowledge.AdaptiveSettings {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.kb.AdaptiveSettings
}

// UpdateAdaptiveSettings updates adaptive settings
func (s *Store) UpdateAdaptiveSettings(settings knowledge.AdaptiveSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.kb.AdaptiveSettings = settings
}

// AddAgentFeedback adds feedback from one agent to another
func (s *Store) AddAgentFeedback(feedback knowledge.AgentFeedback) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.kb.AgentFeedback = append(s.kb.AgentFeedback, feedback)
}

// GetPendingFeedback returns unacknowledged feedback for an agent
func (s *Store) GetPendingFeedback(agentName string) []knowledge.AgentFeedback {
	s.mu.RLock()
	defer s.mu.RUnlock()

	feedback := make([]knowledge.AgentFeedback, 0)
	for _, fb := range s.kb.AgentFeedback {
		if fb.ToAgent == agentName && !fb.Acknowledged {
			feedback = append(feedback, fb)
		}
	}
	return feedback
}

// AcknowledgeFeedback marks feedback as acknowledged
func (s *Store) AcknowledgeFeedback(feedbackID, actionTaken string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.kb.AgentFeedback {
		if s.kb.AgentFeedback[i].ID == feedbackID {
			s.kb.AgentFeedback[i].Acknowledged = true
			s.kb.AgentFeedback[i].ActionTaken = actionTaken
			break
		}
	}
}

// ShouldSkipTest determines if a test should be skipped based on adaptive settings
func (s *Store) ShouldSkipTest(endpoint, method string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if !s.kb.AdaptiveSettings.SkipStableTests {
		return false
	}

	key := fmt.Sprintf("%s %s", method, endpoint)
	cover, exists := s.kb.CoverageData.EndpointCoverage[key]
	if !exists {
		return false
	}

	return cover.IsStable && cover.StabilityScore >= 0.9
}

// GetTestPriority returns the priority for testing an endpoint
func (s *Store) GetTestPriority(endpoint, method string) int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	key := fmt.Sprintf("%s %s", method, endpoint)

	// Check explicit priority
	if priority, exists := s.kb.AdaptiveSettings.EndpointPriorities[key]; exists {
		return priority
	}

	// Calculate priority based on stability and failures
	cover, exists := s.kb.CoverageData.EndpointCoverage[key]
	if !exists {
		return 5 // Default medium priority for untested endpoints
	}

	if !cover.IsStable {
		return 8 // High priority for unstable endpoints
	}

	return 3 // Low priority for stable endpoints
}

// matchesEndpoint checks if an endpoint pattern matches an endpoint
func matchesEndpoint(pattern, endpoint string) bool {
	// Simple pattern matching - can be enhanced with regex
	if pattern == endpoint {
		return true
	}

	// Handle path parameters like /users/{id}
	patternParts := strings.Split(pattern, "/")
	endpointParts := strings.Split(endpoint, "/")

	if len(patternParts) != len(endpointParts) {
		return false
	}

	for i := range patternParts {
		if patternParts[i] != endpointParts[i] {
			// Check if it's a path parameter
			if strings.HasPrefix(patternParts[i], "{") && strings.HasSuffix(patternParts[i], "}") {
				continue
			}
			return false
		}
	}

	return true
}
