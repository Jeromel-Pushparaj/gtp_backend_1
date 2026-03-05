package handlers

import (
	"log"
	"net/http"
	"pd-service/clients"
	"pd-service/models"
	"pd-service/storage"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	storage      *storage.InMemoryStorage
	mongoStorage *storage.MongoDBStorage
	pd           *clients.PagerDutyClient
	slack        *clients.SlackClient
	github       *clients.GitHubClient
	orgName      string
}

func NewHandler(
	storage *storage.InMemoryStorage,
	mongoStorage *storage.MongoDBStorage,
	pd *clients.PagerDutyClient,
	slack *clients.SlackClient,
	github *clients.GitHubClient,
	orgName string,
) *Handler {
	return &Handler{
		storage:      storage,
		mongoStorage: mongoStorage,
		pd:           pd,
		slack:        slack,
		github:       github,
		orgName:      orgName,
	}
}

// GetOrganizations returns available organizations
func (h *Handler) GetOrganizations(c *gin.Context) {
	orgs := []models.Organization{
		{ID: "1", Name: h.orgName},
	}
	c.JSON(http.StatusOK, orgs)
}

// ListServices returns all services from PagerDuty
func (h *Handler) ListServices(c *gin.Context) {
	log.Println("📋 Fetching services from PagerDuty...")

	pdServices, err := h.pd.ListServices(c.Request.Context())
	if err != nil {
		log.Printf("❌ Failed to fetch PagerDuty services: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PagerDuty services: " + err.Error()})
		return
	}

	log.Printf("✅ Found %d services in PagerDuty", len(pdServices))

	// Convert PagerDuty services to our model format
	var services []*models.Service
	for _, pdSvc := range pdServices {
		service := &models.Service{
			ID:          pdSvc.ID,
			Name:        pdSvc.Name,
			PDServiceID: pdSvc.ID,
		}

		// Try to get additional info from local storage if it exists
		localSvc, _ := h.storage.GetServiceByPDID(pdSvc.ID)
		if localSvc != nil {
			service.ID = localSvc.ID // Use the local ID if it exists
			service.GitHubRepo = localSvc.GitHubRepo
			service.SlackAssignee = localSvc.SlackAssignee
			service.SlackAssigneeID = localSvc.SlackAssigneeID
			service.OrgName = localSvc.OrgName
			service.CreatedAt = localSvc.CreatedAt
			service.UpdatedAt = localSvc.UpdatedAt
		}

		services = append(services, service)

		// Upsert service to MongoDB
		if h.mongoStorage != nil {
			if err := h.mongoStorage.UpsertService(service); err != nil {
				log.Printf("⚠️ Failed to upsert service to MongoDB (continuing): %v", err)
			}
		}
	}

	if services == nil {
		services = []*models.Service{}
	}

	c.JSON(http.StatusOK, services)
}

// GetService returns a specific service from PagerDuty
func (h *Handler) GetService(c *gin.Context) {
	id := c.Param("id")

	// Fetch from PagerDuty
	pdService, err := h.pd.GetService(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch service from PagerDuty: " + err.Error()})
		return
	}

	if pdService == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	// Convert to our model
	service := &models.Service{
		ID:          pdService.ID,
		Name:        pdService.Name,
		PDServiceID: pdService.ID,
	}

	// Try to get additional info from local storage
	localSvc, _ := h.storage.GetServiceByPDID(id)
	if localSvc != nil {
		service.GitHubRepo = localSvc.GitHubRepo
		service.SlackAssignee = localSvc.SlackAssignee
		service.SlackAssigneeID = localSvc.SlackAssigneeID
		service.OrgName = localSvc.OrgName
		service.CreatedAt = localSvc.CreatedAt
		service.UpdatedAt = localSvc.UpdatedAt
	}

	c.JSON(http.StatusOK, service)
}

