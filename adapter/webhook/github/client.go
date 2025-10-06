package github

import (
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/webhookapp/schedule/recovery"
	"io"
	"net/http"
	"time"
)

type GitHubClient struct {
	httpClient *http.Client
	baseURL    string
}

func NewGitHubClient() *GitHubClient {
	return &GitHubClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://api.github.com",
	}
}

func (c *GitHubClient) doRequest(method, url string, body io.Reader, token string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	return c.httpClient.Do(req)
}

func (c *GitHubClient) GetDeliveries(webhookConfig recovery.WebhookConfig, page int, perPage int) ([]recovery.WebhookDelivery, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%s/deliveries?per_page=%d&page=%d",
		c.baseURL, webhookConfig.Owner, webhookConfig.Repo, webhookConfig.HookID, perPage, page)

	resp, err := c.doRequest("GET", url, nil, webhookConfig.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deliveries: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var deliveries []recovery.WebhookDelivery
	if err := json.NewDecoder(resp.Body).Decode(&deliveries); err != nil {
		return nil, fmt.Errorf("failed to decode deliveries: %w", err)
	}

	return deliveries, nil

}

func (c *GitHubClient) ReattemptDelivery(webhookConfig recovery.WebhookConfig, deliveryID string) error {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%s/deliveries/%s/attempts",
		c.baseURL, webhookConfig.Owner, webhookConfig.Repo, webhookConfig.HookID, deliveryID)

	resp, err := c.doRequest("POST", url, nil, webhookConfig.Token)
	if err != nil {
		return fmt.Errorf("failed to reattempt delivery: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
