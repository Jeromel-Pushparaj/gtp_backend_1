package service

import (
	"fmt"
	"log"
	"time"

	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/client"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/db"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/models"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/resources"
)

// ServiceService handles business logic for service operations
type ServiceService struct {
	sonarClient *client.SonarClient
	dbService   *db.DatabaseService
	cacheMaxAge time.Duration
}

// NewServiceService creates a new service service
func NewServiceService(sonarClient *client.SonarClient, dbService *db.DatabaseService) *ServiceService {
	return &ServiceService{
		sonarClient: sonarClient,
		dbService:   dbService,
		cacheMaxAge: 5 * time.Minute, // Cache for 5 minutes
	}
}

// FetchServices fetches services for an organization
func (s *ServiceService) FetchServices(orgID int64) ([]resources.ServiceResponse, error) {
	log.Printf("Fetching services for org_id=%d from sonar-shell-test API", orgID)

	// Call sonar-shell-test API
	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// Cache the repositories
	for i := range repos {
		if err := s.dbService.CacheRepository(&repos[i]); err != nil {
			log.Printf("Warning: failed to cache repository %s: %v", repos[i].Name, err)
		}
	}

	// Convert to comprehensive service response DTOs
	var responses []resources.ServiceResponse
	for i := range repos {
		responses = append(responses, s.convertToServiceResponse(&repos[i]))
	}
	return responses, nil
}

// GetService gets a specific service by ID (format: svc_12345), fetches from API if cache miss
func (s *ServiceService) GetService(serviceID string) (*resources.ServiceResponse, error) {
	// 1. Parse service ID to extract repository ID
	// Format: svc_12345 -> extract 12345
	var repoID int64
	_, err := fmt.Sscanf(serviceID, "svc_%d", &repoID)
	if err != nil {
		return nil, fmt.Errorf("invalid service ID format (expected svc_<number>): %w", err)
	}

	log.Printf("Looking up service %s (repository ID: %d)", serviceID, repoID)

	// 2. Check cache first by repository ID
	cachedRepo, err := s.dbService.GetCachedRepositoryByID(repoID)
	if err != nil {
		return nil, fmt.Errorf("failed to check cache: %w", err)
	}

	// 3. If found in cache and not stale, return it
	if cachedRepo != nil {
		isStale, err := s.dbService.IsCacheStale(cachedRepo.Name, s.cacheMaxAge)
		if err != nil {
			log.Printf("Warning: failed to check cache staleness: %v", err)
		}

		if !isStale {
			log.Printf("Returning service %s from cache", serviceID)
			response := s.convertCachedToServiceResponse(cachedRepo)
			return &response, nil
		}
		log.Printf("Service %s cache is stale, fetching from API", serviceID)
	} else {
		log.Printf("Service %s not found in cache, fetching from API", serviceID)
	}

	// 4. Cache is stale or not found, fetch from API
	// Get first available organization
	orgs, err := s.sonarClient.GetOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}

	if len(orgs) == 0 {
		return nil, fmt.Errorf("no organizations available to fetch repository")
	}

	// 5. Fetch all repositories using first org_id
	orgID := orgs[0].ID
	log.Printf("Fetching repositories for org_id=%d to find service %s", orgID, serviceID)

	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// 6. Cache all repositories and find the requested one by ID
	var foundRepo *models.Repository
	for i := range repos {
		if err := s.dbService.CacheRepository(&repos[i]); err != nil {
			log.Printf("Warning: failed to cache repository %s: %v", repos[i].Name, err)
		}
		if repos[i].ID == repoID {
			foundRepo = &repos[i]
		}
	}

	// 7. Return the specific repository
	if foundRepo == nil {
		return nil, fmt.Errorf("service %s not found", serviceID)
	}

	log.Printf("Successfully fetched and cached service %s", serviceID)

	// Convert to comprehensive service response
	response := s.convertToServiceResponse(foundRepo)
	return &response, nil
}

