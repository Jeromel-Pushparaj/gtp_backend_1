package api

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/google/go-github/v58/github"
	"sonar-automation/models"
	"sonar-automation/services"
)

// GitHubMetricsRequest represents a request for GitHub metrics
type GitHubMetricsRequest struct {
	RepositoryName string `json:"repository_name"`
}

// ═══════════════════════════════════════════════════════════════
// Pull Requests Handlers
// ═══════════════════════════════════════════════════════════════

// listPullRequestsHandler handles GET /api/v1/github/pulls
func (s *Server) listPullRequestsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	state := r.URL.Query().Get("state")
	if state == "" {
		state = "all"
	}

	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	prs, err := gs.ListPullRequests(s.config.Organization, repo, state)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list pull requests: %v", err),
		})
		return
	}

	var prInfos []models.PullRequestInfo
	for _, pr := range prs {
		prInfo := models.PullRequestInfo{
			Number:         pr.GetNumber(),
			Title:          pr.GetTitle(),
			State:          pr.GetState(),
			CreatedAt:      pr.GetCreatedAt().Time,
			UpdatedAt:      pr.GetUpdatedAt().Time,
			User:           pr.GetUser().GetLogin(),
			Additions:      pr.GetAdditions(),
			Deletions:      pr.GetDeletions(),
			ChangedFiles:   pr.GetChangedFiles(),
			Commits:        pr.GetCommits(),
			Comments:       pr.GetComments(),
			ReviewComments: pr.GetReviewComments(),
			Head:           pr.GetHead().GetRef(),
			Base:           pr.GetBase().GetRef(),
		}

		if pr.ClosedAt != nil {
			prInfo.ClosedAt = &pr.ClosedAt.Time
		}
		if pr.MergedAt != nil {
			prInfo.MergedAt = &pr.MergedAt.Time
		}
		if pr.Mergeable != nil {
			prInfo.Mergeable = pr.Mergeable
			prInfo.HasConflicts = !*pr.Mergeable
		}

		prInfos = append(prInfos, prInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d pull requests", len(prInfos)),
		Data:    prInfos,
	})
}

// getPullRequestHandler handles GET /api/v1/github/pulls/{number}
func (s *Server) getPullRequestHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	numberStr := r.URL.Query().Get("number")

	if repo == "" || numberStr == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo and number parameters are required",
		})
		return
	}

	number, err := strconv.Atoi(numberStr)
	if err != nil {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "Invalid pull request number",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	pr, err := gs.GetPullRequest(s.config.Organization, repo, number)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to get pull request: %v", err),
		})
		return
	}

	prInfo := models.PullRequestInfo{
		Number:         pr.GetNumber(),
		Title:          pr.GetTitle(),
		State:          pr.GetState(),
		CreatedAt:      pr.GetCreatedAt().Time,
		UpdatedAt:      pr.GetUpdatedAt().Time,
		User:           pr.GetUser().GetLogin(),
		Additions:      pr.GetAdditions(),
		Deletions:      pr.GetDeletions(),
		ChangedFiles:   pr.GetChangedFiles(),
		Commits:        pr.GetCommits(),
		Comments:       pr.GetComments(),
		ReviewComments: pr.GetReviewComments(),
		Head:           pr.GetHead().GetRef(),
		Base:           pr.GetBase().GetRef(),
	}

	if pr.ClosedAt != nil {
		prInfo.ClosedAt = &pr.ClosedAt.Time
	}
	if pr.MergedAt != nil {
		prInfo.MergedAt = &pr.MergedAt.Time
	}
	if pr.Mergeable != nil {
		prInfo.Mergeable = pr.Mergeable
		prInfo.HasConflicts = !*pr.Mergeable
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    prInfo,
	})
}

// ═══════════════════════════════════════════════════════════════
// Commits Handlers
// ═══════════════════════════════════════════════════════════════

