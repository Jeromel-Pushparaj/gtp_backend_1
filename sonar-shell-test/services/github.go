package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/google/go-github/v58/github"
	"golang.org/x/crypto/nacl/box"
	"golang.org/x/oauth2"
)

// GitHubService wraps the GitHub API client
type GitHubService struct {
	client *github.Client
	ctx    context.Context
}

// NewGitHubService creates a new GitHub API service
func NewGitHubService(token string) *GitHubService {
	ctx := context.Background()
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(ctx, ts)

	return &GitHubService{
		client: github.NewClient(tc),
		ctx:    ctx,
	}
}

// GetRepository gets a single repository
func (gs *GitHubService) GetRepository(org, repo string) (*github.Repository, error) {
	repository, _, err := gs.client.Repositories.Get(gs.ctx, org, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	return repository, nil
}

// ListRepositories lists all repositories in an organization
func (gs *GitHubService) ListRepositories(org string) ([]*github.Repository, error) {
	opt := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allRepos []*github.Repository
	for {
		repos, resp, err := gs.client.Repositories.ListByOrg(gs.ctx, org, opt)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}
		allRepos = append(allRepos, repos...)
		if resp.NextPage == 0 {
			break
		}
		opt.Page = resp.NextPage
	}

	return allRepos, nil
}

// CheckFileExists checks if a file exists in a repository
func (gs *GitHubService) CheckFileExists(owner, repo, path, branch string) (bool, error) {
	opts := &github.RepositoryContentGetOptions{Ref: branch}
	_, _, resp, err := gs.client.Repositories.GetContents(gs.ctx, owner, repo, path, opts)

	if err != nil {
		if resp != nil && resp.StatusCode == 404 {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file: %w", err)
	}

	return true, nil
}

// CreateFile creates a new file in a repository
func (gs *GitHubService) CreateFile(owner, repo, path, message, content, branch string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
	}

	_, _, err := gs.client.Repositories.CreateFile(gs.ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}

	return nil
}

// GetPublicKey retrieves the public key for encrypting secrets
func (gs *GitHubService) GetPublicKey(owner, repo string) (*github.PublicKey, error) {
	publicKey, _, err := gs.client.Actions.GetRepoPublicKey(gs.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get public key: %w", err)
	}
	return publicKey, nil
}

// EncryptSecret encrypts a secret using the repository's public key
func (gs *GitHubService) EncryptSecret(publicKey *github.PublicKey, secretValue string) (string, error) {
	decodedKey, err := base64.StdEncoding.DecodeString(publicKey.GetKey())
	if err != nil {
		return "", fmt.Errorf("failed to decode public key: %w", err)
	}

	var publicKeyBytes [32]byte
	copy(publicKeyBytes[:], decodedKey)

	secretBytes := []byte(secretValue)
	encrypted, err := box.SealAnonymous(nil, secretBytes, &publicKeyBytes, nil)
	if err != nil {
		return "", fmt.Errorf("failed to encrypt secret: %w", err)
	}

	return base64.StdEncoding.EncodeToString(encrypted), nil
}

// CreateOrUpdateSecret creates or updates a repository secret
func (gs *GitHubService) CreateOrUpdateSecret(owner, repo, secretName, encryptedValue, keyID string) error {
	secret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedValue,
	}

	_, err := gs.client.Actions.CreateOrUpdateRepoSecret(gs.ctx, owner, repo, secret)
	if err != nil {
		return fmt.Errorf("failed to create/update secret %s: %w", secretName, err)
	}

	return nil
}

// GetDefaultBranch gets the default branch of a repository
func (gs *GitHubService) GetDefaultBranch(owner, repo string) (string, error) {
	repository, _, err := gs.client.Repositories.Get(gs.ctx, owner, repo)
	if err != nil {
		return "", fmt.Errorf("failed to get repository: %w", err)
	}
	return repository.GetDefaultBranch(), nil
}

// GetFileContent gets the content and SHA of a file
func (gs *GitHubService) GetFileContent(owner, repo, path, branch string) (content string, sha string, err error) {
	opts := &github.RepositoryContentGetOptions{Ref: branch}
	fileContent, _, _, err := gs.client.Repositories.GetContents(gs.ctx, owner, repo, path, opts)
	if err != nil {
		return "", "", fmt.Errorf("failed to get file content: %w", err)
	}

	decodedContent, err := fileContent.GetContent()
	if err != nil {
		return "", "", fmt.Errorf("failed to decode file content: %w", err)
	}

	return decodedContent, fileContent.GetSHA(), nil
}

