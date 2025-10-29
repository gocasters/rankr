package task

import (
	"time"

	types "github.com/gocasters/rankr/type"
)

type Task struct {
	ID                     types.ID   `json:"id"`
	VersionControlSystemId int64      `json:"version_control_system_id"`
	IssueNumber            int        `json:"issue_number"`
	Title                  string     `json:"title"`
	Description            string     `json:"description"`
	State                  string     `json:"state"`
	RepositoryName         string     `json:"repository_name"`
	Author                 string     `json:"author"`
	Assignees              []int64    `json:"assignees"`
	Labels                 []Label    `json:"labels"`
	Milestone              Milestone  `json:"milestone"`
	CreatedAt              time.Time  `json:"created_at"`
	UpdatedAt              time.Time  `json:"updated_at"`
	ClosedAt               *time.Time `json:"closed_at,omitempty"`
	ClosedBy               string     `json:"closed_by"`
	DeleteAt               bool       `json:"delete_at"`
	Locked                 bool       `json:"locked"`
	Comments               []Comment  `json:"comments"`
	LinkedPR               *int64     `json:"linked_pr,omitempty"`
}
