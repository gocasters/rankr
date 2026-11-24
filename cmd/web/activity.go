package main

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

type contributorActivity struct {
	Type          string    `json:"type"`
	Repository    string    `json:"repository"`
	SubjectType   string    `json:"subject_type"`
	SubjectNumber int32     `json:"subject_number"`
	Points        int64     `json:"points"`
	OccurredAt    time.Time `json:"time"`
}

type activityClient struct {
	baseURL *url.URL
	client  *http.Client
}

func newActivityClient(baseURL string, timeout time.Duration) (*activityClient, error) {
	if timeout <= 0 {
		return nil, fmt.Errorf("web: activity timeout must be positive")
	}

	base, err := url.Parse(baseURL)
	if err != nil {
		return nil, fmt.Errorf("web: parse activity API URL: %w", err)
	}

	if (base.Scheme != "http" && base.Scheme != "https") || base.Host == "" {
		return nil, fmt.Errorf("web: activity API URL must be an absolute HTTP or HTTPS URL")
	}

	return &activityClient{
		baseURL: base,
		client:  &http.Client{Timeout: timeout},
	}, nil
}

type activityResponse struct {
	Activity []contributorActivity `json:"activity"`
}

func (c *activityClient) GetContributorActivity(
	ctx context.Context,
	userID uint64,
	limit int,
) ([]contributorActivity, error) {
	endpoint := c.baseURL.JoinPath(
		"v1", "contributors", strconv.FormatUint(userID, 10), "activity",
	)

	query := endpoint.Query()
	query.Set("limit", strconv.Itoa(limit))
	endpoint.RawQuery = query.Encode()

	request, err := http.NewRequestWithContext(
		ctx, http.MethodGet, endpoint.String(), nil,
	)
	if err != nil {
		return nil, fmt.Errorf("web: create activity request: %w", err)
	}

	request.Header.Set("Accept", "application/json")
	response, err := c.client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("web: request contributor activity: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"web: activity API returned HTTP status: %d",
			response.StatusCode,
		)
	}

	var result activityResponse
	body := io.LimitReader(response.Body, 1<<20)
	if err := json.NewDecoder(body).Decode(&result); err != nil {
		return nil, fmt.Errorf("web: decode contributor activity: %w", err)
	}

	return result.Activity, nil
}
