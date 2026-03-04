package resources

import "time"

// OnboardServiceRequest represents the request body for onboarding a service
type OnboardServiceRequest struct {
	Name         string            `json:"name" binding:"required"`
	Description  string            `json:"description"`
	Team         string            `json:"team" binding:"required"`
	Repository   string            `json:"repository" binding:"required"`
	Lifecycle    string            `json:"lifecycle" binding:"required"`
	Language     string            `json:"language"`
	Integrations map[string]string `json:"integrations"`
	Tags         []string          `json:"tags"`
}

// ServiceResponse represents the response for a service
type ServiceResponse struct {
	ServiceID    string            `json:"service_id"`
	Name         string            `json:"name"`
	Description  string            `json:"description"`
	Team         string            `json:"team"`
	Repository   string            `json:"repository"`
	Lifecycle    string            `json:"lifecycle"`
	Language     string            `json:"language"`
	Integrations map[string]string `json:"integrations"`
	Tags         []string          `json:"tags"`
	OnboardedAt  time.Time         `json:"onboarded_at"`
}

// APIResponse represents a standard API response
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Errors  []string    `json:"errors,omitempty"`
}

