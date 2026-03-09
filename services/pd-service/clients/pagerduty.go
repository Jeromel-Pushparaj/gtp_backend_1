package clients

import (
	"context"
	"pd-service/models"
	"time"

	"github.com/PagerDuty/go-pagerduty"
)

type PagerDutyClient struct {
	client *pagerduty.Client
}

func NewPagerDutyClient(apiKey string) *PagerDutyClient {
	return &PagerDutyClient{
		client: pagerduty.NewClient(apiKey),
	}
}

func (p *PagerDutyClient) ListServices(ctx context.Context) ([]pagerduty.Service, error) {
	opts := pagerduty.ListServiceOptions{
		Limit: 100,
	}

	services, err := p.client.ListServicesWithContext(ctx, opts)
	if err != nil {
		return nil, err
	}

	return services.Services, nil
}

func (p *PagerDutyClient) ListEscalationPolicies(ctx context.Context) ([]pagerduty.EscalationPolicy, error) {
	opts := pagerduty.ListEscalationPoliciesOptions{
		Limit: 100,
	}

	policies, err := p.client.ListEscalationPoliciesWithContext(ctx, opts)
	if err != nil {
		return nil, err
	}

	return policies.EscalationPolicies, nil
}

func (p *PagerDutyClient) CreateService(ctx context.Context, serviceName, escalationPolicyID string) (*pagerduty.Service, error) {
	service := &pagerduty.Service{
		Name: serviceName,
		EscalationPolicy: pagerduty.EscalationPolicy{
			APIObject: pagerduty.APIObject{
				ID:   escalationPolicyID,
				Type: "escalation_policy_reference",
			},
		},
		IncidentUrgencyRule: &pagerduty.IncidentUrgencyRule{
			Type:    "constant",
			Urgency: "high",
		},
	}

	created, err := p.client.CreateServiceWithContext(ctx, *service)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (p *PagerDutyClient) GetService(ctx context.Context, serviceID string) (*pagerduty.Service, error) {
	opts := &pagerduty.GetServiceOptions{}
	service, err := p.client.GetServiceWithContext(ctx, serviceID, opts)
	if err != nil {
		return nil, err
	}
	return service, nil
}

func (p *PagerDutyClient) GetIncidents(ctx context.Context, serviceID string) ([]*models.Incident, error) {
	opts := pagerduty.ListIncidentsOptions{
		ServiceIDs: []string{serviceID},
		Limit:      100,
		SortBy:     "created_at:desc",
	}
	
	incidents, err := p.client.ListIncidentsWithContext(ctx, opts)
	if err != nil {
		return nil, err
	}
	
	var result []*models.Incident
	for _, inc := range incidents.Incidents {
		incident := &models.Incident{
			ID:          inc.ID,
			Title:       inc.Title,
			Status:      inc.Status,
			ServiceID:   inc.Service.ID,
			ServiceName: inc.Service.Summary,
			CreatedAt:   parseTime(inc.CreatedAt),
		}
		
		if inc.Priority != nil {
			incident.Priority = inc.Priority.Summary
		}
		
		if inc.Status == "resolved" && inc.LastStatusChangeAt != "" {
			t := parseTime(inc.LastStatusChangeAt)
			incident.ResolvedAt = &t
		}
		
		result = append(result, incident)
	}
	
	return result, nil
}

// ListUsers returns all PagerDuty users
func (p *PagerDutyClient) ListUsers(ctx context.Context) ([]pagerduty.User, error) {
	opts := pagerduty.ListUsersOptions{
		Limit: 100,
	}

	users, err := p.client.ListUsersWithContext(ctx, opts)
	if err != nil {
		return nil, err
	}

	return users.Users, nil
}

// GetUserByEmail finds a PagerDuty user by email
func (p *PagerDutyClient) GetUserByEmail(ctx context.Context, email string) (*pagerduty.User, error) {
	users, err := p.ListUsers(ctx)
	if err != nil {
		return nil, err
	}

	for _, user := range users {
		if user.Email == email {
			return &user, nil
		}
	}

	return nil, nil
}

// CreateUser creates a new PagerDuty user with the given email and name
func (p *PagerDutyClient) CreateUser(ctx context.Context, email, name string) (*pagerduty.User, error) {
	user := pagerduty.User{
		Name:  name,
		Email: email,
		Role:  "user", // Basic user role
	}

	created, err := p.client.CreateUserWithContext(ctx, user)
	if err != nil {
		return nil, err
	}

	return created, nil
}

// GetOrCreateUser gets an existing user by email or creates a new one
func (p *PagerDutyClient) GetOrCreateUser(ctx context.Context, email, name string) (*pagerduty.User, error) {
	// First try to find existing user
	existingUser, err := p.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		return existingUser, nil
	}

	// User doesn't exist, create new one
	return p.CreateUser(ctx, email, name)
}

// CreateEscalationPolicy creates a new escalation policy with a single user
func (p *PagerDutyClient) CreateEscalationPolicy(ctx context.Context, name string, userID string) (*pagerduty.EscalationPolicy, error) {
	policy := pagerduty.EscalationPolicy{
		Name: name,
		EscalationRules: []pagerduty.EscalationRule{
			{
				Delay: 30, // 30 minutes before escalating
				Targets: []pagerduty.APIObject{
					{
						ID:   userID,
						Type: "user_reference",
					},
				},
			},
		},
	}

	created, err := p.client.CreateEscalationPolicyWithContext(ctx, policy)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (p *PagerDutyClient) CreateIncident(ctx context.Context, serviceID, title, description, priority, fromEmail string) (*pagerduty.Incident, error) {
	incident := &pagerduty.CreateIncidentOptions{
		Type:  "incident",
		Title: title,
		Service: &pagerduty.APIReference{
			ID:   serviceID,
			Type: "service_reference",
		},
		Body: &pagerduty.APIDetails{
			Type:    "incident_body",
			Details: description,
		},
	}

	if priority != "" {
		incident.Priority = &pagerduty.APIReference{
			Type: "priority_reference",
			ID:   priority,
		}
	}

	result, err := p.client.CreateIncidentWithContext(ctx, fromEmail, incident)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func (p *PagerDutyClient) GetServiceMetrics(ctx context.Context, serviceID string) (*models.ServiceMetrics, error) {
	incidents, err := p.GetIncidents(ctx, serviceID)
	if err != nil {
		return nil, err
	}
	
	metrics := &models.ServiceMetrics{
		ServiceID: serviceID,
	}
	
	var totalResolveTime, totalResponseTime float64
	var resolvedCount int
	
	for _, inc := range incidents {
		metrics.TotalIncidents++
		
		if inc.Status == "triggered" || inc.Status == "acknowledged" {
			metrics.OpenIncidents++
		}
		
		if inc.Priority == "P1" || inc.Priority == "high" {
			metrics.HighPriority++
		}
		
		if inc.ResolvedAt != nil {
			resolvedCount++
			resolveTime := inc.ResolvedAt.Sub(inc.CreatedAt).Minutes()
			totalResolveTime += resolveTime
		}
		
		if metrics.LastIncidentTime == nil || inc.CreatedAt.After(*metrics.LastIncidentTime) {
			metrics.LastIncidentTime = &inc.CreatedAt
		}
	}
	
	if resolvedCount > 0 {
		metrics.AvgTimeToResolve = totalResolveTime / float64(resolvedCount)
		metrics.AvgTimeToRespond = totalResponseTime / float64(resolvedCount)
	}
	
	return metrics, nil
}

func parseTime(timeStr string) time.Time {
	t, _ := time.Parse(time.RFC3339, timeStr)
	return t
}

