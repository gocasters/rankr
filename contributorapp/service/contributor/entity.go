package contributor

import (
	"github.com/gocasters/rankr/type"
)


type Contributor struct {
	ID             int64     `json:"id" db:"id"`
	GitHubID       int64     `json:"github_id" db:"github_id"`
	GitHubUsername *string   `json:"github_username" db:"github_username"`
	Email          *string   `json:"email,omitempty" db:"email"`
	IsVerified     bool      `json:"is_verified" db:"is_verified"`
	TwoFactor      bool      `json:"two_factor_enabled" db:"two_factor_enabled"`
	PrivacyMode    string    `json:"privacy_mode" db:"privacy_mode"`
	DisplayName    *string   `json:"display_name,omitempty" db:"display_name"`
	ProfileImage   *string   `json:"profile_image,omitempty" db:"profile_image"`
	Bio            *string   `json:"bio,omitempty" db:"bio"`
	CreatedAt      time.Time `json:"created_at" db:"created_at"`
	UpdatedAt      time.Time `json:"updated_at" db:"updated_at"`
}
