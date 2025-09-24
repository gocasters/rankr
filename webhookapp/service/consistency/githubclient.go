package consistency

import (
	"context"
	"encoding/json"
	"github.com/gocasters/rankr/webhookapp/service"
	"strings"

	//"encoding/json"
	"fmt"
	//"log/slog"
	"net/http"
	"time"
)

// GitHubClient wraps GitHub API interactions
type GitHubClient struct {
	httpClient *http.Client
	token      string
	baseURL    string
}

// NewGitHubClient creates a new GitHub API client
func NewGitHubClient(token string) *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		token:      token,
		baseURL:    "https://api.github.com",
	}
}

// GetRepositoryWebhookEvents GetRepositoryEvents fetches events for a repository from GitHub API
func (c *GitHubClient) GetRepositoryWebhookEvents(ctx context.Context, owner, repoName string, hookID int64, page, perPage int) ([]service.DeliveryEvent, bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d/deliveries?page=%d&per_page=%d", c.baseURL, owner, repoName, hookID, page, perPage)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, false, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	var events []service.DeliveryEvent
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		return nil, false, fmt.Errorf("failed to decode response: %w", err)
	}

	//Check if there are more pages
	linkHeader := resp.Header.Get("Link")
	hasMore := strings.Contains(linkHeader, "rel=\"next\"")

	return events, hasMore, nil
}

func (c *GitHubClient) RedeliverLostEvent(ctx context.Context, owner, repoName string, hookID int64, deliveryID string) error {
	///repos/{owner}/{repo}/hooks/{hook_id}/deliveries/{delivery_id}/attempts
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%d/deliveries/%s/attempts", c.baseURL, owner, repoName, hookID, deliveryID)

	req, err := http.NewRequestWithContext(ctx, "POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.token)
	req.Header.Set("Accept", "application/vnd.github+json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("GitHub API returned status %d", resp.StatusCode)
	}

	return nil
}
