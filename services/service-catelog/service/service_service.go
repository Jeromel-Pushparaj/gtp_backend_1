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

	// Use channels for parallel API calls
	type githubResult struct {
		metrics *models.GitHubMetrics
		err     error
	}
	type prResult struct {
		prs []models.PullRequest
		err error
	}
	type jiraMetricsResult struct {
		metrics *models.JiraMetrics
		err     error
	}
	type jiraBugsResult struct {
		bugs []models.JiraIssue
		err  error
	}
	type jiraTasksResult struct {
		tasks []models.JiraIssue
		err   error
	}
	type sonarResult struct {
		metrics *models.SonarMetrics
		err     error
	}

	githubChan := make(chan githubResult, 1)
	prChan := make(chan prResult, 1)
	jiraMetricsChan := make(chan jiraMetricsResult, 1)
	jiraBugsChan := make(chan jiraBugsResult, 1)
	jiraTasksChan := make(chan jiraTasksResult, 1)
	sonarChan := make(chan sonarResult, 1)

	// Fetch GitHub metrics in parallel
	go func() {
		metrics, err := s.sonarClient.GetGitHubMetrics(repo.ID)
		githubChan <- githubResult{metrics, err}
	}()

	// Fetch pull requests in parallel
	go func() {
		prs, err := s.sonarClient.GetPullRequests(repo.Name, "open")
		prChan <- prResult{prs, err}
	}()

	// Fetch Jira metrics in parallel
	go func() {
		metrics, err := s.sonarClient.GetJiraMetrics(repo.ID)
		jiraMetricsChan <- jiraMetricsResult{metrics, err}
	}()

	// Fetch Sonar metrics in parallel
	go func() {
		metrics, err := s.sonarClient.GetSonarMetrics(repo.ID)
		sonarChan <- sonarResult{metrics, err}
	}()

	// Fetch Jira bugs in parallel (only if Jira project key exists)
	if repo.JiraProjectKey != "" {
		go func() {
			bugs, err := s.sonarClient.GetJiraOpenBugs(repo.JiraProjectKey)
			jiraBugsChan <- jiraBugsResult{bugs, err}
		}()

		go func() {
			tasks, err := s.sonarClient.GetJiraOpenTasks(repo.JiraProjectKey)
			jiraTasksChan <- jiraTasksResult{tasks, err}
		}()
	} else {
		// Send empty results if no Jira project key
		jiraBugsChan <- jiraBugsResult{nil, nil}
		jiraTasksChan <- jiraTasksResult{nil, nil}
	}

	// Collect results from all goroutines
	var openPRsCount int
	var commitsLast90Days int
	var contributors int
	var githubMetrics *models.GitHubMetrics

	githubRes := <-githubChan
	if githubRes.err != nil {
		log.Printf("Warning: failed to fetch GitHub metrics for repo %s: %v", repo.Name, githubRes.err)
	} else {
		githubMetrics = githubRes.metrics
		openPRsCount = int(githubMetrics.OpenPRs)
		commitsLast90Days = int(githubMetrics.CommitsLast90Days)
		contributors = int(githubMetrics.Contributors)
	}

	pullRequests := []resources.PullRequestInfo{}
	prRes := <-prChan
	if prRes.err != nil {
		log.Printf("Warning: failed to fetch pull requests for repo %s: %v", repo.Name, prRes.err)
	} else {
		for _, pr := range prRes.prs {
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

	var jiraOpenBugs int
	var jiraOpenTasks int
	var jiraActiveSprints int
	var jiraMetrics *models.JiraMetrics
	jiraIssues := []resources.JiraIssueInfo{}

	jiraMetricsRes := <-jiraMetricsChan
	if jiraMetricsRes.err != nil {
		log.Printf("Warning: failed to fetch Jira metrics for repo %s: %v", repo.Name, jiraMetricsRes.err)
	} else {
		jiraMetrics = jiraMetricsRes.metrics
		jiraOpenBugs = int(jiraMetrics.OpenBugs)
		jiraOpenTasks = int(jiraMetrics.OpenTasks)
		jiraActiveSprints = int(jiraMetrics.ActiveSprints)
	}

	var sonarMetrics *models.SonarMetrics
	sonarRes := <-sonarChan
	if sonarRes.err != nil {
		log.Printf("Warning: failed to fetch Sonar metrics for repo %s: %v", repo.Name, sonarRes.err)
	} else {
		sonarMetrics = sonarRes.metrics
	}

	if repo.JiraProjectKey != "" {
		jiraBugsRes := <-jiraBugsChan
		if jiraBugsRes.err != nil {
			log.Printf("Warning: failed to fetch Jira open bugs for project %s: %v", repo.JiraProjectKey, jiraBugsRes.err)
		} else {
			for _, bug := range jiraBugsRes.bugs {
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

		jiraTasksRes := <-jiraTasksChan
		if jiraTasksRes.err != nil {
			log.Printf("Warning: failed to fetch Jira open tasks for project %s: %v", repo.JiraProjectKey, jiraTasksRes.err)
		} else {
			for _, task := range jiraTasksRes.tasks {
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
	} else {
		// Drain channels even if not used
		<-jiraBugsChan
		<-jiraTasksChan
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

	// Build evaluation metrics from already-fetched data (REUSE - no duplicate API calls!)
	var evaluationMetrics *resources.EvaluationMetricsInfo

	// We already have sonarMetrics, githubMetrics, and jiraMetrics from parallel calls above
	// No need to call GetEvaluationMetrics which would duplicate the API calls
	evaluationMetrics = &resources.EvaluationMetricsInfo{
		ServiceName:            repo.Name,
		Coverage:               0,
		CodeSmells:             0,
		Vulnerabilities:        0,
		DuplicatedLinesDensity: 0,
		HasReadme:              0,
		DeploymentFrequency:    0,
		MTTR:                   0,
	}

	// Populate from SonarCloud metrics
	if sonarMetrics != nil {
		evaluationMetrics.Coverage = sonarMetrics.Coverage
		evaluationMetrics.CodeSmells = int(sonarMetrics.CodeSmells)
		evaluationMetrics.Vulnerabilities = int(sonarMetrics.Vulnerabilities)
		evaluationMetrics.DuplicatedLinesDensity = sonarMetrics.DuplicatedLinesDensity
	}

	// Populate from GitHub metrics
	if githubMetrics != nil {
		if githubMetrics.HasReadme {
			evaluationMetrics.HasReadme = 1
		} else {
			evaluationMetrics.HasReadme = 0
		}
	}

	// Populate from Jira metrics
	if jiraMetrics != nil {
		// Convert avg_time_to_resolve (in hours) to days
		evaluationMetrics.MTTR = int(jiraMetrics.AvgTimeToResolve / 24)
	}

	// Deployment frequency not available
	evaluationMetrics.DeploymentFrequency = 0

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
