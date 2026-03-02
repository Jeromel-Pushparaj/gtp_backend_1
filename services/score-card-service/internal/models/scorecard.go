package models

import "time"

// ScoreCard represents a scorecard entity
type ScoreCard struct {
	ID          int64     `json:"id" db:"id"`
	ServiceName string    `json:"service_name" db:"service_name"`
	Score       float64   `json:"score" db:"score"`
	Metrics     Metrics   `json:"metrics" db:"metrics"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" db:"updated_at"`
}

// Metrics represents the metrics used to calculate the score
type Metrics struct {
	CodeQuality      float64 `json:"code_quality"`
	TestCoverage     float64 `json:"test_coverage"`
	SecurityScore    float64 `json:"security_score"`
	PerformanceScore float64 `json:"performance_score"`
	DocumentationScore float64 `json:"documentation_score"`
}

// ScoreCardRequest represents a request to calculate a scorecard
type ScoreCardRequest struct {
	ServiceName string  `json:"service_name" binding:"required"`
	Metrics     Metrics `json:"metrics" binding:"required"`
}

// ScoreCardResponse represents the response after calculating a scorecard
type ScoreCardResponse struct {
	ID          int64     `json:"id"`
	ServiceName string    `json:"service_name"`
	Score       float64   `json:"score"`
	Grade       string    `json:"grade"`
	Metrics     Metrics   `json:"metrics"`
	CreatedAt   time.Time `json:"created_at"`
}

// CalculateScore calculates the overall score from metrics
func (m *Metrics) CalculateScore() float64 {
	// Weighted average of all metrics
	weights := map[string]float64{
		"code_quality":       0.25,
		"test_coverage":      0.25,
		"security_score":     0.20,
		"performance_score":  0.15,
		"documentation_score": 0.15,
	}

	score := (m.CodeQuality * weights["code_quality"]) +
		(m.TestCoverage * weights["test_coverage"]) +
		(m.SecurityScore * weights["security_score"]) +
		(m.PerformanceScore * weights["performance_score"]) +
		(m.DocumentationScore * weights["documentation_score"])

	return score
}

// GetGrade returns the letter grade based on score
func GetGrade(score float64) string {
	switch {
	case score >= 90:
		return "A"
	case score >= 80:
		return "B"
	case score >= 70:
		return "C"
	case score >= 60:
		return "D"
	default:
		return "F"
	}
}

