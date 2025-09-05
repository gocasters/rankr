package project

import (
	"context"
	"log"
	"log/slog"
	"strings"
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/google/uuid"
)

type Repository interface {
	Create(ctx context.Context, project *ProjectEntity) error
	FindByID(ctx context.Context, id string) (*ProjectEntity, error)
	FindBySlug(ctx context.Context, slug string) (*ProjectEntity, error)
	List(ctx context.Context) ([]*ProjectEntity, error)
	Update(ctx context.Context, project *ProjectEntity) error
	Delete(ctx context.Context, id string) error
}

type Service struct {
	projectRepo Repository
	validator   *Validator
	logger      *slog.Logger
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

func (s Service) CreateProject(ctx context.Context, input CreateProjectInput) (CreateProjectResponse, error) {
	if err := s.validator.ValidateCreateProject(ctx, input); err != nil {
		return CreateProjectResponse{}, err
	}

	now := time.Now().UTC()
	p := &ProjectEntity{
		ID:                 uuid.NewString(),
		Name:               stringsTrim(input.Name),
		Slug:               stringsTrim(input.Slug),
		Description:        stringsTrimPtr(input.Description),
		DesignReferenceURL: stringsTrimPtr(input.DesignReferenceURL),
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
		Status:             project.Status,
		CreatedAt:          project.CreatedAt,
		UpdatedAt:          project.UpdatedAt,
		ArchivedAt:         project.ArchivedAt,
	}, nil
}

func (s Service) ListProjects(ctx context.Context) (ListProjectsResponse, error) {
	projects, err := s.projectRepo.List(ctx)
	if err != nil {
		return ListProjectsResponse{}, err
	}

	log.Printf("Retrieved %d projects", len(projects))

	log.Printf("Projects: %+v", projects)

	response := ListProjectsResponse{
		Projects: make([]GetProjectByIDResponse, len(projects)),
	}

	for i, p := range projects {
		response.Projects[i] = GetProjectByIDResponse{
			ID:                 p.ID,
			Name:               p.Name,
			Slug:               p.Slug,
			Description:        p.Description,
			DesignReferenceURL: p.DesignReferenceURL,
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
		p.Description = stringsTrimPtr(*input.Description)
	}
	if input.DesignReferenceURL != nil {
		p.DesignReferenceURL = stringsTrimPtr(*input.DesignReferenceURL)
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