// listCommitsHandler handles GET /api/v1/github/commits
func (s *Server) listCommitsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	sinceStr := r.URL.Query().Get("since")

	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	var since *time.Time
	if sinceStr != "" {
		parsedTime, err := time.Parse(time.RFC3339, sinceStr)
		if err != nil {
			sendJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Error:   "Invalid since parameter format (use RFC3339)",
			})
			return
		}
		since = &parsedTime
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	commits, err := gs.ListCommits(s.config.Organization, repo, since)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list commits: %v", err),
		})
		return
	}

	var commitInfos []models.CommitInfo
	for _, commit := range commits {
		commitInfo := models.CommitInfo{
			SHA:     commit.GetSHA(),
			Message: commit.GetCommit().GetMessage(),
			Author:  commit.GetCommit().GetAuthor().GetName(),
			Date:    commit.GetCommit().GetAuthor().GetDate().Time,
		}

		if commit.Stats != nil {
			commitInfo.Additions = commit.Stats.GetAdditions()
			commitInfo.Deletions = commit.Stats.GetDeletions()
			commitInfo.Total = commit.Stats.GetTotal()
		}

		commitInfos = append(commitInfos, commitInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d commits", len(commitInfos)),
		Data:    commitInfos,
	})
}

// getCommitActivityHandler handles GET /api/v1/github/commits/activity
func (s *Server) getCommitActivityHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)

	// Get all commits
	allCommits, err := gs.ListCommits(s.config.Organization, repo, nil)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list commits: %v", err),
		})
		return
	}

	// Get commits from last 90 days
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	recentCommits, err := gs.ListCommits(s.config.Organization, repo, &ninetyDaysAgo)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list recent commits: %v", err),
		})
		return
	}

	// Get unique contributors
	contributorsMap := make(map[string]bool)
	for _, commit := range allCommits {
		if commit.GetAuthor() != nil {
			contributorsMap[commit.GetAuthor().GetLogin()] = true
		}
	}

	var contributors []string
	for contributor := range contributorsMap {
		contributors = append(contributors, contributor)
	}

	activity := models.CommitActivity{
		TotalCommits:      len(allCommits),
		CommitsLast90Days: len(recentCommits),
		IsActive:          len(recentCommits) > 0,
		Contributors:      contributors,
	}

	if len(allCommits) > 0 {
		activity.LastCommitDate = allCommits[0].GetCommit().GetAuthor().GetDate().Time
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    activity,
	})
}

// ═══════════════════════════════════════════════════════════════
// Issues Handlers
// ═══════════════════════════════════════════════════════════════

// listIssuesHandler handles GET /api/v1/github/issues
func (s *Server) listIssuesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	state := r.URL.Query().Get("state")
	if state == "" {
		state = "all"
	}

	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	issues, err := gs.ListIssues(s.config.Organization, repo, state)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list issues: %v", err),
		})
		return
	}

	var issueInfos []models.IssueInfo
	for _, issue := range issues {
		issueInfo := models.IssueInfo{
			Number:        issue.GetNumber(),
			Title:         issue.GetTitle(),
			State:         issue.GetState(),
			CreatedAt:     issue.GetCreatedAt().Time,
			UpdatedAt:     issue.GetUpdatedAt().Time,
			User:          issue.GetUser().GetLogin(),
			Comments:      issue.GetComments(),
			IsPullRequest: issue.IsPullRequest(),
		}

		if issue.ClosedAt != nil {
			issueInfo.ClosedAt = &issue.ClosedAt.Time
		}

		for _, label := range issue.Labels {
			issueInfo.Labels = append(issueInfo.Labels, label.GetName())
		}

		issueInfos = append(issueInfos, issueInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d issues", len(issueInfos)),
		Data:    issueInfos,
	})
}

