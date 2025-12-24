package contributor

import (
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
	Password       string      `json:"password"`
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

type UpdatePasswordRequest struct {
	ID          types.ID `json:"id"`
	OldPassword string   `json:"old_password"`
	NewPassword string   `json:"new_password"`
}

type UpdatePasswordResponse struct {
	Success bool `json:"success"`
}

type VerifyPasswordRequest struct {
	ID             types.ID `json:"id"`
	GitHubUsername string   `json:"github_username"`
	Password       string   `json:"password"`
}

type VerifyPasswordResponse struct {
	Valid bool     `json:"valid"`
	ID    types.ID `json:"contributor_id"`
}

type VcsProvider string

const (
	VcsProviderGitHub    VcsProvider = "GITHUB"
	VcsProviderGitLab    VcsProvider = "GITLAB"
	VcsProviderBitbucket VcsProvider = "BITBUCKET"
)

var validVcsProviders = map[VcsProvider]struct{}{
	VcsProviderGitHub:    {},
	VcsProviderGitLab:    {},
	VcsProviderBitbucket: {},
}

func IsValidVcsProvider(p string) bool {
	_, ok := validVcsProviders[VcsProvider(p)]
	return ok
}

type GetContributorsByVCSRequest struct {
	VcsProvider VcsProvider `json:"vcs_provider"`
	Usernames   []string    `json:"usernames"`
}

type ContributorMapping struct {
	ContributorID int64  `json:"contributor_id"`
	VcsUsername   string `json:"vcs_username"`
	VcsUserID     int64  `json:"vcs_user_id"`
}

type GetContributorsByVCSResponse struct {
	VcsProvider  VcsProvider          `json:"vcs_provider"`
	Contributors []ContributorMapping `json:"contributors"`
}
