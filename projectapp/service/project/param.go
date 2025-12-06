package project

import (
	"time"

	"github.com/gocasters/rankr/projectapp/constant"
)

type CreateProjectInput struct {
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Description        *string                `json:"description,omitempty"`
	DesignReferenceURL *string                `json:"designReferenceUrl,omitempty"`
	GitRepoID          *string                `json:"gitRepoId,omitempty"`
	RepoProvider       *constant.VcsProvider  `json:"repoProvider,omitempty"`
	Status             constant.ProjectStatus `json:"status"`
}

type CreateProjectResponse struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Description        *string                `json:"description,omitempty"`
	DesignReferenceURL *string                `json:"designReferenceUrl,omitempty"`
	GitRepoID          *string                `json:"gitRepoId,omitempty"`
	RepoProvider       *constant.VcsProvider  `json:"repoProvider,omitempty"`
	Status             constant.ProjectStatus `json:"status"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
}

type UpdateProjectInput struct {
	ID                 string                  `json:"id"`
	Name               *string                 `json:"name,omitempty"`
	Slug               *string                 `json:"slug,omitempty"`
	Description        *string                 `json:"description,omitempty"`
	DesignReferenceURL *string                 `json:"designReferenceUrl,omitempty"`
	GitRepoID          *string                 `json:"gitRepoId,omitempty"`
	RepoProvider       *constant.VcsProvider   `json:"repoProvider,omitempty"`
	Status             *constant.ProjectStatus `json:"status,omitempty"`
}

type UpdateProjectResponse struct {
	ID string `json:"id"`
}

type GetProjectByIDInput struct {
	ID string `json:"id"`
}

type GetProjectByIDResponse struct {
	ID                 string                 `json:"id"`
	Name               string                 `json:"name"`
	Slug               string                 `json:"slug"`
	Description        *string                `json:"description,omitempty"`
	DesignReferenceURL *string                `json:"designReferenceUrl,omitempty"`
	GitRepoID          *string                `json:"gitRepoId,omitempty"`
	RepoProvider       *constant.VcsProvider  `json:"repoProvider,omitempty"`
	Status             constant.ProjectStatus `json:"status"`
	CreatedAt          time.Time              `json:"createdAt"`
	UpdatedAt          time.Time              `json:"updatedAt"`
	ArchivedAt         *time.Time             `json:"archivedAt,omitempty"`
}

type ListProjectsInput struct{}

type ListProjectsResponse struct {
	Projects []GetProjectByIDResponse `json:"projects"`
}

type DeleteProjectInput struct {
	ID string `json:"id"`
}

type DeleteProjectResponse struct {
	ID string `json:"id"`
}

type GetProjectByVCSRepoRequest struct {
	Provider constant.VcsProvider `json:"provider"`
	RepoID   string               `json:"repo_id"`
}

type GetProjectByVCSRepoResponse struct {
	ID           string                `json:"id"`
	Slug         string                `json:"slug"`
	Name         string                `json:"name"`
	RepoProvider *constant.VcsProvider `json:"repo_provider,omitempty"`
	GitRepoID    *string               `json:"git_repo_id,omitempty"`
}
