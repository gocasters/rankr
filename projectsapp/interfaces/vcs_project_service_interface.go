package interfaces

import (
	"context"
	"time"

	"github.com/gocasters/rankr/projectsapp/types"
)

type IVcsRepoService interface {
	CreateVcsRepo(ctx context.Context, in CreateVcsRepoInput) (*types.VcsRepoEntity, error)

	GetVcsRepo(ctx context.Context, id string) (*types.VcsRepoEntity, error)

	GetVcsReposByProject(ctx context.Context, projectID string) ([]*types.VcsRepoEntity, error)

	GetVcsRepoByProviderID(ctx context.Context, provider types.VcsProvider, providerRepoID string, projectID string) (*types.VcsRepoEntity, error)

	UpdateVcsRepo(ctx context.Context, in UpdateVcsRepoInput) (*types.VcsRepoEntity, error)

	DeleteVcsRepo(ctx context.Context, id string) error

	ListVcsRepo(ctx context.Context) ([]*types.VcsRepoEntity, error)
}

type CreateVcsRepoInput struct {
	ProjectID      string
	Provider       types.VcsProvider
	ProviderRepoID string
	Owner          string
	Name           string
	RemoteURL      string
	DefaultBranch  *string
	Visibility     types.VcsVisibility
	InstallationID *string
}

type UpdateVcsRepoInput struct {
	ID             string
	Owner          *string
	Name           *string
	RemoteURL      *string
	DefaultBranch  **string
	Visibility     *types.VcsVisibility
	InstallationID **string
	LastSyncedAt   **time.Time
}
