package interfaces

import (
	"context"

	"github.com/gocasters/rankr/projectsapp/types"
)

type IProjectRepository interface {
	Create(ctx context.Context, project *types.ProjectEntity) error

	FindByID(ctx context.Context, id string) (*types.ProjectEntity, error)

	FindBySlug(ctx context.Context, slug string) (*types.ProjectEntity, error)

	List(ctx context.Context) ([]*types.ProjectEntity, error)

	Update(ctx context.Context, project *types.ProjectEntity) error

	Delete(ctx context.Context, id string) error
}
