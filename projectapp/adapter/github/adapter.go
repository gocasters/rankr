package github

import (
	"github.com/gocasters/rankr/adapter/webhook/github"
	"github.com/gocasters/rankr/projectapp/service/project"
)

type Adapter struct {
	client *github.GitHubClient
}

func NewAdapter() *Adapter {
	return &Adapter{
		client: github.NewGitHubClient(),
	}
}

func (a *Adapter) GetRepository(owner, repo, token string) (*project.GitHubRepository, error) {
	ghRepo, err := a.client.GetRepository(owner, repo, token)
	if err != nil {
		return nil, err
	}

	return &project.GitHubRepository{
		ID:            ghRepo.ID,
		Name:          ghRepo.Name,
		FullName:      ghRepo.FullName,
		Description:   ghRepo.Description,
		DefaultBranch: ghRepo.DefaultBranch,
		Private:       ghRepo.Private,
		CloneURL:      ghRepo.CloneURL,
	}, nil
}