// UpdateFile updates an existing file in a repository
func (gs *GitHubService) UpdateFile(owner, repo, path, message, content, branch, sha string) error {
	opts := &github.RepositoryContentFileOptions{
		Message: github.String(message),
		Content: []byte(content),
		Branch:  github.String(branch),
		SHA:     github.String(sha),
	}

	_, _, err := gs.client.Repositories.UpdateFile(gs.ctx, owner, repo, path, opts)
	if err != nil {
		return fmt.Errorf("failed to update file: %w", err)
	}

	return nil
}

// ListSecrets lists all secrets in a repository (returns names and metadata only)
func (gs *GitHubService) ListSecrets(owner, repo string) ([]*github.Secret, error) {
	opts := &github.ListOptions{PerPage: 100}

	secrets, _, err := gs.client.Actions.ListRepoSecrets(gs.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list secrets: %w", err)
	}

	return secrets.Secrets, nil
}

// ListEnvironments lists all environments in a repository
func (gs *GitHubService) ListEnvironments(owner, repo string) ([]*github.Environment, error) {
	opts := &github.EnvironmentListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	envs, _, err := gs.client.Repositories.ListEnvironments(gs.ctx, owner, repo, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments: %w", err)
	}

	return envs.Environments, nil
}

// CreateEnvironment creates a new environment in a repository
func (gs *GitHubService) CreateEnvironment(owner, repo, envName string) error {
	_, _, err := gs.client.Repositories.CreateUpdateEnvironment(gs.ctx, owner, repo, envName, nil)
	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}
	return nil
}