// CreateService creates a new service
func (h *Handler) CreateService(c *gin.Context) {
	log.Println("📥 CreateService handler called")

	var req struct {
		Name            string `json:"name"`
		GithubRepo      string `json:"github_repo"`
		SlackAssignee   string `json:"slack_assignee"`
		SlackAssigneeID string `json:"slack_assignee_id"`
		OrgName         string `json:"org_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		log.Printf("❌ Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body: " + err.Error()})
		return
	}

	log.Printf("📋 Request data: Name=%s, GithubRepo=%s, SlackAssignee=%s", req.Name, req.GithubRepo, req.SlackAssignee)

	// Get Slack user's email
	log.Printf("🔍 Finding Slack user email for: %s", req.SlackAssigneeID)
	slackUsers, err := h.slack.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slack users: " + err.Error()})
		return
	}

	var slackUserEmail string
	for _, user := range slackUsers {
		if user.ID == req.SlackAssigneeID {
			slackUserEmail = user.Email
			log.Printf("✅ Found Slack user email: %s", slackUserEmail)
			break
		}
	}

	if slackUserEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find email for Slack assignee"})
		return
	}

	// Get or create PagerDuty user
	log.Printf("👤 Getting or creating PagerDuty user for: %s (%s)", req.SlackAssignee, slackUserEmail)
	pdUser, err := h.pd.GetOrCreateUser(c.Request.Context(), slackUserEmail, req.SlackAssignee)
	if err != nil {
		log.Printf("❌ Failed to get/create PagerDuty user: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PagerDuty user: " + err.Error()})
		return
	}
	log.Printf("✅ PagerDuty user ready: %s (ID: %s)", pdUser.Name, pdUser.ID)

	// Create escalation policy for this user
	escalationPolicyName := req.Name + " - Escalation Policy"
	log.Printf("📋 Creating escalation policy: %s for user: %s", escalationPolicyName, pdUser.ID)
	escalationPolicy, err := h.pd.CreateEscalationPolicy(c.Request.Context(), escalationPolicyName, pdUser.ID)
	if err != nil {
		log.Printf("❌ Failed to create escalation policy: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create escalation policy: " + err.Error()})
		return
	}
	log.Printf("✅ Escalation policy created with ID: %s", escalationPolicy.ID)

	// Create PagerDuty service
	log.Printf("🔄 Creating PagerDuty service: %s with escalation policy: %s", req.Name, escalationPolicy.ID)
	pdService, err := h.pd.CreateService(c.Request.Context(), req.Name, escalationPolicy.ID)
	if err != nil {
		log.Printf("❌ Failed to create PagerDuty service: %v", err)
		// Return a more user-friendly error message
		errorMsg := err.Error()
		if contains(errorMsg, "already been taken") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "A service with this name already exists in PagerDuty. Please use a different name."})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create PagerDuty service: " + errorMsg})
		}
		return
	}
	log.Printf("✅ PagerDuty service created with ID: %s", pdService.ID)

	// Create local service record
	log.Println("💾 Creating local service record...")
	service := models.Service{
		ID:              uuid.New().String(),
		Name:            req.Name,
		PDServiceID:     pdService.ID,
		GitHubRepo:      req.GithubRepo,
		SlackAssignee:   req.SlackAssignee,
		SlackAssigneeID: req.SlackAssigneeID,
		OrgName:         req.OrgName,
	}

	if service.OrgName == "" {
		service.OrgName = h.orgName
	}

	log.Printf("💾 Saving service to storage: %+v", service)

	if err := h.storage.CreateService(&service); err != nil {
		log.Printf("❌ Failed to save service to storage: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Store service in MongoDB
	if h.mongoStorage != nil {
		if err := h.mongoStorage.StoreService(&service); err != nil {
			log.Printf("⚠️ Failed to store service in MongoDB (continuing): %v", err)
			// Don't fail the request if MongoDB storage fails
		}
	}

	log.Printf("✅ Service saved successfully, returning response...")
	c.JSON(http.StatusCreated, service)
	log.Println("✅ Response sent to client")
}

// DeleteService deletes a service
func (h *Handler) DeleteService(c *gin.Context) {
	id := c.Param("id")

	if err := h.storage.DeleteService(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Service deleted"})
}

// ListPDServices returns all PagerDuty services
func (h *Handler) ListPDServices(c *gin.Context) {
	services, err := h.pd.ListServices(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PagerDuty API error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, services)
}

// ListEscalationPolicies returns all PagerDuty escalation policies
func (h *Handler) ListEscalationPolicies(c *gin.Context) {
	policies, err := h.pd.ListEscalationPolicies(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "PagerDuty API error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, policies)
}

// ListGitHubRepos returns all GitHub repositories
func (h *Handler) ListGitHubRepos(c *gin.Context) {
	org := c.Query("org")
	if org == "" {
		org = h.orgName
	}

	repos, err := h.github.ListRepositories(c.Request.Context(), org)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "GitHub API error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, repos)
}

// ListSlackUsers returns all Slack users
func (h *Handler) ListSlackUsers(c *gin.Context) {
	users, err := h.slack.ListUsers(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Slack API error: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, users)
}

// GetServiceMetrics returns metrics for a specific service
func (h *Handler) GetServiceMetrics(c *gin.Context) {
	id := c.Param("id")

	service, err := h.storage.GetService(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if service == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
		return
	}

	metrics, err := h.pd.GetServiceMetrics(c.Request.Context(), service.PDServiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	metrics.ServiceName = service.Name
	metrics.AssigneeName = service.SlackAssignee
	metrics.AssigneeSlackID = service.SlackAssigneeID

	c.JSON(http.StatusOK, metrics)
}

// GetAllMetrics returns metrics for all services from PagerDuty
func (h *Handler) GetAllMetrics(c *gin.Context) {
	log.Println("📊 Fetching metrics for all PagerDuty services...")

	// Fetch all services from PagerDuty
	pdServices, err := h.pd.ListServices(c.Request.Context())
	if err != nil {
		log.Printf("❌ Failed to fetch PagerDuty services: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch PagerDuty services: " + err.Error()})
		return
	}

	log.Printf("📊 Calculating metrics for %d services...", len(pdServices))

	var allMetrics []*models.ServiceMetrics
	for _, pdSvc := range pdServices {
		metrics, err := h.pd.GetServiceMetrics(c.Request.Context(), pdSvc.ID)
		if err != nil {
			log.Printf("⚠️ Failed to get metrics for service %s: %v", pdSvc.Name, err)
			continue // Skip services with errors
		}

		metrics.ServiceName = pdSvc.Name

		// Try to get assignee info from local storage if it exists
		localSvc, _ := h.storage.GetServiceByPDID(pdSvc.ID)
		if localSvc != nil {
			metrics.AssigneeName = localSvc.SlackAssignee
			metrics.AssigneeSlackID = localSvc.SlackAssigneeID
		}

		allMetrics = append(allMetrics, metrics)
	}

	if allMetrics == nil {
		allMetrics = []*models.ServiceMetrics{}
	}

	// Store metrics in MongoDB
	if h.mongoStorage != nil && len(allMetrics) > 0 {
		if err := h.mongoStorage.StoreMetrics(allMetrics); err != nil {
			log.Printf("⚠️ Failed to store metrics in MongoDB (continuing): %v", err)
			// Don't fail the request if MongoDB storage fails
		}
	}

	log.Printf("✅ Returning metrics for %d services", len(allMetrics))
	c.JSON(http.StatusOK, allMetrics)
}

// GetMetricsByServiceName returns metrics for a service by name
func (h *Handler) GetMetricsByServiceName(c *gin.Context) {
	var req struct {
		ServiceName string `json:"service_name" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "service_name is required"})
		return
	}

	log.Printf("📊 Fetching metrics for service: %s", req.ServiceName)

	// First, try to find the service in MongoDB
	var service *models.Service
	var err error

	if h.mongoStorage != nil {
		service, err = h.mongoStorage.GetServiceByName(req.ServiceName)
		if err != nil {
			log.Printf("⚠️ Error querying MongoDB: %v", err)
		}
	}

	// If not found in MongoDB, try local storage
	if service == nil {
		log.Printf("🔍 Service not found in MongoDB, checking local storage...")
		services, _ := h.storage.ListServices("")
		for _, svc := range services {
			if svc.Name == req.ServiceName {
				service = svc
				break
			}
		}
	}

	// If still not found, search in PagerDuty
	if service == nil {
		log.Printf("🔍 Service not found locally, searching PagerDuty...")
		pdServices, err := h.pd.ListServices(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch services from PagerDuty: " + err.Error()})
			return
		}

		for _, pdSvc := range pdServices {
			if pdSvc.Name == req.ServiceName {
				service = &models.Service{
					ID:          pdSvc.ID,
					Name:        pdSvc.Name,
					PDServiceID: pdSvc.ID,
				}
				break
			}
		}
	}

	if service == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found with name: " + req.ServiceName})
		return
	}

	log.Printf("✅ Found service: %s (PD ID: %s)", service.Name, service.PDServiceID)

	// Fetch metrics from PagerDuty
	metrics, err := h.pd.GetServiceMetrics(c.Request.Context(), service.PDServiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch metrics: " + err.Error()})
		return
	}

	metrics.ServiceName = service.Name
	metrics.AssigneeName = service.SlackAssignee
	metrics.AssigneeSlackID = service.SlackAssigneeID

	// Store metrics in MongoDB
	if h.mongoStorage != nil {
		metricsArray := []*models.ServiceMetrics{metrics}
		if err := h.mongoStorage.StoreMetrics(metricsArray); err != nil {
			log.Printf("⚠️ Failed to store metrics in MongoDB (continuing): %v", err)
		}
	}

	log.Printf("✅ Returning metrics for service: %s", service.Name)
	c.JSON(http.StatusOK, metrics)
}

