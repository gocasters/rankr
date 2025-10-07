package contributor

import (
	"time"

	types "github.com/gocasters/rankr/type"
)

type GetProfileRequest struct {
	ID types.ID `json:"id"`
}

type GetProfileResponse struct {
	ID             int64       `json:"id"`
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username"`
	DisplayName    *string     `json:"display_name,omitempty"`
	ProfileImage   *string     `json:"profile_image,omitempty"`
	Bio            *string     `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode"`
	CreatedAt      time.Time   `json:"created_at"`
}

type CreateContributorRequest struct {
	GitHubID       int64       `json:"github_id"`
	GitHubUsername string      `json:"github_username"`
	DisplayName    *string     `json:"display_name,omitempty"`
	ProfileImage   *string     `json:"profile_image,omitempty"`
	Bio            *string     `json:"bio,omitempty"`
	PrivacyMode    PrivacyMode `json:"privacy_mode"`
}

type CreateContributorResponse struct {
	ID types.ID `json:"id"`
}

type UpdateProfileRequest struct {
	ID           types.ID    `json:"id"`
	DisplayName  *string     `json:"display_name,omitempty"`
	ProfileImage *string     `json:"profile_image,omitempty"`
	Bio          *string     `json:"bio,omitempty"`
	PrivacyMode  PrivacyMode `json:"privacy_mode,omitempty"`
}

type UpdateProfileResponse struct {
	ID             int64      `json:"id"`
	GitHubID       int64      `json:"github_id"`
	GitHubUsername string     `json:"github_username"`
	DisplayName    *string    `json:"display_name"`
	ProfileImage   *string    `json:"profile_image"`
	Bio            *string    `json:"bio"`
	PrivacyMode    string     `json:"privacy_mode"`
	UpdatedAt      *time.Time `json:"updated_at"`
}