// listIssueCommentsHandler handles GET /api/v1/github/issues/comments
func (s *Server) listIssueCommentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	numberStr := r.URL.Query().Get("number")

	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)

	var comments []*github.IssueComment
	var err error

	if numberStr != "" {
		number, parseErr := strconv.Atoi(numberStr)
		if parseErr != nil {
			sendJSON(w, http.StatusBadRequest, Response{
				Success: false,
				Error:   "Invalid issue number",
			})
			return
		}
		comments, err = gs.ListIssueComments(s.config.Organization, repo, number)
	} else {
		comments, err = gs.ListAllIssueComments(s.config.Organization, repo)
	}

	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list comments: %v", err),
		})
		return
	}

	var commentInfos []models.CommentInfo
	for _, comment := range comments {
		commentInfo := models.CommentInfo{
			ID:        comment.GetID(),
			User:      comment.GetUser().GetLogin(),
			Body:      comment.GetBody(),
			CreatedAt: comment.GetCreatedAt().Time,
			UpdatedAt: comment.GetUpdatedAt().Time,
		}
		commentInfos = append(commentInfos, commentInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d comments", len(commentInfos)),
		Data:    commentInfos,
	})
}

// ═══════════════════════════════════════════════════════════════
// Repository & README Handlers
// ═══════════════════════════════════════════════════════════════

// checkReadmeHandler handles GET /api/v1/github/readme
func (s *Server) checkReadmeHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	readme, err := gs.GetReadme(s.config.Organization, repo)

	readmeInfo := models.ReadmeInfo{
		Exists: err == nil,
	}

	if err == nil {
		readmeInfo.Name = readme.GetName()
		readmeInfo.Path = readme.GetPath()
		readmeInfo.SHA = readme.GetSHA()
		readmeInfo.Size = readme.GetSize()

		// Optionally decode content
		if r.URL.Query().Get("content") == "true" {
			content, _ := readme.GetContent()
			readmeInfo.Content = content
		}
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    readmeInfo,
	})
}

// listBranchesHandler handles GET /api/v1/github/branches
func (s *Server) listBranchesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	branches, err := gs.ListBranches(s.config.Organization, repo)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list branches: %v", err),
		})
		return
	}

	var branchInfos []models.BranchInfo
	for _, branch := range branches {
		branchInfo := models.BranchInfo{
			Name:      branch.GetName(),
			Protected: branch.GetProtected(),
			SHA:       branch.GetCommit().GetSHA(),
		}
		branchInfos = append(branchInfos, branchInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d branches", len(branchInfos)),
		Data:    branchInfos,
	})
}

// ═══════════════════════════════════════════════════════════════
// Organization Handlers
// ═══════════════════════════════════════════════════════════════

// listOrgMembersHandler handles GET /api/v1/github/org/members
func (s *Server) listOrgMembersHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	members, err := gs.ListOrgMembers(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list organization members: %v", err),
		})
		return
	}

	var memberInfos []models.OrganizationMember
	for _, member := range members {
		memberInfo := models.OrganizationMember{
			Login:     member.GetLogin(),
			ID:        member.GetID(),
			Type:      member.GetType(),
			SiteAdmin: member.GetSiteAdmin(),
		}
		memberInfos = append(memberInfos, memberInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d members", len(memberInfos)),
		Data:    memberInfos,
	})
}

// listOrgTeamsHandler handles GET /api/v1/github/org/teams
func (s *Server) listOrgTeamsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)
	teams, err := gs.ListOrgTeams(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list organization teams: %v", err),
		})
		return
	}

	var teamInfos []models.TeamInfo
	for _, team := range teams {
		teamInfo := models.TeamInfo{
			ID:           team.GetID(),
			Name:         team.GetName(),
			Slug:         team.GetSlug(),
			Description:  team.GetDescription(),
			Privacy:      team.GetPrivacy(),
			MembersCount: team.GetMembersCount(),
		}
		teamInfos = append(teamInfos, teamInfo)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Found %d teams", len(teamInfos)),
		Data:    teamInfos,
	})
}

// ═══════════════════════════════════════════════════════════════
// Repository Metrics & Score Calculation
// ═══════════════════════════════════════════════════════════════

