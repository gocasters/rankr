package types

import "time"

type ProjectStatus string

const (
	ProjectStatusActive   ProjectStatus = "ACTIVE"
	ProjectStatusArchived ProjectStatus = "ARCHIVED"
)

type ProjectEntity struct {
	ID                 string        `db:"id" json:"id"`
	Name               string        `db:"name" json:"name"`
	Slug               string        `db:"slug" json:"slug"`
	Description        *string       `db:"description" json:"description,omitempty"`
	DesignReferenceURL *string       `db:"design_reference_url" json:"designReferenceUrl,omitempty"`
	Status             ProjectStatus `db:"status" json:"status"`
	CreatedAt          time.Time     `db:"created_at" json:"createdAt"`
	UpdatedAt          time.Time     `db:"updated_at" json:"updatedAt"`
	ArchivedAt         *time.Time    `db:"archived_at" json:"archivedAt,omitempty"`
}

type VcsProvider string

const (
	VcsProviderGitHub    VcsProvider = "GITHUB"
	VcsProviderGitLab    VcsProvider = "GITLAB"
	VcsProviderBitbucket VcsProvider = "BITBUCKET"
)

type VcsVisibility string

const (
	VcsVisibilityPublic   VcsVisibility = "PUBLIC"
	VcsVisibilityPrivate  VcsVisibility = "PRIVATE"
	VcsVisibilityInternal VcsVisibility = "INTERNAL"
)

type VcsRepoEntity struct {
	ID             string        `db:"id" json:"id"`
	ProjectID      string        `db:"project_id" json:"projectId"`
	Provider       VcsProvider   `db:"provider" json:"provider"`
	ProviderRepoID string        `db:"provider_repo_id" json:"providerRepoId"`
	Owner          string        `db:"owner" json:"owner"`
	Name           string        `db:"name" json:"name"`
	RemoteURL      string        `db:"remote_url" json:"remoteUrl"`
	DefaultBranch  *string       `db:"default_branch" json:"defaultBranch,omitempty"`
	Visibility     VcsVisibility `db:"visibility" json:"visibility"`
	InstallationID *string       `db:"installation_id" json:"installationId,omitempty"`
	LastSyncedAt   *time.Time    `db:"last_synced_at" json:"lastSyncedAt,omitempty"`
	CreatedAt      time.Time     `db:"created_at" json:"createdAt"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updatedAt"`
}
