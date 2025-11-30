package github

import (
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/webhookapp/schedule/recovery"
	"io"
	"net/http"
	"strconv"
	"strings"
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

func (c *GitHubClient) doRequestWithRateLimit(method, url string, body io.Reader, token string) (*http.Response, error) {
	resp, err := c.doRequest(method, url, body, token)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		remaining := resp.Header.Get("X-RateLimit-Remaining")
		if remaining == "0" {
			resetHeader := resp.Header.Get("X-RateLimit-Reset")
			if resetHeader != "" {
				unixTime, parseErr := strconv.ParseInt(resetHeader, 10, 64)
				if parseErr == nil {
					resetTime := time.Unix(unixTime, 0)
					sleepDuration := time.Until(resetTime)
					if sleepDuration > 0 {
						fmt.Printf("Rate limit hit. Sleeping until %s (%s)\n",
							resetTime.Format(time.RFC3339), sleepDuration.Round(time.Second))
						resp.Body.Close()
						time.Sleep(sleepDuration)
						return c.doRequest(method, url, body, token)
					}
				}
			}
		}
	}

	return resp, nil
}

func (c *GitHubClient) GetDeliveries(webhookConfig recovery.WebhookConfig, page int, perPage int) ([]recovery.WebhookDelivery, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/hooks/%s/deliveries?per_page=%d&page=%d",
		c.baseURL, webhookConfig.Owner, webhookConfig.Repo, webhookConfig.HookID, perPage, page)

	resp, err := c.doRequest("GET", url, nil, webhookConfig.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch deliveries: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

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
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	if resp.StatusCode != http.StatusAccepted && resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	return nil
}

type PullRequest struct {
	ID           uint64     `json:"id"`
	Number       int32      `json:"number"`
	State        string     `json:"state"`
	Title        string     `json:"title"`
	Body         *string    `json:"body"`
	User         User       `json:"user"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ClosedAt     *time.Time `json:"closed_at"`
	MergedAt     *time.Time `json:"merged_at"`
	Merged       bool       `json:"merged"`
	Mergeable    *bool      `json:"mergeable"`
	Head         GitRef     `json:"head"`
	Base         GitRef     `json:"base"`
	Labels       []Label    `json:"labels"`
	Assignees    []User     `json:"assignees"`
	MergedBy     *User      `json:"merged_by"`
	Additions    int32      `json:"additions"`
	Deletions    int32      `json:"deletions"`
	ChangedFiles int32      `json:"changed_files"`
	Commits      int32      `json:"commits"`
}

type User struct {
	ID    uint64  `json:"id"`
	Login string  `json:"login"`
	Email *string `json:"email"`
}

type GitRef struct {
	Ref  string     `json:"ref"`
	SHA  string     `json:"sha"`
	Repo Repository `json:"repo"`
}

type Repository struct {
	ID       uint64 `json:"id"`
	Name     string `json:"name"`
	FullName string `json:"full_name"`
}

type Label struct {
	Name string `json:"name"`
}

type Review struct {
	ID          uint64    `json:"id"`
	User        User      `json:"user"`
	State       string    `json:"state"`
	SubmittedAt time.Time `json:"submitted_at"`
}

func (c *GitHubClient) ListPullRequests(owner, repo, token string, page, perPage int) ([]*PullRequest, bool, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls?state=all&per_page=%d&page=%d&direction=asc&sort=created",
		c.baseURL, owner, repo, perPage, page)

	resp, err := c.doRequestWithRateLimit("GET", url, nil, token)
	if err != nil {
		return nil, false, fmt.Errorf("failed to fetch PRs: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, false, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var prs []*PullRequest
	if err := json.NewDecoder(resp.Body).Decode(&prs); err != nil {
		return nil, false, fmt.Errorf("failed to decode PRs: %w", err)
	}

	linkHeader := resp.Header.Get("Link")
	hasMore := strings.Contains(linkHeader, `rel="next"`)

	return prs, hasMore, nil
}

func (c *GitHubClient) ListPRReviews(owner, repo string, prNumber int32, token string) ([]*Review, error) {
	url := fmt.Sprintf("%s/repos/%s/%s/pulls/%d/reviews",
		c.baseURL, owner, repo, prNumber)

	resp, err := c.doRequestWithRateLimit("GET", url, nil, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch reviews: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitHub API error: %d - %s", resp.StatusCode, string(body))
	}

	var reviews []*Review
	if err := json.NewDecoder(resp.Body).Decode(&reviews); err != nil {
		return nil, fmt.Errorf("failed to decode reviews: %w", err)
	}

	return reviews, nil
}
