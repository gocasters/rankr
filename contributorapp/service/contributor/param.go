package contributor

import (
	"time"
)

type GetProfileRequest struct {
	ID types.ID `json:"id"`
}

type GetProfileResponse struct {
	ID             int64     `json:"id"`
	GitHubID       int64     `json:"github_id"`
	GitHubUsername string    `json:"github_username"`
	DisplayName    *string   `json:"display_name,omitempty"`
	ProfileImage   *string   `json:"profile_image,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	PrivacyMode    string    `json:"privacy_mode"`
	CreatedAt      time.Time `json:"created_at"`
}
