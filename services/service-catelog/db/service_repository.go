package db

import (
	"errors"
	"sync"

	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/models"
)

// ServiceRepository handles data access for services
type ServiceRepository struct {
	services []models.Service
	mu       sync.RWMutex
}

// NewServiceRepository creates a new service repository
func NewServiceRepository() *ServiceRepository {
	return &ServiceRepository{
		services: make([]models.Service, 0),
	}
}

// Create adds a new service to the repository
func (r *ServiceRepository) Create(service models.Service) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.services = append(r.services, service)
	return nil
}

// FindAll returns all services
func (r *ServiceRepository) FindAll() ([]models.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	// Return a copy to prevent external modifications
	result := make([]models.Service, len(r.services))
	copy(result, r.services)
	return result, nil
}

// FindByID returns a service by ID
func (r *ServiceRepository) FindByID(id string) (*models.Service, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, svc := range r.services {
		if svc.ID == id {
			// Return a copy
			serviceCopy := svc
			return &serviceCopy, nil
		}
	}
	return nil, errors.New("service not found")
}

// Count returns the total number of services
func (r *ServiceRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.services)
}

