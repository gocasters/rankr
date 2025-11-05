package auth

import (
	"time"

	types "github.com/gocasters/rankr/type"
)

type CreateGrantRequest struct {
	Subject string   `json:"subject"`
	Object  string   `json:"object"`
	Action  string   `json:"action"`
	Field   []string `json:"field,omitempty"`
}

type UpdateGrantRequest struct {
	ID      types.ID `json:"id"`
	Subject string   `json:"subject,omitempty"`
	Object  string   `json:"object,omitempty"`
	Action  string   `json:"action,omitempty"`
	Field   []string `json:"field,omitempty"`
}

type GrantResponse struct {
	ID        types.ID  `json:"id"`
	Subject   string    `json:"subject"`
	Object    string    `json:"object"`
	Action    string    `json:"action"`
	Field     []string  `json:"field,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListGrantsResponse struct {
	Grants []GrantResponse `json:"grants"`
	Total  int             `json:"total"`
}
