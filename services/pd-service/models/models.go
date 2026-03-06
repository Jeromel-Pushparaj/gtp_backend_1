package models

import "time"


type Service struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	PDServiceID     string    `json:"pd_service_id"`
	GitHubRepo      string    `json:"github_repo"`
	SlackAssignee   string    `json:"slack_assignee"`
	SlackAssigneeID string    `json:"slack_assignee_id"`
	OrgName         string    `json:"org_name"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}


type ServiceMetrics struct {
	ServiceID         string  `json:"service_id"`
	ServiceName       string  `json:"service_name"`
	OpenIncidents     int     `json:"open_incidents"`
	TotalIncidents    int     `json:"total_incidents"`
	HighPriority      int     `json:"high_priority"`
	AvgTimeToResolve  float64 `json:"avg_time_to_resolve"`  // in minutes
	AvgTimeToRespond  float64 `json:"avg_time_to_respond"`  // in minutes
	AssigneeName      string  `json:"assignee_name"`
	AssigneeSlackID   string  `json:"assignee_slack_id"`
	LastIncidentTime  *time.Time `json:"last_incident_time,omitempty"`
}

type Incident struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Status      string    `json:"status"`
	Priority    string    `json:"priority"`
	ServiceID   string    `json:"service_id"`
	ServiceName string    `json:"service_name"`
	CreatedAt   time.Time `json:"created_at"`
	ResolvedAt  *time.Time `json:"resolved_at,omitempty"`
}


type GitHubRepo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	URL      string `json:"url"`
}


type SlackUser struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	RealName string `json:"real_name"`
	Email    string `json:"email"`
}

type TriggerIncidentRequest struct {
	ServiceID   string `json:"service_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Priority    string `json:"priority"`
}


type Organization struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}