package resources

import "time"

// APIResponse represents a standard API response wrapper
type APIResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ServiceResponse represents a comprehensive service object for frontend
type ServiceResponse struct {
	ID                   string            `json:"id"`
	Title                string            `json:"title"`
	RepositorySystem     string            `json:"repositorySystem"`
	RepositoryURL        string            `json:"repositoryUrl"`
	Language             string            `json:"language"`
	Disposition          string            `json:"disposition"`
	Region               string            `json:"region"`
	CloudMigrationStatus string            `json:"cloudMigrationStatus"`
	Product              ProductInfo       `json:"product"`
	Module               ModuleInfo        `json:"module"`
	Ownership            OwnershipInfo     `json:"ownership"`
	Metrics              MetricsInfo       `json:"metrics"`
	OpenPullRequests     []PullRequestInfo `json:"openPullRequests"`
	JenkinsJobs          []JenkinsJobInfo  `json:"jenkinsJobs"`
	JiraIssues           []JiraIssueInfo   `json:"jiraIssues"`
	WizIssues            []WizIssueInfo    `json:"wizIssues"`
	PdIncidents          []PdIncidentInfo  `json:"pdIncidents"`
}

// ProductInfo represents product information
type ProductInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// ModuleInfo represents module information
type ModuleInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// OwnershipInfo represents ownership hierarchy
type OwnershipInfo struct {
	Manager  PersonInfo `json:"manager"`
	Director PersonInfo `json:"director"`
	VP       PersonInfo `json:"vp"`
}

// PersonInfo represents a person in the organization
type PersonInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// MetricsInfo represents aggregated metrics
type MetricsInfo struct {
	JiraIssuesCount         int `json:"jiraIssuesCount"`
	PullRequestsCount       int `json:"pullRequestsCount"`
	MergeRequestsCount      int `json:"mergeRequestsCount"`
	RcaReportsCount         int `json:"rcaReportsCount"`
	JenkinsJobsCount        int `json:"jenkinsJobsCount"`
	PassingJenkinsJobsCount int `json:"passingJenkinsJobsCount"`
	WizIssuesCount          int `json:"wizIssuesCount"`
	PdIncidentsCount        int `json:"pdIncidentsCount"`
}

// PullRequestInfo represents a pull request
type PullRequestInfo struct {
	ID         string     `json:"id"`
	ExternalID string     `json:"externalId"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	URL        string     `json:"url"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

// JenkinsJobInfo represents a Jenkins job
type JenkinsJobInfo struct {
	ID                 string     `json:"id"`
	Title              string     `json:"title"`
	Status             string     `json:"status"`
	LastUpdate         *time.Time `json:"lastUpdate,omitempty"`
	EntityCreationDate *time.Time `json:"entityCreationDate,omitempty"`
	URL                string     `json:"url"`
}

// JiraIssueInfo represents a Jira issue
type JiraIssueInfo struct {
	ID         string     `json:"id"`
	Key        string     `json:"key"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	Priority   string     `json:"priority"`
	LastUpdate *time.Time `json:"lastUpdate,omitempty"`
	URL        string     `json:"url"`
}

// WizIssueInfo represents a Wiz security issue
type WizIssueInfo struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Severity   string     `json:"severity"`
	Status     string     `json:"status"`
	DetectedAt *time.Time `json:"detectedAt,omitempty"`
	URL        string     `json:"url"`
}

// PdIncidentInfo represents a PagerDuty incident
type PdIncidentInfo struct {
	ID             string     `json:"id"`
	IncidentNumber string     `json:"incidentNumber"`
	Title          string     `json:"title"`
	Status         string     `json:"status"`
	Severity       string     `json:"severity"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
	ResolvedAt     *time.Time `json:"resolvedAt,omitempty"`
	URL            string     `json:"url"`
}

// ServicesResponse represents a list of services for frontend
type ServicesResponse struct {
	Total    int               `json:"total"`
	Services []ServiceResponse `json:"services"`
}
