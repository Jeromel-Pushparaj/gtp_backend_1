package clients

import (
	"context"
	"pd-service/models"

	"github.com/google/go-github/v57/github"
	"golang.org/x/oauth2"
)

type GitHubClient struct {
	client *github.Client
}

func NewGitHubClient(token string) *GitHubClient {
	ts := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	tc := oauth2.NewClient(context.Background(), ts)
	
	return &GitHubClient{
		client: github.NewClient(tc),
	}
}

func (g *GitHubClient) ListRepositories(ctx context.Context, org string) ([]*models.GitHubRepo, error) {
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}
	
	repos, _, err := g.client.Repositories.ListByOrg(ctx, org, opts)
	if err != nil {
		return nil, err
	}
	
	var result []*models.GitHubRepo
	for _, repo := range repos {
		result = append(result, &models.GitHubRepo{
			Name:     repo.GetName(),
			FullName: repo.GetFullName(),
			URL:      repo.GetHTMLURL(),
		})
	}
	
	return result, nil
}

func (g *GitHubClient) GetRepository(ctx context.Context, owner, repo string) (*models.GitHubRepo, error) {
	repository, _, err := g.client.Repositories.Get(ctx, owner, repo)
	if err != nil {
		return nil, err
	}
	
	return &models.GitHubRepo{
		Name:     repository.GetName(),
		FullName: repository.GetFullName(),
		URL:      repository.GetHTMLURL(),
	}, nil
}

