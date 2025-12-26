package project

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, project *ProjectEntity) error
	FindByID(ctx context.Context, id string) (*ProjectEntity, error)
	FindBySlug(ctx context.Context, slug string) (*ProjectEntity, error)
	List(ctx context.Context, limit, offset int32) ([]*ProjectEntity, error)
	Count(ctx context.Context) (int32, error)
	Update(ctx context.Context, project *ProjectEntity) error
	Delete(ctx context.Context, id string) error
	FindByVCSRepo(ctx context.Context, provider constant.VcsProvider, repoID string) (*ProjectEntity, error)
}

type GitHubRepository struct {
	ID            uint64  `json:"id"`
	Name          string  `json:"name"`
	FullName      string  `json:"full_name"`
	Description   *string `json:"description"`
	DefaultBranch string  `json:"default_branch"`
	Private       bool    `json:"private"`
	CloneURL      string  `json:"clone_url"`
}

type GitHubClient interface {
	GetRepository(owner, repo, token string) (*GitHubRepository, error)
}

type Service struct {
	projectRepo  Repository
	validator    *Validator
	logger       *slog.Logger
	githubClient GitHubClient
}

func NewService(
	projectRepo Repository,
	validator *Validator,
	logger *slog.Logger,
) Service {
	return Service{
		projectRepo: projectRepo,
		validator:   validator,
		logger:      logger,
	}
}

func (s *Service) SetGitHubClient(client GitHubClient) {
	s.githubClient = client
}

func (s Service) CreateProject(ctx context.Context, input CreateProjectInput) (CreateProjectResponse, error) {
	if err := s.validator.ValidateCreateProject(ctx, input); err != nil {
		return CreateProjectResponse{}, err
	}

	gitRepoID := stringsTrimPtr(input.GitRepoID)

	if gitRepoID == nil && input.Owner != nil && input.Repo != nil && input.VcsToken != nil {
		if input.RepoProvider != nil && *input.RepoProvider == constant.VcsProviderGitHub {
			if s.githubClient != nil {
				ghRepo, err := s.githubClient.GetRepository(*input.Owner, *input.Repo, *input.VcsToken)
				if err != nil {
					s.logger.Error("failed to fetch repository from GitHub",
						slog.String("owner", *input.Owner),
						slog.String("repo", *input.Repo),
						slog.String("error", err.Error()))
					return CreateProjectResponse{}, fmt.Errorf("failed to fetch repository from GitHub: %w", err)
				}
				repoIDStr := strconv.FormatUint(ghRepo.ID, 10)
				gitRepoID = &repoIDStr
			}
		}
	}

	now := time.Now().UTC()
	p := &ProjectEntity{
		ID:                 uuid.NewString(),
		Name:               stringsTrim(input.Name),
		Slug:               stringsTrim(input.Slug),
		Description:        stringsTrimPtr(input.Description),
		DesignReferenceURL: stringsTrimPtr(input.DesignReferenceURL),
		GitRepoID:          gitRepoID,
		RepoProvider:       input.RepoProvider,
		Status:             constant.ProjectStatusActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.projectRepo.Create(ctx, p); err != nil {
		return CreateProjectResponse{}, err
	}
	return CreateProjectResponse{
		ID:                 p.ID,
		Name:               p.Name,
		Slug:               p.Slug,
		Description:        p.Description,
		DesignReferenceURL: p.DesignReferenceURL,
		GitRepoID:          p.GitRepoID,
		RepoProvider:       p.RepoProvider,
		Status:             p.Status,
		CreatedAt:          p.CreatedAt,
		UpdatedAt:          p.UpdatedAt,
	}, nil
}

func (s Service) GetProject(ctx context.Context, id string) (*GetProjectByIDResponse, error) {
	project, err := s.projectRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return &GetProjectByIDResponse{
		ID:                 project.ID,
		Name:               project.Name,
		Slug:               project.Slug,
		Description:        project.Description,
		DesignReferenceURL: project.DesignReferenceURL,
		GitRepoID:          project.GitRepoID,
		RepoProvider:       project.RepoProvider,
		Status:             project.Status,
		CreatedAt:          project.CreatedAt,
		UpdatedAt:          project.UpdatedAt,
		ArchivedAt:         project.ArchivedAt,
	}, nil
}

func (s Service) ListProjects(ctx context.Context, input ListProjectsInput) (ListProjectsResponse, error) {
	pageSize := input.PageSize
	if pageSize <= 0 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}

	offset := input.Offset
	if offset < 0 {
		offset = 0
	}

	projects, err := s.projectRepo.List(ctx, pageSize, offset)
	if err != nil {
		return ListProjectsResponse{}, err
	}

	totalCount, err := s.projectRepo.Count(ctx)
	if err != nil {
		return ListProjectsResponse{}, err
	}

	response := ListProjectsResponse{
		Projects:   make([]GetProjectByIDResponse, len(projects)),
		TotalCount: totalCount,
	}

	for i, p := range projects {
		response.Projects[i] = GetProjectByIDResponse{
			ID:                 p.ID,
			Name:               p.Name,
			Slug:               p.Slug,
			Description:        p.Description,
			DesignReferenceURL: p.DesignReferenceURL,
			GitRepoID:          p.GitRepoID,
			RepoProvider:       p.RepoProvider,
			Status:             p.Status,
			CreatedAt:          p.CreatedAt,
			UpdatedAt:          p.UpdatedAt,
			ArchivedAt:         p.ArchivedAt,
		}
	}

	return response, nil
}

