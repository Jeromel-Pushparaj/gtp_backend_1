package storage

import (
	"encoding/json"
	"os"
	"pd-service/models"
	"sync"
	"time"
)


type InMemoryStorage struct {
	services map[string]*models.Service
	mu       sync.RWMutex
	filePath string
}

func NewInMemoryStorage(filePath string) *InMemoryStorage {
	s := &InMemoryStorage{
		services: make(map[string]*models.Service),
		filePath: filePath,
	}
	s.load()
	return s
}

func (s *InMemoryStorage) load() {
	data, err := os.ReadFile(s.filePath)
	if err != nil {
		return // File doesn't exist yet
	}
	
	var services []*models.Service
	if err := json.Unmarshal(data, &services); err != nil {
		return
	}
	
	for _, svc := range services {
		s.services[svc.ID] = svc
	}
}

func (s *InMemoryStorage) save() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.saveUnlocked()
}

// saveUnlocked saves without acquiring a lock (caller must hold lock)
func (s *InMemoryStorage) saveUnlocked() error {
	services := make([]*models.Service, 0, len(s.services))
	for _, svc := range s.services {
		services = append(services, svc)
	}

	data, err := json.MarshalIndent(services, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(s.filePath, data, 0644)
}

func (s *InMemoryStorage) CreateService(service *models.Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	service.CreatedAt = time.Now()
	service.UpdatedAt = time.Now()
	s.services[service.ID] = service

	return s.saveUnlocked()
}

func (s *InMemoryStorage) GetService(id string) (*models.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if svc, ok := s.services[id]; ok {
		return svc, nil
	}
	return nil, nil
}

func (s *InMemoryStorage) GetServiceByPDID(pdServiceID string) (*models.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	for _, svc := range s.services {
		if svc.PDServiceID == pdServiceID {
			return svc, nil
		}
	}
	return nil, nil
}

func (s *InMemoryStorage) ListServices(orgName string) ([]*models.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var result []*models.Service
	for _, svc := range s.services {
		if orgName == "" || svc.OrgName == orgName {
			result = append(result, svc)
		}
	}
	return result, nil
}

func (s *InMemoryStorage) UpdateService(service *models.Service) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	service.UpdatedAt = time.Now()
	s.services[service.ID] = service

	return s.saveUnlocked()
}

func (s *InMemoryStorage) DeleteService(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.services, id)
	return s.saveUnlocked()
}

