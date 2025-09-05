package versioncontrollersystemproject

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, versionControllerSystemProjectEntity *VersionControllerSystemProjectEntity) (*VersionControllerSystemProjectEntity, error)
	FindByID(ctx context.Context, id string) (*VersionControllerSystemProjectEntity, error)
	FindByProjectID(ctx context.Context, projectID string) ([]*VersionControllerSystemProjectEntity, error)
	FindByProviderID(ctx context.Context, provider constant.VcsProvider, providerRepoID string, projectID string) (*VersionControllerSystemProjectEntity, error)
	Update(ctx context.Context, repo *VersionControllerSystemProjectEntity) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context) ([]*VersionControllerSystemProjectEntity, error)
}

type Service struct {
	VersionControllerSystemProject Repository
	validator                      *Validator
	logger                         *slog.Logger
}

func NewService(
	VersionControllerSystemProject Repository,
	validator *Validator,
	logger *slog.Logger,
) Service {
	return Service{
		VersionControllerSystemProject: VersionControllerSystemProject,
		validator:                      validator,
		logger:                         logger,
	}
}

func (s Service) CreateVersionControllerSystemProject(ctx context.Context, input CreateVersionControllerSystemProjectInput) (CreateVersionControllerSystemProjectResponse, error) {

	if err := s.validator.ValidateCreateVersionControllerSystemProject(input); err != nil {
		return CreateVersionControllerSystemProjectResponse{}, err
	}

	now := time.Now().UTC()
	repo := &VersionControllerSystemProjectEntity{
		ID:             uuid.NewString(),
		ProjectID:      input.ProjectID,
		Provider:       input.Provider,
		ProviderRepoID: stringsTrim(input.ProviderRepoID),
		Owner:          stringsTrim(input.Owner),
		Name:           stringsTrim(input.Name),
		RemoteURL:      stringsTrim(input.RemoteURL),
		DefaultBranch:  stringsTrimPtr(input.DefaultBranch),
		Visibility:     input.Visibility,
		InstallationID: stringsTrimPtr(input.InstallationID),
		CreatedAt:      now,
		UpdatedAt:      now,
	}

	entity, err := s.VersionControllerSystemProject.Create(ctx, repo)
	if err != nil {
		return CreateVersionControllerSystemProjectResponse{}, err
	}
	return CreateVersionControllerSystemProjectResponse{
		ID:             entity.ID,
		ProjectID:      entity.ProjectID,
		Provider:       entity.Provider,
		ProviderRepoID: entity.ProviderRepoID,
		Owner:          entity.Owner,
		Name:           entity.Name,
		RemoteURL:      entity.RemoteURL,
		DefaultBranch:  entity.DefaultBranch,
		Visibility:     entity.Visibility,
		InstallationID: entity.InstallationID,
		LastSyncedAt:   entity.LastSyncedAt,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}, nil
}

func (s Service) GetVcsRepo(ctx context.Context, id string) (*GetVersionControllerSystemProjectByIDResponse, error) {
	entity, err := s.VersionControllerSystemProject.FindByID(ctx, id)

	return &GetVersionControllerSystemProjectByIDResponse{
		ID:             entity.ID,
		ProjectID:      entity.ProjectID,
		Provider:       entity.Provider,
		ProviderRepoID: entity.ProviderRepoID,
		Owner:          entity.Owner,
		Name:           entity.Name,
		RemoteURL:      entity.RemoteURL,
		DefaultBranch:  entity.DefaultBranch,
		Visibility:     entity.Visibility,
		InstallationID: entity.InstallationID,
		LastSyncedAt:   entity.LastSyncedAt,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}, err
}

func (s Service) GetVcsRepoByProviderID(ctx context.Context, provider constant.VcsProvider, providerRepoID string, projectID string) (*GetVersionControllerSystemProjectByIDResponse, error) {
	entity, err := s.VersionControllerSystemProject.FindByProviderID(ctx, provider, providerRepoID, projectID)
	return &GetVersionControllerSystemProjectByIDResponse{
		ID:             entity.ID,
		ProjectID:      entity.ProjectID,
		Provider:       entity.Provider,
		ProviderRepoID: entity.ProviderRepoID,
		Owner:          entity.Owner,
		Name:           entity.Name,
		RemoteURL:      entity.RemoteURL,
		DefaultBranch:  entity.DefaultBranch,
		Visibility:     entity.Visibility,
		InstallationID: entity.InstallationID,
		LastSyncedAt:   entity.LastSyncedAt,
		CreatedAt:      entity.CreatedAt,
		UpdatedAt:      entity.UpdatedAt,
	}, err
}

func (s Service) GetVcsReposByProject(ctx context.Context, projectID string) (GetVersionControllerSystemProjectListedResponse, error) {

	data, err := s.VersionControllerSystemProject.FindByProjectID(ctx, projectID)

	return GetVersionControllerSystemProjectListedResponse{
		Items: data,
	}, err
}

func (s Service) ListVcsRepo(ctx context.Context) (ListVersionControllerSystemProjectsResponse, error) {

	repos, err := s.VersionControllerSystemProject.List(ctx)
	if err != nil {
		return ListVersionControllerSystemProjectsResponse{}, err
	}
	return ListVersionControllerSystemProjectsResponse{
		VersionControllerSystemProjects: repos,
	}, nil

	return ListVersionControllerSystemProjectsResponse{}, nil
}

func (s Service) UpdateVcsRepo(ctx context.Context, input UpdateVersionControllerSystemProjectInput) (UpdateVersionControllerSystemProjectResponse, error) {

	if err := s.validator.ValidateUpdateVcsRepo(input); err != nil {
		return UpdateVersionControllerSystemProjectResponse{}, err
	}

	repo, err := s.VersionControllerSystemProject.FindByID(ctx, input.ID)
	if err != nil {
		return UpdateVersionControllerSystemProjectResponse{}, err
	}

	if input.Owner != nil {
		repo.Owner = stringsTrim(*input.Owner)
	}
	if input.Name != nil {
		repo.Name = stringsTrim(*input.Name)
	}
	if input.RemoteURL != nil {
		repo.RemoteURL = stringsTrim(*input.RemoteURL)
	}
	if input.DefaultBranch != nil {
		repo.DefaultBranch = stringsTrimPtr(*input.DefaultBranch)
	}
	if input.Visibility != nil {
		repo.Visibility = *input.Visibility
	}
	if input.InstallationID != nil {
		repo.InstallationID = stringsTrimPtr(*input.InstallationID)
	}

	repo.UpdatedAt = time.Now().UTC()

	if err := s.VersionControllerSystemProject.Update(ctx, repo); err != nil {
		return UpdateVersionControllerSystemProjectResponse{}, err
	}

	return UpdateVersionControllerSystemProjectResponse{
		Data: repo,
	}, nil
}

func (s Service) DeleteVcsRepo(ctx context.Context, id string) error {
	return s.VersionControllerSystemProject.Delete(ctx, id)
}

func stringsTrim(s string) string {
	return strings.TrimSpace(s)
}

func stringsTrimPtr(p *string) *string {
	if p == nil {
		return nil
	}
	t := strings.TrimSpace(*p)
	if t == "" {
		return nil
	}
	return &t
}
