package contributor

import (
	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/gocasters/rankr/type"
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

type GetContributorsByVCSRequest struct {
	Provider  constant.VcsProvider `json:"provider"`
	Usernames []string             `json:"usernames"`
}

type ContributorVCSMapping struct {
	ContributorID int64  `json:"contributor_id"`
	VCSUsername   string `json:"vcs_username"`
	VCSUserID     int64  `json:"vcs_user_id"`
}

type GetContributorsByVCSResponse struct {
	Provider     constant.VcsProvider    `json:"provider"`
	Contributors []ContributorVCSMapping `json:"contributors"`
}
