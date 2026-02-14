package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Config struct {
	Url string `koanf:"url"`
}

type AuthClient struct {
	config Config
	client *http.Client
}

func NewAuthClient(cfg Config) AuthClient {
	return AuthClient{
		config: cfg,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

type GetRoleResponse struct {
	Role struct {
		Name string `json:"name"`
	} `json:"role"`
}

func (a AuthClient) GetRole(ctx context.Context, id string) (GetRoleResponse, error) {
	url := fmt.Sprintf("%s/%s", a.config.Url, id)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return GetRoleResponse{}, err
	}

	res, err := a.client.Do(req)
	if err != nil {
		return GetRoleResponse{}, fmt.Errorf("auth service call failed: %w", err)
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(res.Body)
		return GetRoleResponse{}, fmt.Errorf("auth error: %d %s", res.StatusCode, string(body))
	}

	var getRoleResponse GetRoleResponse
	if err := json.NewDecoder(res.Body).Decode(&getRoleResponse); err != nil {
		return GetRoleResponse{}, err
	}

	return getRoleResponse, nil
}
