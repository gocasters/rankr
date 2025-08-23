package contributor

import (
	"github.com/gocasters/rankr/types",
	"time"
)


type Contributor struct {
    ID           string    `json:"id" db:"id"`
    Username     string    `json:"username" db:"username"`
    GitHubID     int64     `json:"github_id" db:"github_id"`
    AvatarURL    string    `json:"avatar_url" db:"avatar_url"`
    Score        int       `json:"score" db:"score"`
    Rank         int       `json:"rank" db:"rank"`
    Contributions int      `json:"contributions" db:"contributions"`
    LastUpdated  time.Time `json:"last_updated" db:"last_updated"`
    CreatedAt    time.Time `json:"created_at" db:"created_at"`
}