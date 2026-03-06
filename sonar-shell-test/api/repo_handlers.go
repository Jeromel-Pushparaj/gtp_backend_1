package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"sonar-automation/models"
	"sonar-automation/services"
)

// ═══════════════════════════════════════════════════════════════
// Repository Management Handlers
// ═══════════════════════════════════════════════════════════════

// fetchReposByOrgHandler handles GET /api/v1/orgs/{org_id}/repos
func (s *Server) fetchReposByOrgHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	orgIDStr := r.URL.Query().Get("org_id")
	if orgIDStr == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "org_id parameter is required",
		})
		return
	}

	orgID, err := strconv.ParseInt(orgIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid org_id parameter",
		})
		return
	}

	// Get organization
	org, err := s.db.GetOrganizationByID(orgID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Organization not found: %v", err),
		})
		return
	}

	// Fetch repositories from GitHub
	gs := services.NewGitHubService(org.GitHubPAT)
	ghRepos, err := gs.ListRepositories(org.Name)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to fetch repositories from GitHub: %v", err),
		})
		return
	}

	// Update database with repository details
	var repos []models.Repository
	for _, ghRepo := range ghRepos {
		// Get or create repository
		dbRepo, err := s.db.GetRepositoryByName(orgID, ghRepo.GetName())
		if err != nil {

			// Get repository owner (user with admin permissions)
			repoOwner, err := gs.GetRepositoryOwner(org.Name, ghRepo.GetName())
			if err != nil {
				repoOwner = org.Name // Fallback to org name
			}

			// Create new repository
			dbRepo = &models.Repository{
				OrgID:         orgID,
				Name:          ghRepo.GetName(),
				GitHubURL:     ghRepo.GetHTMLURL(),
				Owner:         repoOwner,
				DefaultBranch: ghRepo.GetDefaultBranch(),
				IsActive:      !ghRepo.GetArchived(),
			}

			// Get last commit info (get latest commits without time filter)
			commits, err := gs.ListCommits(org.Name, ghRepo.GetName(), nil)
			if err == nil && len(commits) > 0 {
				lastCommit := commits[0]
				if lastCommit.GetCommit() != nil && lastCommit.GetCommit().GetCommitter() != nil {
					commitTime := lastCommit.GetCommit().GetCommitter().GetDate().Time
					dbRepo.LastCommitTime = &commitTime
					dbRepo.LastCommitBy = lastCommit.GetCommit().GetCommitter().GetName()
				}
			}

			languages, err := gs.GetRepositoryLanguages(org.Name, ghRepo.GetName())
			if err == nil && len(languages) > 0 {
				primaryLang := gs.GetPrimaryLanguage(languages)
				dbRepo.PrimaryLanguage = &primaryLang
			}

			if err := s.db.CreateRepository(dbRepo); err != nil {
				log.Printf("ERROR creating repo '%s': %v", dbRepo.Name, err)
				continue
			}
		} else {
			// Get repository owner (user with admin permissions)
			repoOwner, err := gs.GetRepositoryOwner(org.Name, ghRepo.GetName())
			if err != nil {
				repoOwner = org.Name // Fallback to org name
			}

			// Update existing repository
			dbRepo.GitHubURL = ghRepo.GetHTMLURL()
			dbRepo.Owner = repoOwner
			dbRepo.DefaultBranch = ghRepo.GetDefaultBranch()
			dbRepo.IsActive = !ghRepo.GetArchived()

			// Get last commit info (get latest commits without time filter)
			commits, err := gs.ListCommits(org.Name, ghRepo.GetName(), nil)
			if err == nil && len(commits) > 0 {
				lastCommit := commits[0]
				if lastCommit.GetCommit() != nil && lastCommit.GetCommit().GetCommitter() != nil {
					commitTime := lastCommit.GetCommit().GetCommitter().GetDate().Time
					dbRepo.LastCommitTime = &commitTime
					dbRepo.LastCommitBy = lastCommit.GetCommit().GetCommitter().GetName()
				}
			}

			languages, err := gs.GetRepositoryLanguages(org.Name, ghRepo.GetName())
			if err == nil && len(languages) > 0 {
				primaryLang := gs.GetPrimaryLanguage(languages)
				dbRepo.PrimaryLanguage = &primaryLang
			}

			if err := s.db.UpdateRepository(dbRepo); err != nil {
				// Log error but continue
				log.Printf("ERROR updating repo '%s': %v", dbRepo.Name, err)
				continue
			}
		}

		repos = append(repos, *dbRepo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d repositories", len(repos)),
		Data:    repos,
	})
}

// updateRepoHandler handles PUT /api/v1/repos/{repo_id}
func (s *Server) updateRepoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repoIDStr := r.URL.Query().Get("repo_id")
	if repoIDStr == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo_id parameter is required",
		})
		return
	}

	repoID, err := strconv.ParseInt(repoIDStr, 10, 64)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid repo_id parameter",
		})
		return
	}

	var updateData struct {
		JiraProjectKey  string `json:"jira_project_key"`
		EnvironmentName string `json:"environment_name"`
	}

	if err := json.NewDecoder(r.Body).Decode(&updateData); err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   fmt.Sprintf("Invalid request body: %v", err),
		})
		return
	}

	// Get repository
	repo, err := s.db.GetRepositoryByID(repoID)
	if err != nil {
		sendJSON(w, http.StatusNotFound, Response{
			Success: false,
			Error:   fmt.Sprintf("Repository not found: %v", err),
		})
		return
	}

	// Update fields
	if updateData.JiraProjectKey != "" {
		repo.JiraProjectKey = updateData.JiraProjectKey
		repo.JiraSpaceKey = updateData.JiraProjectKey // Keep both in sync
	}
	if updateData.EnvironmentName != "" {
		repo.EnvironmentName = updateData.EnvironmentName
		repo.EnvName = updateData.EnvironmentName // Keep both in sync
	}
	repo.UpdatedAt = time.Now()

	if err := s.db.UpdateRepository(repo); err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to update repository: %v", err),
		})
		return
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: "Repository updated successfully",
		Data:    repo,
	})
}
