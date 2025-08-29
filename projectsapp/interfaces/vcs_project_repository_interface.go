package interfaces

import (
	"context"

	"github.com/gocasters/rankr/projectsapp/types"
)

type VcsRepoRepository interface {
	Create(ctx context.Context, repo *types.VcsRepoEntity) error

	FindByID(ctx context.Context, id string) (*types.VcsRepoEntity, error)

	FindByProjectID(ctx context.Context, projectID string) ([]*types.VcsRepoEntity, error)

	FindByProviderID(ctx context.Context, provider types.VcsProvider, providerRepoID string, projectID string) (*types.VcsRepoEntity, error)

	Update(ctx context.Context, repo *types.VcsRepoEntity) error

	Delete(ctx context.Context, id string) error

	List(ctx context.Context) ([]*types.VcsRepoEntity, error)
}
