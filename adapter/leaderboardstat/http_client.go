package leaderboardstat

import (
	"context"
	"time"
)

type PublicLeaderboard struct {
	ProjectID string                 `json:"project_id"`
	Rows      []PublicLeaderboardRow `json:"rows"`
}

type PublicLeaderboardRow struct {
	Rank   uint64  `json:"rank"`
	UserID uint64  `json:"user_id"`
	Score  float64 `json:"score"`
}

type HTTPClient struct{}

func NewHTTPClient(baseURL string, timeout time.Duration) (*HTTPClient, error) {
	return &HTTPClient{}, nil
}

func (c *HTTPClient) GetPublicLeaderboard(
	ctx context.Context,
	repositoryID uint64,
	pageSize, offser int64,
) (PublicLeaderboard, error) {
	return PublicLeaderboard{}, nil
}
