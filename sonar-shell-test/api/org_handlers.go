package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"sonar-automation/models"
)

// ═══════════════════════════════════════════════════════════════
// Organization Management Handlers
// ═══════════════════════════════════════════════════════════════

// fetchOrgsHandler handles GET /api/v1/orgs
func (s *Server) fetchOrgsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	orgs, err := s.db.GetAllOrganizations()
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch organizations: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d organizations", len(orgs)),
		Data:    orgs,
	})
}

// createOrgHandler handles POST /api/v1/orgs
func (s *Server) createOrgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	var org models.Organization
	if err := json.NewDecoder(r.Body).Decode(&org); err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Validate required fields
	if org.Name == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Organization name is required",
		})
		return
	}

	// Check if organization already exists
	existingOrg, err := s.db.GetOrganizationByName(org.Name)
	if err == nil && existingOrg != nil {
		sendJSON(w, http.StatusConflict, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization '%s' already exists", org.Name),
		})
		return
	}

	if err := s.db.CreateOrganization(&org); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to create organization: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusCreated, Response{
		Success: true,
		Message: fmt.Sprintf("Organization '%s' created successfully", org.Name),
		Data:    org,
	})
}

