package service

import (
	"strings"
	"time"

	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/db"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/models"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/resources"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/utils"
)

// OnboardingService handles business logic for service onboarding
type OnboardingService struct {
	serviceRepo *db.ServiceRepository
}

// NewOnboardingService creates a new onboarding service
func NewOnboardingService(serviceRepo *db.ServiceRepository) *OnboardingService {
	return &OnboardingService{
		serviceRepo: serviceRepo,
	}
}

// OnboardService onboards a new service
func (s *OnboardingService) OnboardService(req resources.OnboardServiceRequest) (*resources.ServiceResponse, []string) {
	// Validate request
	if errors := s.validateOnboardRequest(req); len(errors) > 0 {
		return nil, errors
	}

	// Create service model
	service := models.Service{
		ID:           utils.GenerateServiceID(),
		Name:         req.Name,
		Description:  req.Description,
		Team:         req.Team,
		Repository:   req.Repository,
		Lifecycle:    req.Lifecycle,
		Language:     req.Language,
		Integrations: req.Integrations,
		Tags:         req.Tags,
		OnboardedAt:  time.Now(),
	}

	// Save to repository
	if err := s.serviceRepo.Create(service); err != nil {
		return nil, []string{"failed to save service"}
	}

	// Convert to response
	response := s.toServiceResponse(service)
	return &response, nil
}

// GetAllServices returns all onboarded services
func (s *OnboardingService) GetAllServices() ([]resources.ServiceResponse, error) {
	services, err := s.serviceRepo.FindAll()
	if err != nil {
		return nil, err
	}

	responses := make([]resources.ServiceResponse, len(services))
	for i, svc := range services {
		responses[i] = s.toServiceResponse(svc)
	}
	return responses, nil
}

// GetServiceByID returns a service by ID
func (s *OnboardingService) GetServiceByID(id string) (*resources.ServiceResponse, error) {
	service, err := s.serviceRepo.FindByID(id)
	if err != nil {
		return nil, err
	}

	response := s.toServiceResponse(*service)
	return &response, nil
}

// GetServiceCount returns the total number of services
func (s *OnboardingService) GetServiceCount() int {
	return s.serviceRepo.Count()
}

// validateOnboardRequest validates the onboard request
func (s *OnboardingService) validateOnboardRequest(req resources.OnboardServiceRequest) []string {
	errors := []string{}

	// Validate name
	if strings.TrimSpace(req.Name) == "" {
		errors = append(errors, "name is required")
	}

	// Validate team
	if strings.TrimSpace(req.Team) == "" {
		errors = append(errors, "team is required")
	}

	// Validate repository
	if strings.TrimSpace(req.Repository) == "" {
		errors = append(errors, "repository is required")
	}

	// Validate lifecycle
	validLifecycles := []string{"development", "staging", "production"}
	if !contains(validLifecycles, req.Lifecycle) {
		errors = append(errors, "lifecycle must be one of: development, staging, production")
	}

	return errors
}

// toServiceResponse converts a service model to response DTO
func (s *OnboardingService) toServiceResponse(service models.Service) resources.ServiceResponse {
	return resources.ServiceResponse{
		ServiceID:    service.ID,
		Name:         service.Name,
		Description:  service.Description,
		Team:         service.Team,
		Repository:   service.Repository,
		Lifecycle:    service.Lifecycle,
		Language:     service.Language,
		Integrations: service.Integrations,
		Tags:         service.Tags,
		OnboardedAt:  service.OnboardedAt,
	}
}

// contains checks if a string slice contains a value
func contains(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}
