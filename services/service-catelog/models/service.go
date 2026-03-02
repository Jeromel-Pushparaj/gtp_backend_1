package models

import "time"

// Service represents a service in the catalog
type Service struct {
	ID           string            `json:"service_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Team         string            `json:"team"`
	Repository   string            `json:"repository"`
	Lifecycle    string            `json:"lifecycle"` // development, staging, production
	Language     string            `json:"language"`
	Integrations map[string]string `json:"integrations"`
	Tags         []string          `json:"tags"`
	OnboardedAt  time.Time         `json:"onboarded_at"`
}

// OnboardRequest represents the request body for onboarding a service
type OnboardRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Team         string            `json:"team" binding:"required"`
	Repository   string            `json:"repository" binding:"required"`
	Lifecycle    string            `json:"lifecycle" binding:"required"`
	Language     string            `json:"language"`
	Integrations map[string]string `json:"integrations"`
	Tags         []string          `json:"tags"`
}

