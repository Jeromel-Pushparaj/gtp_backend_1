package models

import "time"

// Analysis represents a SonarCloud analysis
type Analysis struct {
	Key            string    `json:"key"`
	Date           time.Time `json:"date"`
	ProjectVersion string    `json:"projectVersion"`
	Revision       string    `json:"revision"`
}

// QualityGateStatus represents the quality gate status
type QualityGateStatus struct {
	Status     string `json:"status"` // OK, WARN, ERROR
	Conditions []struct {
		Status         string `json:"status"`
		MetricKey      string `json:"metricKey"`
		Comparator     string `json:"comparator"`
		ErrorThreshold string `json:"errorThreshold"`
		ActualValue    string `json:"actualValue"`
	} `json:"conditions"`
}

// Measure represents a SonarCloud measure
type Measure struct {
	Metric string `json:"metric"`
	Value  string `json:"value"`
}

// IssuesResponse represents the issues API response
type IssuesResponse struct {
	Total  int     `json:"total"`
	Issues []Issue `json:"issues"`
}

// Issue represents a SonarCloud issue
type Issue struct {
	Key       string `json:"key"`
	Rule      string `json:"rule"`
	Severity  string `json:"severity"`
	Component string `json:"component"`
	Line      int    `json:"line"`
	Message   string `json:"message"`
	Type      string `json:"type"` // BUG, VULNERABILITY, CODE_SMELL
	Status    string `json:"status"`
}

