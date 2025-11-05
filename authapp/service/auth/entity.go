package auth

import (
	"time"

	types "github.com/gocasters/rankr/type"
)

type Grant struct {
	ID        types.ID  `json:"id"`
	Subject   string    `json:"subject"`
	Object    string    `json:"object"`
	Action    string    `json:"action"`
	Field     []string  `json:"field,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