// TriggerIncident creates a test incident
func (h *Handler) TriggerIncident(c *gin.Context) {
	var req models.TriggerIncidentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	log.Printf("🚨 Triggering incident for service: %s", req.ServiceID)

	// Fetch service from PagerDuty (req.ServiceID is the PagerDuty service ID)
	pdService, err := h.pd.GetService(c.Request.Context(), req.ServiceID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch service from PagerDuty: " + err.Error()})
		return
	}

	if pdService == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Service not found in PagerDuty"})
		return
	}

	// Try to get assignee info from local storage
	localSvc, _ := h.storage.GetServiceByPDID(req.ServiceID)

	var slackUserEmail string
	var slackAssigneeID string
	var serviceName = pdService.Name

	if localSvc != nil && localSvc.SlackAssigneeID != "" {
		// We have local metadata with Slack assignee
		log.Printf("🔍 Finding PagerDuty user for Slack assignee: %s", localSvc.SlackAssigneeID)
		slackAssigneeID = localSvc.SlackAssigneeID

		slackUsers, err := h.slack.ListUsers(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get Slack users: " + err.Error()})
			return
		}

		for _, user := range slackUsers {
			if user.ID == localSvc.SlackAssigneeID {
				slackUserEmail = user.Email
				log.Printf("✅ Found Slack user email: %s", slackUserEmail)
				break
			}
		}

		if slackUserEmail == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Could not find email for Slack assignee"})
			return
		}
	} else {
		// No local metadata - get the first user from the escalation policy
		log.Printf("⚠️ No local metadata found for service %s, using escalation policy to find user", req.ServiceID)

		// Get the escalation policy from the service
		if pdService.EscalationPolicy.ID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Service has no escalation policy configured"})
			return
		}

		// For now, we'll use a default email - in production you'd fetch the escalation policy details
		// and get the first user's email from it
		c.JSON(http.StatusBadRequest, gin.H{"error": "Service was not created through this app. Please add it first to associate a Slack user."})
		return
	}

	// Find matching PagerDuty user by email
	log.Printf("🔍 Finding PagerDuty user with email: %s", slackUserEmail)
	pdUser, err := h.pd.GetUserByEmail(c.Request.Context(), slackUserEmail)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find PagerDuty user: " + err.Error()})
		return
	}

	if pdUser == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No PagerDuty user found with email: " + slackUserEmail + ". Please ensure the Slack user has a matching PagerDuty account."})
		return
	}

	log.Printf("✅ Found PagerDuty user: %s (%s)", pdUser.Name, pdUser.Email)

	// Create incident in PagerDuty using the PD user's email
	log.Printf("🚨 Creating incident from user: %s", pdUser.Email)
	incident, err := h.pd.CreateIncident(c.Request.Context(), req.ServiceID, req.Title, req.Description, req.Priority, pdUser.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create incident: " + err.Error()})
		return
	}

	log.Printf("✅ Incident created: %s", incident.ID)

	// Send Slack notification if we have an assignee
	if slackAssigneeID != "" {
		incidentURL := incident.HTMLURL
		if err := h.slack.SendIncidentNotification(c.Request.Context(), slackAssigneeID, serviceName, req.Title, incidentURL); err != nil {
			// Log error but don't fail the request
			log.Printf("⚠️ Failed to send Slack notification: %v", err)
			c.Writer.Header().Add("X-Slack-Error", err.Error())
		}
	}

	c.JSON(http.StatusCreated, gin.H{
		"incident_id":  incident.ID,
		"incident_url": incident.HTMLURL,
		"message":      "Incident created and notification sent",
	})
}

// Helper function to check if a string contains a substring
func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
