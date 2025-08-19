package interfaces

import (
	"context"

	"github.com/gocasters/rankr/projectsapp/types"
)

type IProjectService interface {
	CreateProject(ctx context.Context, in CreateProjectInput) (*types.ProjectEntity, error)
	GetProject(ctx context.Context, id string) (*types.ProjectEntity, error)
	ListProjects(ctx context.Context) ([]*types.ProjectEntity, error)
	UpdateProject(ctx context.Context, in UpdateProjectInput) (*types.ProjectEntity, error)
	DeleteProject(ctx context.Context, id string) error
}

type CreateProjectInput struct {
	Name               string
	Slug               string
	Description        *string
	DesignReferenceURL *string
}

type UpdateProjectInput struct {
	ID                 string
	Name               *string
	Slug               *string
	Description        **string
	DesignReferenceURL **string
	Archive            *bool
}
