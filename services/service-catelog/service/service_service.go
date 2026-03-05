package service

import (
	"fmt"
	"log"

	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/client"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/models"
	"github.com/jeromelp/gtp_backend_1/services/service-catalogue/resources"
)

// ServiceService handles business logic for service operations
type ServiceService struct {
	sonarClient *client.SonarClient
}

// NewServiceService creates a new service service
func NewServiceService(sonarClient *client.SonarClient) *ServiceService {
	return &ServiceService{
		sonarClient: sonarClient,
	}
}

// FetchServices fetches services for an organization directly from API
func (s *ServiceService) FetchServices(orgID int64) ([]resources.ServiceResponse, error) {
	log.Printf("Fetching services for org_id=%d from sonar-shell-test API", orgID)

	// Fetch all organizations to get org info
	orgs, err := s.sonarClient.GetOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to fetch organizations: %w", err)
	}

	// Find the requested organization
	var org *models.Organization
	for i := range orgs {
		if orgs[i].ID == orgID {
			org = &orgs[i]
			break
		}
	}

	if org == nil {
		return nil, fmt.Errorf("organization with ID %d not found", orgID)
	}

	// Fetch repositories
	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// Convert to service response DTOs
	var responses []resources.ServiceResponse
	for i := range repos {
		response, err := s.convertToServiceResponse(&repos[i], org)
		if err != nil {
			log.Printf("Warning: failed to convert repository %s: %v", repos[i].Name, err)
			continue
		}
		responses = append(responses, response)
	}
	return responses, nil
}

// GetServiceByOrgAndID gets a specific service by organization ID and service ID
func (s *ServiceService) GetServiceByOrgAndID(orgID int64, serviceID string) (*resources.ServiceResponse, error) {
	// 1. Parse service ID to extract repository ID
	// Format: svc_12345 -> extract 12345
	var repoID int64
	_, err := fmt.Sscanf(serviceID, "svc_%d", &repoID)
	if err != nil {
		return nil, fmt.Errorf("invalid service ID format (expected svc_<number>): %w", err)
	}

	log.Printf("Looking up service %s (repository ID: %d) for org_id=%d", serviceID, repoID, orgID)

	// 2. Get organization info
	orgs, err := s.sonarClient.GetOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}

	// Find the requested organization
	var org *models.Organization
	for i := range orgs {
		if orgs[i].ID == orgID {
			org = &orgs[i]
			break
		}
	}

	if org == nil {
		return nil, fmt.Errorf("organization with ID %d not found", orgID)
	}

	// 3. Fetch all repositories for the organization
	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// 4. Find the requested repository by ID
	var foundRepo *models.Repository
	for i := range repos {
		if repos[i].ID == repoID {
			foundRepo = &repos[i]
			break
		}
	}

	// 5. Return the specific repository
	if foundRepo == nil {
		return nil, fmt.Errorf("service %s not found in organization %d", serviceID, orgID)
	}

	log.Printf("Successfully fetched service %s for org_id=%d", serviceID, orgID)

	// Convert to service response
	response, err := s.convertToServiceResponse(foundRepo, org)
	if err != nil {
		return nil, fmt.Errorf("failed to convert repository: %w", err)
	}
	return &response, nil
}

