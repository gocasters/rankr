package services

import (
	"context"
	"time"

	interfaces2 "github.com/gocasters/rankr/projectsapp/interfaces"
	"github.com/gocasters/rankr/projectsapp/types"
	"github.com/google/uuid"
)

type ProjectService struct {
	projects  interfaces2.IProjectRepository
	validator *Validator
}

func NewProjectService(projects interfaces2.IProjectRepository, validator *Validator) ProjectService {
	return ProjectService{
		projects:  projects,
		validator: validator,
	}
}

func (s ProjectService) CreateProject(ctx context.Context, in interfaces2.CreateProjectInput) (*types.ProjectEntity, error) {
	if err := s.validator.ValidateCreateProject(ctx, in); err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	p := &types.ProjectEntity{
		ID:                 uuid.NewString(),
		Name:               stringsTrim(in.Name),
		Slug:               stringsTrim(in.Slug),
		Description:        stringsTrimPtr(in.Description),
		DesignReferenceURL: stringsTrimPtr(in.DesignReferenceURL),
		Status:             types.ProjectStatusActive,
		CreatedAt:          now,
		UpdatedAt:          now,
	}

	if err := s.projects.Create(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s ProjectService) GetProject(ctx context.Context, id string) (*types.ProjectEntity, error) {
	return s.projects.FindByID(ctx, id)
}

func (s ProjectService) ListProjects(ctx context.Context) ([]*types.ProjectEntity, error) {
	return s.projects.List(ctx)
}

func (s ProjectService) UpdateProject(ctx context.Context, in interfaces2.UpdateProjectInput) (*types.ProjectEntity, error) {
	if err := s.validator.ValidateUpdateProject(ctx, in); err != nil {
		return nil, err
	}

	p, err := s.projects.FindByID(ctx, in.ID)
	if err != nil {
		return nil, err
	}

	if in.Name != nil {
		p.Name = stringsTrim(*in.Name)
	}
	if in.Slug != nil {
		p.Slug = stringsTrim(*in.Slug)
	}
	if in.Description != nil {
		p.Description = stringsTrimPtr(*in.Description)
	}
	if in.DesignReferenceURL != nil {
		p.DesignReferenceURL = stringsTrimPtr(*in.DesignReferenceURL)
	}
	if in.Archive != nil {
		if *in.Archive {
			p.Status = types.ProjectStatusArchived
			now := time.Now().UTC()
			p.ArchivedAt = &now
		} else {
			p.Status = types.ProjectStatusActive
			p.ArchivedAt = nil
		}
	}

	if err := s.projects.Update(ctx, p); err != nil {
		return nil, err
	}
	return p, nil
}

func (s ProjectService) DeleteProject(ctx context.Context, id string) error {
	return s.projects.Delete(ctx, id)
}
