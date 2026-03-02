package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/keerthanau/go/config"
	"github.com/keerthanau/go/models"
	"github.com/keerthanau/go/services"
)

// HealthHandler handles health check requests
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// CreateIssueHandler handles POST requests to create issues
func CreateIssueHandler(cfg *config.Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Content-Type", "application/json")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Only accept POST requests
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(models.CreateIssueResponse{
				Success: false,
				Error:   "Method not allowed. Use POST",
			})
			return
		}

		// Parse request body
		var req models.CreateIssueRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.CreateIssueResponse{
				Success: false,
				Error:   "Invalid JSON body: " + err.Error(),
			})
			return
		}

		// Validate required fields
		if req.Summary == "" {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(models.CreateIssueResponse{
				Success: false,
				Error:   "summary is required",
			})
			return
		}

		// Set defaults for optional fields
		if req.IssueType == "" {
			req.IssueType = "Task"
		}
		if req.Description == "" {
			req.Description = "Created via API"
		}

		// Use projectKey from request, or fall back to config default
		projectKey := req.ProjectKey
		if projectKey == "" {
			projectKey = cfg.ProjectKey
		}

		// Create Jira client
		jiraClient := &models.JiraAPIClient{
			BaseURL: cfg.BaseURL,
			Email:   cfg.Email,
			Token:   cfg.APIToken,
			HTTP:    &http.Client{},
		}

		// Process the request
		response := services.ProcessCreateIssue(jiraClient, projectKey, &req)

		// Send response
		if response.Success {
			w.WriteHeader(http.StatusCreated)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		json.NewEncoder(w).Encode(response)
	}
}
