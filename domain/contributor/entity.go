package contributor

import (
    "time"
    "github.com/gocasters/rankr/pkg/uuid"
)

// Contributor represents a user in the system
type Contributor struct {
    ID          string    `json:"id"`
    Username    string    `json:"username"`
    Email       string    `json:"email"`
    DisplayName string    `json:"display_name"`
    AvatarURL   string    `json:"avatar_url,omitempty"`
    GitHubID    string    `json:"github_id,omitempty"`
    IsActive    bool      `json:"is_active"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}

// ContributorCreate represents the data needed to create a new contributor
type ContributorCreate struct {
    Username    string `json:"username"`
    Email       string `json:"email"`
    DisplayName string `json:"display_name"`
    AvatarURL   string `json:"avatar_url,omitempty"`
    GitHubID    string `json:"github_id,omitempty"`
}

// ContributorUpdate represents the data needed to update a contributor
type ContributorUpdate struct {
    Username    string `json:"username,omitempty"`
    Email       string `json:"email,omitempty"`
    DisplayName string `json:"display_name,omitempty"`
    AvatarURL   string `json:"avatar_url,omitempty"`
    GitHubID    string `json:"github_id,omitempty"`
    IsActive    *bool  `json:"is_active,omitempty"`
}

// NewContributor creates a new contributor instance
func NewContributor(username, email, displayName string) *Contributor {
    now := time.Now()
    return &Contributor{
        ID:          generateID(),
        Username:    username,
        Email:       email,
        DisplayName: displayName,
        IsActive:    true,
        CreatedAt:   now,
        UpdatedAt:   now,
    }
}

// generateID generates a unique ID for the contributor
func generateID() string {
    return uuid.Generate()
}