// GetService gets a specific service by ID (format: svc_12345) directly from API (legacy method)
func (s *ServiceService) GetService(serviceID string) (*resources.ServiceResponse, error) {
	// 1. Parse service ID to extract repository ID
	// Format: svc_12345 -> extract 12345
	var repoID int64
	_, err := fmt.Sscanf(serviceID, "svc_%d", &repoID)
	if err != nil {
		return nil, fmt.Errorf("invalid service ID format (expected svc_<number>): %w", err)
	}

	log.Printf("Looking up service %s (repository ID: %d)", serviceID, repoID)

	// 2. Get first available organization
	orgs, err := s.sonarClient.GetOrganizations()
	if err != nil {
		return nil, fmt.Errorf("failed to get organizations: %w", err)
	}

	if len(orgs) == 0 {
		return nil, fmt.Errorf("no organizations available to fetch repository")
	}

	// 3. Fetch all repositories using first org_id
	orgID := orgs[0].ID
	org := &orgs[0]
	log.Printf("Fetching repositories for org_id=%d to find service %s", orgID, serviceID)

	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// 4. Find the requested repository by ID
	var foundRepo *models.Repository
	for i := range repos {
		if repos[i].ID == repoID {
			foundRepo = &repos[i]
			break
		}
	}

	// 5. Return the specific repository
	if foundRepo == nil {
		return nil, fmt.Errorf("service %s not found", serviceID)
	}

	log.Printf("Successfully fetched service %s", serviceID)

	// Convert to service response
	response, err := s.convertToServiceResponse(foundRepo, org)
	if err != nil {
		return nil, fmt.Errorf("failed to convert repository: %w", err)
	}
	return &response, nil
}

// GetAllServices gets all services directly from API
func (s *ServiceService) GetAllServices() (*resources.ServicesResponse, error) {
	log.Printf("Fetching all services from sonar-shell-test API")

	// 1. Get first available organization
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

	// 2. Fetch repositories using first org_id
	orgID := orgs[0].ID
	org := &orgs[0]
	log.Printf("Fetching repositories for org_id=%d (%s)", orgID, orgs[0].Name)

	repos, err := s.sonarClient.FetchRepositoriesByOrg(orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch repositories from API: %w", err)
	}

	// 3. Convert to service responses
	var responses []resources.ServiceResponse
	for i := range repos {
		response, err := s.convertToServiceResponse(&repos[i], org)
		if err != nil {
			log.Printf("Warning: failed to convert repository %s: %v", repos[i].Name, err)
			continue
		}
		responses = append(responses, response)
	}
	log.Printf("Successfully fetched %d repositories", len(responses))

	return &resources.ServicesResponse{
		Total:    len(responses),
		Services: responses,
	}, nil
}

// generateServiceID generates a unique service ID from repository ID
func (s *ServiceService) generateServiceID(repoID int64) string {
	return fmt.Sprintf("svc_%d", repoID)
}

