package services

import (
	"context"
	"time"

	interfaces2 "github.com/gocasters/rankr/projectsapp/interfaces"
	"github.com/gocasters/rankr/projectsapp/types"
	"github.com/google/uuid"
)

type VcsRepoService struct {
	vcsRepos  interfaces2.VcsRepoRepository
	projects  interfaces2.IProjectRepository
	validator *Validator
}

func NewVcsRepoService(vcsRepos interfaces2.VcsRepoRepository, projects interfaces2.IProjectRepository, validator *Validator) VcsRepoService {
	return VcsRepoService{
		vcsRepos:  vcsRepos,
		projects:  projects,
		validator: validator,
	}
}

func (s VcsRepoService) CreateVcsRepo(ctx context.Context, in interfaces2.CreateVcsRepoInput) (*types.VcsRepoEntity, error) {

	if err := s.validator.ValidateCreateVcsRepo(ctx, in); err != nil {
		return nil, err
	}

	if _, err := s.projects.FindByID(ctx, in.ProjectID); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	repo := &types.VcsRepoEntity{
		ID:             uuid.NewString(),
		ProjectID:      in.ProjectID,
		Provider:       in.Provider,
		ProviderRepoID: stringsTrim(in.ProviderRepoID),
		Owner:          stringsTrim(in.Owner),
		Name:           stringsTrim(in.Name),
		RemoteURL:      stringsTrim(in.RemoteURL),
		DefaultBranch:  stringsTrimPtr(in.DefaultBranch),
		Visibility:     in.Visibility,
		InstallationID: stringsTrimPtr(in.InstallationID),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	if err := s.vcsRepos.Create(ctx, repo); err != nil {
		return nil, err
	}
	return repo, nil
}

func (s VcsRepoService) GetVcsRepo(ctx context.Context, id string) (*types.VcsRepoEntity, error) {
	return s.vcsRepos.FindByID(ctx, id)
}

func (s VcsRepoService) GetVcsReposByProject(ctx context.Context, projectID string) ([]*types.VcsRepoEntity, error) {

	if _, err := s.projects.FindByID(ctx, projectID); err != nil {
		return nil, err
	}

	return s.vcsRepos.FindByProjectID(ctx, projectID)
}

func (s VcsRepoService) GetVcsRepoByProviderID(ctx context.Context, provider types.VcsProvider, providerRepoID string, projectID string) (*types.VcsRepoEntity, error) {
	return s.vcsRepos.FindByProviderID(ctx, provider, providerRepoID, projectID)
}

func (s VcsRepoService) UpdateVcsRepo(ctx context.Context, in interfaces2.UpdateVcsRepoInput) (*types.VcsRepoEntity, error) {

	if err := s.validator.ValidateUpdateVcsRepo(ctx, in); err != nil {
		return nil, err
	}

	repo, err := s.vcsRepos.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Owner != nil {
		repo.Owner = stringsTrim(*in.Owner)
	}
	if in.Name != nil {
		repo.Name = stringsTrim(*in.Name)
	}
	if in.RemoteURL != nil {
		repo.RemoteURL = stringsTrim(*in.RemoteURL)
	}
	if in.DefaultBranch != nil {
		repo.DefaultBranch = stringsTrimPtr(*in.DefaultBranch)
	}
	if in.Visibility != nil {
		repo.Visibility = *in.Visibility
	}
	if in.InstallationID != nil {
		repo.InstallationID = stringsTrimPtr(*in.InstallationID)
	}
	if in.LastSyncedAt != nil {
		repo.LastSyncedAt = *in.LastSyncedAt
	}

	repo.UpdatedAt = time.Now().UTC()

	if err := s.vcsRepos.Update(ctx, repo); err != nil {
		return nil, err
	}
	return repo, nil
}

func (s VcsRepoService) DeleteVcsRepo(ctx context.Context, id string) error {
	return s.vcsRepos.Delete(ctx, id)
}

func (s VcsRepoService) ListVcsRepo(ctx context.Context) ([]*types.VcsRepoEntity, error) {
	repos, err := s.vcsRepos.List(ctx)
	if err != nil {
		return nil, err
	}
	return repos, nil

}
