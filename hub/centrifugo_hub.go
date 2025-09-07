package hub

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

type Hub interface {
	Publish(ctx context.Context, channel string, data any) error
}

type CentrifugoHub struct {
	ApiURL string
	ApiKey string
	Client *http.Client
}

func NewCentrifugoHub(apiURL, apiKey string) *CentrifugoHub {
	return &CentrifugoHub{
		ApiURL: strings.TrimSuffix(apiURL, "/"),
		ApiKey: apiKey,
		Client: &http.Client{},
	}
}

func (c *CentrifugoHub) Publish(ctx context.Context, channel string, data any) error {
	body, _ := json.Marshal(map[string]any{
		"method": "publish",
		"params": map[string]any{
			"channel": channel,
			"data":    data,
		},
	})

	req, _ := http.NewRequestWithContext(ctx, "POST", c.ApiURL+"/api", strings.NewReader(string(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "apikey "+c.ApiKey)

	resp, err := c.Client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("centrifugo error: %s", resp.Status)
	}
	return nil
}