// GetEnvironmentPublicKey retrieves the public key for encrypting environment secrets
func (gs *GitHubService) GetEnvironmentPublicKey(owner, repo, envName string) (*github.PublicKey, error) {
	// First get the repository to get its ID
	repository, _, err := gs.client.Repositories.Get(gs.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	repoID := int(repository.GetID())

	publicKey, _, err := gs.client.Actions.GetEnvPublicKey(gs.ctx, repoID, envName)
	if err != nil {
		return nil, fmt.Errorf("failed to get environment public key: %w", err)
	}
	return publicKey, nil
}

// CreateOrUpdateEnvSecret creates or updates an environment secret
func (gs *GitHubService) CreateOrUpdateEnvSecret(owner, repo, envName, secretName, encryptedValue, keyID string) error {
	// First get the repository to get its ID
	repository, _, err := gs.client.Repositories.Get(gs.ctx, owner, repo)
	if err != nil {
		return fmt.Errorf("failed to get repository: %w", err)
	}
	repoID := int(repository.GetID())

	secret := &github.EncryptedSecret{
		Name:           secretName,
		KeyID:          keyID,
		EncryptedValue: encryptedValue,
	}

	_, err = gs.client.Actions.CreateOrUpdateEnvSecret(gs.ctx, repoID, envName, secret)
	if err != nil {
		return fmt.Errorf("failed to create/update environment secret %s: %w", secretName, err)
	}

	return nil
}

// ListEnvSecrets lists all secrets in an environment
func (gs *GitHubService) ListEnvSecrets(owner, repo, envName string) ([]*github.Secret, error) {
	// First get the repository to get its ID
	repository, _, err := gs.client.Repositories.Get(gs.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get repository: %w", err)
	}
	repoID := int(repository.GetID())

	opts := &github.ListOptions{PerPage: 100}

	secrets, _, err := gs.client.Actions.ListEnvSecrets(gs.ctx, repoID, envName, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to list environment secrets: %w", err)
	}

	return secrets.Secrets, nil
}

// ═══════════════════════════════════════════════════════════════
// GitHub Metrics Service Methods
// ═══════════════════════════════════════════════════════════════

// ListPullRequests lists pull requests for a repository
func (gs *GitHubService) ListPullRequests(owner, repo, state string) ([]*github.PullRequest, error) {
	opts := &github.PullRequestListOptions{
		State:       state, // "open", "closed", "all"
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allPRs []*github.PullRequest
	for {
		prs, resp, err := gs.client.PullRequests.List(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list pull requests: %w", err)
		}
		allPRs = append(allPRs, prs...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allPRs, nil
}

// GetPullRequest gets a specific pull request
func (gs *GitHubService) GetPullRequest(owner, repo string, number int) (*github.PullRequest, error) {
	pr, _, err := gs.client.PullRequests.Get(gs.ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get pull request: %w", err)
	}
	return pr, nil
}

// ListPRCommits lists commits for a pull request
func (gs *GitHubService) ListPRCommits(owner, repo string, number int) ([]*github.RepositoryCommit, error) {
	opts := &github.ListOptions{PerPage: 100}

	var allCommits []*github.RepositoryCommit
	for {
		commits, resp, err := gs.client.PullRequests.ListCommits(gs.ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list PR commits: %w", err)
		}
		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}

// ListPRFiles lists files changed in a pull request
func (gs *GitHubService) ListPRFiles(owner, repo string, number int) ([]*github.CommitFile, error) {
	opts := &github.ListOptions{PerPage: 100}

	var allFiles []*github.CommitFile
	for {
		files, resp, err := gs.client.PullRequests.ListFiles(gs.ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list PR files: %w", err)
		}
		allFiles = append(allFiles, files...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allFiles, nil
}

// ListCommits lists commits for a repository
func (gs *GitHubService) ListCommits(owner, repo string, since *time.Time) ([]*github.RepositoryCommit, error) {
	opts := &github.CommitsListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	if since != nil {
		opts.Since = *since
	}

	var allCommits []*github.RepositoryCommit
	for {
		commits, resp, err := gs.client.Repositories.ListCommits(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list commits: %w", err)
		}
		allCommits = append(allCommits, commits...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCommits, nil
}

// GetCommit gets a specific commit
func (gs *GitHubService) GetCommit(owner, repo, sha string) (*github.RepositoryCommit, error) {
	commit, _, err := gs.client.Repositories.GetCommit(gs.ctx, owner, repo, sha, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get commit: %w", err)
	}
	return commit, nil
}

// ListBranches lists all branches in a repository
func (gs *GitHubService) ListBranches(owner, repo string) ([]*github.Branch, error) {
	opts := &github.BranchListOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allBranches []*github.Branch
	for {
		branches, resp, err := gs.client.Repositories.ListBranches(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list branches: %w", err)
		}
		allBranches = append(allBranches, branches...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allBranches, nil
}

// GetBranch gets a specific branch
func (gs *GitHubService) GetBranch(owner, repo, branch string) (*github.Branch, error) {
	branchInfo, _, err := gs.client.Repositories.GetBranch(gs.ctx, owner, repo, branch, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch: %w", err)
	}
	return branchInfo, nil
}

// GetReadme gets the README file for a repository
func (gs *GitHubService) GetReadme(owner, repo string) (*github.RepositoryContent, error) {
	readme, _, err := gs.client.Repositories.GetReadme(gs.ctx, owner, repo, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get README: %w", err)
	}
	return readme, nil
}

// CheckReadmeExists checks if a README file exists
func (gs *GitHubService) CheckReadmeExists(owner, repo string) (bool, error) {
	_, _, err := gs.client.Repositories.GetReadme(gs.ctx, owner, repo, nil)
	if err != nil {
		if _, ok := err.(*github.ErrorResponse); ok {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// ListIssues lists issues for a repository
func (gs *GitHubService) ListIssues(owner, repo, state string) ([]*github.Issue, error) {
	opts := &github.IssueListByRepoOptions{
		State:       state, // "open", "closed", "all"
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allIssues []*github.Issue
	for {
		issues, resp, err := gs.client.Issues.ListByRepo(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issues: %w", err)
		}
		allIssues = append(allIssues, issues...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allIssues, nil
}

// ListContributors lists contributors for a repository
func (gs *GitHubService) ListContributors(owner, repo string) ([]*github.Contributor, error) {
	opts := &github.ListContributorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allContributors []*github.Contributor
	for {
		contributors, resp, err := gs.client.Repositories.ListContributors(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list contributors: %w", err)
		}
		allContributors = append(allContributors, contributors...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allContributors, nil
}

// ListCollaborators lists collaborators for a repository with their permissions
func (gs *GitHubService) ListCollaborators(owner, repo string) ([]*github.User, error) {
	opts := &github.ListCollaboratorsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allCollaborators []*github.User
	for {
		collaborators, resp, err := gs.client.Repositories.ListCollaborators(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list collaborators: %w", err)
		}
		allCollaborators = append(allCollaborators, collaborators...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allCollaborators, nil
}

// GetRepositoryOwner gets the user with admin/owner permissions for a repository
func (gs *GitHubService) GetRepositoryOwner(owner, repo string) (string, error) {
	// First try to get collaborators with admin permission
	opts := &github.ListCollaboratorsOptions{
		Affiliation: "admin",
		ListOptions: github.ListOptions{PerPage: 1},
	}

	collaborators, _, err := gs.client.Repositories.ListCollaborators(gs.ctx, owner, repo, opts)
	if err != nil {
		// If we can't get collaborators, return the org/owner name
		return owner, nil
	}

	if len(collaborators) > 0 {
		return collaborators[0].GetLogin(), nil
	}

	return owner, nil
}

// GetIssue gets a specific issue
func (gs *GitHubService) GetIssue(owner, repo string, number int) (*github.Issue, error) {
	issue, _, err := gs.client.Issues.Get(gs.ctx, owner, repo, number)
	if err != nil {
		return nil, fmt.Errorf("failed to get issue: %w", err)
	}
	return issue, nil
}

// ListIssueComments lists comments for an issue
func (gs *GitHubService) ListIssueComments(owner, repo string, number int) ([]*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allComments []*github.IssueComment
	for {
		comments, resp, err := gs.client.Issues.ListComments(gs.ctx, owner, repo, number, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issue comments: %w", err)
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// ListAllIssueComments lists all issue comments for a repository
func (gs *GitHubService) ListAllIssueComments(owner, repo string) ([]*github.IssueComment, error) {
	opts := &github.IssueListCommentsOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allComments []*github.IssueComment
	for {
		comments, resp, err := gs.client.Issues.ListComments(gs.ctx, owner, repo, 0, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list all issue comments: %w", err)
		}
		allComments = append(allComments, comments...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allComments, nil
}

// ListIssueEvents lists events for a repository
func (gs *GitHubService) ListIssueEvents(owner, repo string) ([]*github.IssueEvent, error) {
	opts := &github.ListOptions{PerPage: 100}

	var allEvents []*github.IssueEvent
	for {
		events, resp, err := gs.client.Issues.ListRepositoryEvents(gs.ctx, owner, repo, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list issue events: %w", err)
		}
		allEvents = append(allEvents, events...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allEvents, nil
}

// ListOrgMembers lists all members of an organization
func (gs *GitHubService) ListOrgMembers(org string) ([]*github.User, error) {
	opts := &github.ListMembersOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	var allMembers []*github.User
	for {
		members, resp, err := gs.client.Organizations.ListMembers(gs.ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list organization members: %w", err)
		}
		allMembers = append(allMembers, members...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allMembers, nil
}

// CheckOrgMembership checks if a user is a member of an organization
func (gs *GitHubService) CheckOrgMembership(org, username string) (bool, error) {
	isMember, _, err := gs.client.Organizations.IsMember(gs.ctx, org, username)
	if err != nil {
		return false, fmt.Errorf("failed to check organization membership: %w", err)
	}
	return isMember, nil
}

// ListOrgTeams lists all teams in an organization
func (gs *GitHubService) ListOrgTeams(org string) ([]*github.Team, error) {
	opts := &github.ListOptions{PerPage: 100}

	var allTeams []*github.Team
	for {
		teams, resp, err := gs.client.Teams.ListTeams(gs.ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list organization teams: %w", err)
		}
		allTeams = append(allTeams, teams...)
		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return allTeams, nil
}

// GetTeam gets a team by slug
func (gs *GitHubService) GetTeam(org, slug string) (*github.Team, error) {
	team, _, err := gs.client.Teams.GetTeamBySlug(gs.ctx, org, slug)
	if err != nil {
		return nil, fmt.Errorf("failed to get team: %w", err)
	}
	return team, nil
}

// GetUser gets user details
func (gs *GitHubService) GetUser(username string) (*github.User, error) {
	user, _, err := gs.client.Users.Get(gs.ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

// CompareCommits compares two commits
func (gs *GitHubService) CompareCommits(owner, repo, base, head string) (*github.CommitsComparison, error) {
	comparison, _, err := gs.client.Repositories.CompareCommits(gs.ctx, owner, repo, base, head, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to compare commits: %w", err)
	}
	return comparison, nil
}

// GetRepositoryLanguages fetches the programming languages used in a repository
func (gs *GitHubService) GetRepositoryLanguages(owner, repo string) (map[string]int, error) {
	languages, _, err := gs.client.Repositories.ListLanguages(gs.ctx, owner, repo)
	if err != nil {
		return nil, fmt.Errorf("failed to get languages: %w", err)
	}

	return languages, nil
}

// GetPrimaryLanguage determines the primary language from language stats
func (gs *GitHubService) GetPrimaryLanguage(languages map[string]int) string {
	if len(languages) == 0 {
		return ""
	}

	var primaryLang string
	var maxBytes int

	for lang, bytes := range languages {
		if bytes > maxBytes {
			maxBytes = bytes
			primaryLang = lang
		}
	}

	return primaryLang
}