func (s Service) UpdateProject(ctx context.Context, input UpdateProjectInput) (UpdateProjectResponse, error) {
	if err := s.validator.ValidateUpdateProject(ctx, input); err != nil {
		return UpdateProjectResponse{}, err
	}

	p, err := s.projectRepo.FindByID(ctx, input.ID)
	if err != nil {
		return UpdateProjectResponse{}, err
	}

	if input.Name != nil {
		p.Name = stringsTrim(*input.Name)
	}
	if input.Slug != nil {
		p.Slug = stringsTrim(*input.Slug)
	}
	if input.Description != nil {
		p.Description = stringsTrimPtr(input.Description)

	}
	if input.DesignReferenceURL != nil {
		p.DesignReferenceURL = stringsTrimPtr(input.DesignReferenceURL)
	}

	if input.GitRepoID != nil {
		p.GitRepoID = stringsTrimPtr(input.GitRepoID)
	}
	if input.RepoProvider != nil {
		p.RepoProvider = input.RepoProvider
	}
	if input.Status != nil {
		p.Status = *input.Status
		if *input.Status == constant.ProjectStatusArchived {
			now := time.Now().UTC()
			p.ArchivedAt = &now
		} else {
			p.ArchivedAt = nil
		}
	}

	p.UpdatedAt = time.Now().UTC()

	if err := s.projectRepo.Update(ctx, p); err != nil {
		return UpdateProjectResponse{}, err
	}

	return UpdateProjectResponse{
		ID: p.ID,
	}, nil
}

func (s Service) DeleteProject(ctx context.Context, id string) error {
	return s.projectRepo.Delete(ctx, id)
}

func (s Service) GetProjectByVCSRepo(ctx context.Context, req GetProjectByVCSRepoRequest) (*GetProjectByVCSRepoResponse, error) {
	project, err := s.projectRepo.FindByVCSRepo(ctx, req.Provider, req.RepoID)
	if err != nil {
		s.logger.Error("failed to get project by VCS repo", "error", err, "provider", req.Provider, "repo_id", req.RepoID)
		return nil, err
	}

	return &GetProjectByVCSRepoResponse{
		ID:           project.ID,
		Slug:         project.Slug,
		Name:         project.Name,
		RepoProvider: project.RepoProvider,
		GitRepoID:    project.GitRepoID,
	}, nil
}

func (s Service) CreateProjectFromGitHub(ctx context.Context, input CreateProjectFromGitHubInput) (CreateProjectFromGitHubResponse, error) {
	if s.githubClient == nil {
		return CreateProjectFromGitHubResponse{}, fmt.Errorf("github client is not configured")
	}

	if input.Owner == "" || input.Repo == "" {
		return CreateProjectFromGitHubResponse{}, fmt.Errorf("owner and repo are required")
	}

	if input.Token == "" {
		return CreateProjectFromGitHubResponse{}, fmt.Errorf("github token is required")
	}

	ghRepo, err := s.githubClient.GetRepository(input.Owner, input.Repo, input.Token)
	if err != nil {
		s.logger.Error("failed to fetch repository from GitHub",
			slog.String("owner", input.Owner),
			slog.String("repo", input.Repo),
			slog.String("error", err.Error()))
		return CreateProjectFromGitHubResponse{}, fmt.Errorf("failed to fetch repository from GitHub: %w", err)
	}

	gitRepoID := strconv.FormatUint(ghRepo.ID, 10)
	provider := constant.VcsProviderGitHub
	slug := strings.ToLower(strings.ReplaceAll(ghRepo.Name, "_", "-"))

	now := time.Now().UTC()
	p := &ProjectEntity{
		ID:           uuid.NewString(),
		Name:         ghRepo.Name,
		Slug:         slug,
		Description:  ghRepo.Description,
		GitRepoID:    &gitRepoID,
		RepoProvider: &provider,
		Status:       constant.ProjectStatusActive,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.projectRepo.Create(ctx, p); err != nil {
		return CreateProjectFromGitHubResponse{}, err
	}

	return CreateProjectFromGitHubResponse{
		ID:           p.ID,
		Name:         p.Name,
		Slug:         p.Slug,
		Description:  p.Description,
		GitRepoID:    p.GitRepoID,
		RepoProvider: p.RepoProvider,
		Status:       p.Status,
		CreatedAt:    p.CreatedAt,
		UpdatedAt:    p.UpdatedAt,
	}, nil
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