// GetAllServices gets all cached services, fetches from API if cache is empty
func (s *ServiceService) GetAllServices() (*resources.ServicesResponse, error) {
	// 1. Check cache first
	cachedRepos, err := s.dbService.GetAllCachedRepositories()
	if err != nil {
		return nil, fmt.Errorf("failed to get cached services: %w", err)
	}

	// 2. If cache has data, return it
	if len(cachedRepos) > 0 {
		log.Printf("Returning %d services from cache", len(cachedRepos))
		var responses []resources.ServiceResponse
		for _, repo := range cachedRepos {
			responses = append(responses, s.convertCachedToServiceResponse(repo))
		}

		return &resources.ServicesResponse{
			Total:    len(responses),
			Services: responses,
		}, nil
	}

	// 3. Cache is empty, fetch from API
	log.Printf("Cache is empty, fetching from sonar-shell-test API")

	// Get first available organization
	orgs, err := s.sonarClient.GetOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}

	if len(orgs) == 0 {
		log.Printf("No organizations found, returning empty list")
		return &resources.ServicesResponse{
			Total:    0,
			Services: []resources.ServiceResponse{},
		}, nil
	}

	// 4. Fetch repositories using first org_id
	orgID := orgs[0].ID
	log.Printf("Fetching repositories for org_id=%d (%s)", orgID, orgs[0].Name)

	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// 5. Cache the repositories
	for i := range repos {
		if err := s.dbService.CacheRepository(&repos[i]); err != nil {
			log.Printf("Warning: failed to cache repository %s: %v", repos[i].Name, err)
		}
	}

	// 6. Convert to comprehensive service responses
	var responses []resources.ServiceResponse
	for i := range repos {
		responses = append(responses, s.convertToServiceResponse(&repos[i]))
	}
	log.Printf("Successfully fetched and cached %d repositories", len(responses))

	return &resources.ServicesResponse{
		Total:    len(responses),
		Services: responses,
	}, nil
}

// generateServiceID generates a unique service ID from repository ID
func (s *ServiceService) generateServiceID(repoID int64) string {
	return fmt.Sprintf("svc_%d", repoID)
}

// getDummyProductInfo returns hardcoded product information
func (s *ServiceService) getDummyProductInfo() resources.ProductInfo {
	return resources.ProductInfo{
		ID:   "prod_001",
		Name: "Tekion",
	}
}

// getDummyModuleInfo returns hardcoded module information
func (s *ServiceService) getDummyModuleInfo() resources.ModuleInfo {
	return resources.ModuleInfo{
		ID:   "mod_001",
		Name: "Tekion / Task & Program",
	}
}

// getDummyOwnershipInfo returns hardcoded ownership information
func (s *ServiceService) getDummyOwnershipInfo() resources.OwnershipInfo {
	return resources.OwnershipInfo{
		Manager: resources.PersonInfo{
			ID:   "user_101",
			Name: "Arjun J",
		},
		Director: resources.PersonInfo{
			ID:   "user_201",
			Name: "Director Name",
		},
		VP: resources.PersonInfo{
			ID:   "user_301",
			Name: "VP Name",
		},
	}
}

// getDummyJenkinsJobs returns hardcoded Jenkins jobs
func (s *ServiceService) getDummyJenkinsJobs(repoName string) []resources.JenkinsJobInfo {
	now := time.Now()
	creationDate := now.AddDate(0, 0, -30)
	lastUpdate1 := now.AddDate(0, 0, -10)
	lastUpdate2 := now.AddDate(0, 0, -3)

	return []resources.JenkinsJobInfo{
		{
			ID:                 "jenkins_001",
			Title:              repoName,
			Status:             "success",
			LastUpdate:         &lastUpdate1,
			EntityCreationDate: &creationDate,
			URL:                "https://jenkins.company.com/job/1",
		},
		{
			ID:                 "jenkins_002",
			Title:              repoName,
			Status:             "success",
			LastUpdate:         &lastUpdate2,
			EntityCreationDate: &creationDate,
			URL:                "https://jenkins.company.com/job/2",
		},
	}
}

// getDummyWizIssues returns hardcoded Wiz security issues
func (s *ServiceService) getDummyWizIssues(repoName string) []resources.WizIssueInfo {
	detectedAt := time.Now().AddDate(0, 0, -5)

	return []resources.WizIssueInfo{
		{
			ID:         "wiz_001",
			Title:      fmt.Sprintf("%s | COMPUTE_INSTANCE", repoName),
			Severity:   "High",
			Status:     "Open",
			DetectedAt: &detectedAt,
			URL:        "https://wiz.company.com/issues/123",
		},
	}
}

