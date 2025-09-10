package versioncontrollersystemproject

import (
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
)

type VersionControllerSystemProjectEntity struct {
	ID             string                 `db:"id" json:"id"`
	ProjectID      string                 `db:"project_id" json:"projectId"`
	Provider       constant.VcsProvider   `db:"provider" json:"provider"`
	ProviderRepoID string                 `db:"provider_repo_id" json:"providerRepoId"`
	Owner          string                 `db:"owner" json:"owner"`
	Name           string                 `db:"name" json:"name"`
	RemoteURL      string                 `db:"remote_url" json:"remoteUrl"`
	DefaultBranch  *string                `db:"default_branch" json:"defaultBranch,omitempty"`
	Visibility     constant.VcsVisibility `db:"visibility" json:"visibility"`
	InstallationID *string                `db:"installation_id" json:"installationId,omitempty"`
	LastSyncedAt   *time.Time             `db:"last_synced_at" json:"lastSyncedAt,omitempty"`
	CreatedAt      time.Time              `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time              `db:"updated_at" json:"updatedAt"`
}
