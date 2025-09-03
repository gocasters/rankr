package task

import "time"

type Milestone struct {
	ID           int64     `json:"id"`
	NodeID       string    `json:"node_id"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Description  string    `json:"description"`
	State        string    `json:"state"`
	OpenIssues   int       `json:"open_issues"`
	ClosedIssues int       `json:"closed_issues"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	DueOn        time.Time `json:"due_on"`
	ClosedAt     time.Time `json:"closed_at"`
}
