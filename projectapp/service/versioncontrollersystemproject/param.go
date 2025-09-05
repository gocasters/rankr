package versioncontrollersystemproject

import (
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
)

type CreateVersionControllerSystemProjectInput struct {
	ProjectID      string                 `json:"projectId"`
	Provider       constant.VcsProvider   `json:"provider"`
	ProviderRepoID string                 `json:"providerRepoId"`
	Owner          string                 `json:"owner"`
	Name           string                 `json:"name"`
	RemoteURL      string                 `json:"remoteUrl"`
	DefaultBranch  *string                `json:"defaultBranch,omitempty"`
	Visibility     constant.VcsVisibility `json:"visibility"`
	InstallationID *string                `json:"installationId,omitempty"`
}

type CreateVersionControllerSystemProjectResponse struct {
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"projectId"`
	Provider       constant.VcsProvider   `json:"provider"`
	ProviderRepoID string                 `json:"providerRepoId"`
	Owner          string                 `json:"owner"`
	Name           string                 `json:"name"`
	RemoteURL      string                 `json:"remoteUrl"`
	DefaultBranch  *string                `json:"defaultBranch,omitempty"`
	Visibility     constant.VcsVisibility `json:"visibility"`
	InstallationID *string                `json:"installationId,omitempty"`
	LastSyncedAt   *time.Time             `json:"lastSyncedAt,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

type UpdateVersionControllerSystemProjectInput struct {
	ID             string                  `json:"id"`
	Owner          *string                 `json:"owner,omitempty"`
	Name           *string                 `json:"name,omitempty"`
	RemoteURL      *string                 `json:"remoteUrl,omitempty"`
	DefaultBranch  **string                `json:"defaultBranch,omitempty"`
	Visibility     *constant.VcsVisibility `json:"visibility,omitempty"`
	InstallationID **string                `json:"installationId,omitempty"`
}

type UpdateVersionControllerSystemProjectResponse struct {
	Data *VersionControllerSystemProjectEntity `json:"data"`
}

type GetVersionControllerSystemProjectByIDInput struct {
	ID string `json:"id"`
}

type GetVersionControllerSystemProjectByIDResponse struct {
	ID             string                 `json:"id"`
	ProjectID      string                 `json:"projectId"`
	Provider       constant.VcsProvider   `json:"provider"`
	ProviderRepoID string                 `json:"providerRepoId"`
	Owner          string                 `json:"owner"`
	Name           string                 `json:"name"`
	RemoteURL      string                 `json:"remoteUrl"`
	DefaultBranch  *string                `json:"defaultBranch,omitempty"`
	Visibility     constant.VcsVisibility `json:"visibility"`
	InstallationID *string                `json:"installationId,omitempty"`
	LastSyncedAt   *time.Time             `json:"lastSyncedAt,omitempty"`
	CreatedAt      time.Time              `json:"createdAt"`
	UpdatedAt      time.Time              `json:"updatedAt"`
}

type GetVersionControllerSystemProjectListedResponse struct {
	Items []*VersionControllerSystemProjectEntity `json:"items"`
}

type ListVersionControllerSystemProjectsInput struct{}

type ListVersionControllerSystemProjectsResponse struct {
	VersionControllerSystemProjects []*VersionControllerSystemProjectEntity `json:"versionControllerSystemProjects"`
}

type DeleteVersionControllerSystemProjectInput struct {
	ID string `json:"id"`
}

type DeleteVcsRepoResponse struct {
	ID string `json:"id"`
}