// convertToServiceResponse converts a repository model to ServiceResponse with real data from API
func (s *ServiceService) convertToServiceResponse(repo *models.Repository, org *models.Organization) (resources.ServiceResponse, error) {
	// Generate service ID
	serviceID := s.generateServiceID(repo.ID)

	// Fetch GitHub metrics
	var openPRsCount int
	var commitsLast90Days int
	var contributors int

	githubMetrics, err := s.sonarClient.GetGitHubMetrics(repo.ID)
	if err != nil {
		log.Printf("Warning: failed to fetch GitHub metrics for repo %s: %v", repo.Name, err)
	} else {
		openPRsCount = int(githubMetrics.OpenPRs)
		commitsLast90Days = int(githubMetrics.CommitsLast90Days)
		contributors = int(githubMetrics.Contributors)
	}

	// Fetch pull requests
	pullRequests := []resources.PullRequestInfo{}
	prs, err := s.sonarClient.GetPullRequests(repo.Name, "open")
	if err != nil {
		log.Printf("Warning: failed to fetch pull requests for repo %s: %v", repo.Name, err)
	} else {
		for _, pr := range prs {
			pullRequests = append(pullRequests, resources.PullRequestInfo{
				Number:    pr.Number,
				Title:     pr.Title,
				State:     pr.State,
				Author:    pr.User,
				CreatedAt: pr.CreatedAt,
				URL:       pr.URL,
			})
		}
		openPRsCount = len(pullRequests)
	}

	// Fetch Jira metrics from database
	var jiraOpenBugs int
	var jiraOpenTasks int
	var jiraActiveSprints int
	jiraIssues := []resources.JiraIssueInfo{}

	if repo.JiraProjectKey != "" {
		// Fetch Jira metrics from database (includes bugs, tasks, and sprints)
		jiraMetrics, err := s.sonarClient.GetJiraMetrics(repo.ID)
		if err != nil {
			log.Printf("Warning: failed to fetch Jira metrics for repo %s: %v", repo.Name, err)
		} else {
			jiraOpenBugs = int(jiraMetrics.OpenBugs)
			jiraOpenTasks = int(jiraMetrics.OpenTasks)
			jiraActiveSprints = int(jiraMetrics.ActiveSprints)
		}

		// Try to fetch detailed Jira issues (LIVE) - if credentials are configured
		bugs, err := s.sonarClient.GetJiraOpenBugs(repo.JiraProjectKey)
		if err != nil {
			log.Printf("Warning: failed to fetch Jira open bugs for project %s: %v", repo.JiraProjectKey, err)
		} else {
			// Add bugs to jiraIssues list
			for _, bug := range bugs {
				jiraIssues = append(jiraIssues, resources.JiraIssueInfo{
					Key:       bug.Key,
					Summary:   bug.Summary,
					IssueType: bug.IssueType,
					Status:    bug.Status,
					Priority:  bug.Priority,
					Assignee:  bug.Assignee,
				})
			}
		}

		// Try to fetch detailed Jira tasks (LIVE) - if credentials are configured
		tasks, err := s.sonarClient.GetJiraOpenTasks(repo.JiraProjectKey)
		if err != nil {
			log.Printf("Warning: failed to fetch Jira open tasks for project %s: %v", repo.JiraProjectKey, err)
		} else {
			// Add tasks to jiraIssues list
			for _, task := range tasks {
				jiraIssues = append(jiraIssues, resources.JiraIssueInfo{
					Key:       task.Key,
					Summary:   task.Summary,
					IssueType: task.IssueType,
					Status:    task.Status,
					Priority:  task.Priority,
					Assignee:  task.Assignee,
				})
			}
		}
	}

	// Build metrics
	metrics := resources.MetricsInfo{
		OpenPullRequests:  openPRsCount,
		CommitsLast90Days: commitsLast90Days,
		Contributors:      contributors,
		JiraOpenBugs:      jiraOpenBugs,
		JiraOpenTasks:     jiraOpenTasks,
		JiraActiveSprints: jiraActiveSprints,
	}

	// Determine language from GitHub metrics
	language := ""
	if githubMetrics != nil {
		// You can add language detection logic here if available in metrics
		language = "JavaScript" // Default for now
	}

	// Fetch evaluation metrics from multiple sonar-shell-test endpoints
	var evaluationMetrics *resources.EvaluationMetricsInfo
	evalMetrics, err := s.sonarClient.GetEvaluationMetrics(repo.ID, repo.Name)
	if err != nil {
		log.Printf("Warning: failed to fetch evaluation metrics for %s (repo_id=%d): %v", repo.Name, repo.ID, err)
		// Don't fail the whole request, just log and continue with nil
	} else {
		evaluationMetrics = &resources.EvaluationMetricsInfo{
			ServiceName:            evalMetrics.ServiceName,
			Coverage:               evalMetrics.Coverage,
			CodeSmells:             evalMetrics.CodeSmells,
			Vulnerabilities:        evalMetrics.Vulnerabilities,
			DuplicatedLinesDensity: evalMetrics.DuplicatedLinesDensity,
			HasReadme:              evalMetrics.HasReadme,
			DeploymentFrequency:    evalMetrics.DeploymentFrequency,
			MTTR:                   evalMetrics.MTTR,
		}
	}

	return resources.ServiceResponse{
		ID:                serviceID,
		Title:             repo.Name,
		RepositoryURL:     repo.GitHubURL,
		Owner:             repo.Owner,
		DefaultBranch:     repo.DefaultBranch,
		Language:          language,
		Organization:      resources.OrganizationInfo{ID: org.ID, Name: org.Name},
		JiraProjectKey:    repo.JiraProjectKey,
		OnCall:            repo.Owner, // Use repository owner as on-call for now
		Metrics:           metrics,
		EvaluationMetrics: evaluationMetrics,
		PullRequests:      pullRequests,
		JiraIssues:        jiraIssues,
	}, nil
}