// getDummyPdIncidents returns hardcoded PagerDuty incidents
func (s *ServiceService) getDummyPdIncidents() []resources.PdIncidentInfo {
	createdAt := time.Now().AddDate(0, 0, -7)
	resolvedAt := createdAt.Add(90 * time.Minute)

	return []resources.PdIncidentInfo{
		{
			ID:             "pd_001",
			IncidentNumber: "INC-9001",
			Title:          "Service latency spike",
			Status:         "Resolved",
			Severity:       "SEV-2",
			CreatedAt:      &createdAt,
			ResolvedAt:     &resolvedAt,
			URL:            "https://pagerduty.company.com/incidents/INC-9001",
		},
	}
}

// convertToServiceResponse converts a repository model to comprehensive ServiceResponse
func (s *ServiceService) convertToServiceResponse(repo *models.Repository) resources.ServiceResponse {
	// Generate service ID
	serviceID := s.generateServiceID(repo.ID)

	// Get dummy data
	product := s.getDummyProductInfo()
	module := s.getDummyModuleInfo()
	ownership := s.getDummyOwnershipInfo()
	jenkinsJobs := s.getDummyJenkinsJobs(repo.Name)
	wizIssues := s.getDummyWizIssues(repo.Name)
	pdIncidents := s.getDummyPdIncidents()

	// Initialize metrics (will be populated with real data later)
	metrics := resources.MetricsInfo{
		JiraIssuesCount:         0,
		PullRequestsCount:       0,
		MergeRequestsCount:      0,
		RcaReportsCount:         0,
		JenkinsJobsCount:        len(jenkinsJobs),
		PassingJenkinsJobsCount: 2, // Both dummy jobs are passing
		WizIssuesCount:          len(wizIssues),
		PdIncidentsCount:        len(pdIncidents),
	}

	// Initialize empty arrays for real data (will be populated later)
	openPullRequests := []resources.PullRequestInfo{}
	jiraIssues := []resources.JiraIssueInfo{}

	// Determine language (hardcoded for now, can be enhanced later)
	language := "react" // Default language

	return resources.ServiceResponse{
		ID:                   serviceID,
		Title:                repo.Name,
		RepositorySystem:     "github",
		RepositoryURL:        repo.GitHubURL,
		Language:             language,
		Disposition:          "active",
		Region:               "us",
		CloudMigrationStatus: "cloud-native",
		Product:              product,
		Module:               module,
		Ownership:            ownership,
		Metrics:              metrics,
		OpenPullRequests:     openPullRequests,
		JenkinsJobs:          jenkinsJobs,
		JiraIssues:           jiraIssues,
		WizIssues:            wizIssues,
		PdIncidents:          pdIncidents,
	}
}

// convertCachedToServiceResponse converts a cached repository to comprehensive ServiceResponse
func (s *ServiceService) convertCachedToServiceResponse(repo *models.CachedRepository) resources.ServiceResponse {
	// Generate service ID
	serviceID := s.generateServiceID(repo.RepositoryID)

	// Get dummy data
	product := s.getDummyProductInfo()
	module := s.getDummyModuleInfo()
	ownership := s.getDummyOwnershipInfo()
	jenkinsJobs := s.getDummyJenkinsJobs(repo.Name)
	wizIssues := s.getDummyWizIssues(repo.Name)
	pdIncidents := s.getDummyPdIncidents()

	// Initialize metrics
	metrics := resources.MetricsInfo{
		JiraIssuesCount:         0,
		PullRequestsCount:       0,
		MergeRequestsCount:      0,
		RcaReportsCount:         0,
		JenkinsJobsCount:        len(jenkinsJobs),
		PassingJenkinsJobsCount: 2,
		WizIssuesCount:          len(wizIssues),
		PdIncidentsCount:        len(pdIncidents),
	}

	// Initialize empty arrays
	openPullRequests := []resources.PullRequestInfo{}
	jiraIssues := []resources.JiraIssueInfo{}

	// Determine language
	language := "react"

	return resources.ServiceResponse{
		ID:                   serviceID,
		Title:                repo.Name,
		RepositorySystem:     "github",
		RepositoryURL:        repo.GitHubURL,
		Language:             language,
		Disposition:          "active",
		Region:               "us",
		CloudMigrationStatus: "cloud-native",
		Product:              product,
		Module:               module,
		Ownership:            ownership,
		Metrics:              metrics,
		OpenPullRequests:     openPullRequests,
		JenkinsJobs:          jenkinsJobs,
		JiraIssues:           jiraIssues,
		WizIssues:            wizIssues,
		PdIncidents:          pdIncidents,
	}
}
