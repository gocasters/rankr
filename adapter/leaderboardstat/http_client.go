package leaderboardstat

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
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

type HTTPClient struct {
	baseURL *url.URL
	client  *http.Client
}

func NewHTTPClient(baseURL string, timeout time.Duration) (*HTTPClient, error) {

	if timeout <= 0 {
		return nil, fmt.Errorf("leaderboardstat: timeout must be positive")
	}

	base, err := url.Parse(baseURL)

	if err != nil {
		return nil, fmt.Errorf("leaderboardstat: parse base URL: %w", err)
	}

	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, fmt.Errorf("leaderboardstat: base URL must be an absolute HTTP or HTTPS URL")
	}

	return &HTTPClient{
		baseURL: base,
		client:  &http.Client{Timeout: timeout},
	}, nil
}

func (c *HTTPClient) GetPublicLeaderboard(
	ctx context.Context,
	repositoryID uint64,
	pageSize, offset int32,
) (PublicLeaderboard, error) {
	endpoint := c.baseURL.JoinPath(
		"v1", "leaderboard", "public",
		strconv.FormatUint(repositoryID, 10),
	)

	query := endpoint.Query()
	query.Set("page_size", strconv.FormatInt(int64(pageSize), 10))
	query.Set("offset", strconv.FormatInt(int64(offset), 10))

	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, endpoint.String(), nil,
	)

	if err != nil {
		return PublicLeaderboard{}, fmt.Errorf("leaderboardstat: create request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	response, err := c.client.Do(request)

	if err != nil {
		return PublicLeaderboard{}, fmt.Errorf("leaderboardstat: request leaderboard: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return PublicLeaderboard{}, fmt.Errorf(
			"leaderboardstat: unexpected HTTP status: %d",
			response.StatusCode,
		)
	}
	var result PublicLeaderboard
	body := io.LimitReader(response.Body, 1<<20)

	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return PublicLeaderboard{}, fmt.Errorf("leaderboardstat: decode leaderboard: %w", err)
	}

	return result, nil
}