// getRepositoryMetricsHandler handles GET /api/v1/github/metrics
func (s *Server) getRepositoryMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	repo := r.URL.Query().Get("repo")
	if repo == "" {
		sendJSON(w, http.StatusBadRequest, Response{
			Success: false,
			Error:   "repo parameter is required",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)

	metrics := models.RepositoryMetrics{
		Repository: repo,
	}

	// Check README
	hasReadme, _ := gs.CheckReadmeExists(s.config.Organization, repo)
	metrics.HasReadme = hasReadme

	// Get default branch
	defaultBranch, err := gs.GetDefaultBranch(s.config.Organization, repo)
	if err == nil {
		metrics.DefaultBranch = defaultBranch
	}

	// Get pull requests
	allPRs, err := gs.ListPullRequests(s.config.Organization, repo, "all")
	if err == nil {
		for _, pr := range allPRs {
			if pr.GetState() == "open" {
				metrics.OpenPRs++
				// Check for conflicts
				if pr.Mergeable != nil && !*pr.Mergeable {
					metrics.PRsWithConflicts++
				}
			} else if pr.GetState() == "closed" {
				if pr.MergedAt != nil {
					metrics.MergedPRs++
				} else {
					metrics.ClosedPRs++
				}
			}
		}
	}

	// Get issues (excluding PRs)
	allIssues, err := gs.ListIssues(s.config.Organization, repo, "all")
	if err == nil {
		for _, issue := range allIssues {
			if !issue.IsPullRequest() {
				if issue.GetState() == "open" {
					metrics.OpenIssues++
				} else {
					metrics.ClosedIssues++
				}
			}
		}
	}

	// Get commit activity
	ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
	allCommits, err := gs.ListCommits(s.config.Organization, repo, nil)
	if err == nil {
		metrics.TotalCommits = len(allCommits)

		if len(allCommits) > 0 {
			lastCommitDate := allCommits[0].GetCommit().GetAuthor().GetDate().Time
			metrics.LastCommitDate = &lastCommitDate
		}

		// Count commits in last 90 days
		for _, commit := range allCommits {
			commitDate := commit.GetCommit().GetAuthor().GetDate().Time
			if commitDate.After(ninetyDaysAgo) {
				metrics.CommitsLast90Days++
			}
		}

		metrics.IsActive = metrics.CommitsLast90Days > 0

		// Count unique contributors
		contributorsMap := make(map[string]bool)
		for _, commit := range allCommits {
			if commit.GetAuthor() != nil {
				contributorsMap[commit.GetAuthor().GetLogin()] = true
			}
		}
		metrics.Contributors = len(contributorsMap)
	}

	// Get branches
	branches, err := gs.ListBranches(s.config.Organization, repo)
	if err == nil {
		metrics.Branches = len(branches)
	}

	// Calculate score
	metrics.Score = calculateRepositoryScore(metrics)

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Data:    metrics,
	})
}

// calculateRepositoryScore calculates a score for the repository
func calculateRepositoryScore(metrics models.RepositoryMetrics) float64 {
	score := 0.0

	// README exists: +20 points
	if metrics.HasReadme {
		score += 20
	}

	// Activity score: +30 points max
	if metrics.IsActive {
		score += 30
		// Bonus for high activity
		if metrics.CommitsLast90Days > 50 {
			score += 10
		} else if metrics.CommitsLast90Days > 20 {
			score += 5
		}
	}

	// PR management: +20 points max
	totalPRs := metrics.OpenPRs + metrics.ClosedPRs + metrics.MergedPRs
	if totalPRs > 0 {
		mergeRate := float64(metrics.MergedPRs) / float64(totalPRs)
		score += mergeRate * 20
	}

	// Issue management: +10 points max
	totalIssues := metrics.OpenIssues + metrics.ClosedIssues
	if totalIssues > 0 {
		closeRate := float64(metrics.ClosedIssues) / float64(totalIssues)
		score += closeRate * 10
	}

	// Conflict penalty: -5 points per PR with conflicts
	score -= float64(metrics.PRsWithConflicts) * 5

	// Contributors bonus: +10 points max
	if metrics.Contributors > 10 {
		score += 10
	} else if metrics.Contributors > 5 {
		score += 5
	} else if metrics.Contributors > 0 {
		score += 2
	}

	// Ensure score is between 0 and 100
	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}

	return score
}

