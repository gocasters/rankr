package task

import (
	"time"
)

type Comment struct {
	ID        int64     `json:"id"`
	NodeID    string    `json:"node_id"`
	User      int64     `json:"user"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
