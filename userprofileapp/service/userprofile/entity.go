package userprofile

import (
	types "github.com/gocasters/rankr/type"
	"time"
)

type ContributorInfo struct {
	ID             int64     `json:"id"`
	GitHubID       int64     `json:"github_id"`
	GitHubUsername string    `json:"github_username"`
	DisplayName    *string   `json:"display_name,omitempty"`
	ProfileImage   *string   `json:"profile_image,omitempty"`
	Bio            *string   `json:"bio,omitempty"`
	PrivacyMode    string    `json:"privacy_mode"`
	CreatedAt      time.Time `json:"created_at"`
}

type Task struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	State       string `json:"state"` // open / closed
}

type ContributorStat struct {
	ContributorID types.ID       `json:"contributor_id"`
	GlobalRank    int            `json:"global_rank"`
	TotalScore    float64        `json:"total_score"`
	ProjectsScore map[string]int `json:"project_score"`
	ScoreHistory  map[string]int `json:"score_history"`
}

type Profile struct {
	ContributorInfo ContributorInfo `json:"contributor"`
	Tasks           []Task          `json:"task"`
	ContributorStat ContributorStat `json:"contributor_stat"`
}
