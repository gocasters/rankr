package contributor

import (
	"github.com/gocasters/rankr/type"
	"mime/multipart"
	"time"
)

type GetProfileResponse struct {
	ID             int64       `json:"id"`
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username"`
	DisplayName    string      `json:"display_name,omitempty"`
	ProfileImage   string      `json:"profile_image,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode"`
	CreatedAt      time.Time   `json:"created_at"`
}

type CreateContributorRequest struct {
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username"`
	DisplayName    string      `json:"display_name,omitempty"`
	ProfileImage   string      `json:"profile_image,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode"`
}

type CreateContributorResponse struct {
	ID types.ID `json:"id"`
}

type UpdateProfileRequest struct {
	ID             types.ID    `json:"id"`
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username,omitempty"`
	DisplayName    string      `json:"display_name,omitempty"`
	ProfileImage   string      `json:"profile_image,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode,omitempty"`
}

type UpdateProfileResponse struct {
	ID             int64       `json:"id"`
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username"`
	DisplayName    string      `json:"display_name,omitempty"`
	ProfileImage   string      `json:"profile_image,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode"`
	CreatedAt      time.Time   `json:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"`
}

type UpsertContributorRequest struct {
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username,omitempty"`
	DisplayName    string      `json:"display_name,omitempty"`
	ProfileImage   string      `json:"profile_image,omitempty"`
	Bio            string      `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode,omitempty"`
}

type UpsertContributorResponse struct {
	ID    types.ID
	IsNew bool
}

type ImportContributorRequest struct {
	File           multipart.File `json:"file"`
	FileName       string         `json:"file_name"`
	FileType       string         `json:"file_type"`
	IdempotencyKey string         `json:"idempotency_key"`
}

type ImportContributorResponse struct {
	JobID   uint   `json:"job_id"`
	Message string `json:"message"`
}