// getAllRepositoriesMetricsHandler handles GET /api/v1/github/metrics/all
func (s *Server) getAllRepositoriesMetricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSON(w, http.StatusMethodNotAllowed, Response{
			Success: false,
			Error:   "Method not allowed",
		})
		return
	}

	gs := services.NewGitHubService(s.config.GitHubPAT)

	repos, err := gs.ListRepositories(s.config.Organization)
	if err != nil {
		sendJSON(w, http.StatusInternalServerError, Response{
			Success: false,
			Error:   fmt.Sprintf("Failed to list repositories: %v", err),
		})
		return
	}

	var allMetrics []models.RepositoryMetrics

	for _, repo := range repos {
		if repo.GetArchived() {
			continue
		}

		repoName := repo.GetName()
		metrics := models.RepositoryMetrics{
			Repository: repoName,
		}

		// Check README
		hasReadme, _ := gs.CheckReadmeExists(s.config.Organization, repoName)
		metrics.HasReadme = hasReadme

		// Get default branch
		metrics.DefaultBranch = repo.GetDefaultBranch()

		// Get pull requests
		allPRs, err := gs.ListPullRequests(s.config.Organization, repoName, "all")
		if err == nil {
			for _, pr := range allPRs {
				if pr.GetState() == "open" {
					metrics.OpenPRs++
					if pr.Mergeable != nil && !*pr.Mergeable {
						metrics.PRsWithConflicts++
					}
				} else if pr.GetState() == "closed" {
					if pr.MergedAt != nil {
						metrics.MergedPRs++
					} else {
						metrics.ClosedPRs++
					}
				}
			}
		}

		// Get issues
		allIssues, err := gs.ListIssues(s.config.Organization, repoName, "all")
		if err == nil {
			for _, issue := range allIssues {
				if !issue.IsPullRequest() {
					if issue.GetState() == "open" {
						metrics.OpenIssues++
					} else {
						metrics.ClosedIssues++
					}
				}
			}
		}

		// Get commit activity
		ninetyDaysAgo := time.Now().AddDate(0, 0, -90)
		allCommits, err := gs.ListCommits(s.config.Organization, repoName, nil)
		if err == nil {
			metrics.TotalCommits = len(allCommits)

			if len(allCommits) > 0 {
				lastCommitDate := allCommits[0].GetCommit().GetAuthor().GetDate().Time
				metrics.LastCommitDate = &lastCommitDate
			}

			for _, commit := range allCommits {
				commitDate := commit.GetCommit().GetAuthor().GetDate().Time
				if commitDate.After(ninetyDaysAgo) {
					metrics.CommitsLast90Days++
				}
			}

			metrics.IsActive = metrics.CommitsLast90Days > 0

			contributorsMap := make(map[string]bool)
			for _, commit := range allCommits {
				if commit.GetAuthor() != nil {
					contributorsMap[commit.GetAuthor().GetLogin()] = true
				}
			}
			metrics.Contributors = len(contributorsMap)
		}

		// Get branches
		branches, err := gs.ListBranches(s.config.Organization, repoName)
		if err == nil {
			metrics.Branches = len(branches)
		}

		// Calculate score
		metrics.Score = calculateRepositoryScore(metrics)

		allMetrics = append(allMetrics, metrics)
	}

	sendJSON(w, http.StatusOK, Response{
		Success: true,
		Message: fmt.Sprintf("Analyzed %d repositories", len(allMetrics)),
		Data:    allMetrics,
	})
}

