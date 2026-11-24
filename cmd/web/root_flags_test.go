package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gocasters/rankr/pkg/httpserver"
)

func clearWebTestEnvironment(t *testing.T) {
	t.Helper()
	for _, entry := range os.Environ() {
		key, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(key, "WEB__") {
			continue
		}
		t.Setenv(key, "")
		if err := os.Unsetenv(key); err != nil {
			t.Fatal(err)
		}
	}
}

func TestServeWebFlags(t *testing.T) {
	clearWebTestEnvironment(t)

	// A bad port identifies server creation without opening a socket. A bad
	// API URL identifies leaderboard client creation before server creation.
	const config = `
http_server:
  port: 0
leaderboard_api:
  base_url: not-a-url
  timeout: 1s
`
	const invalidPort = "web: create server: invalid port: 0"
	const invalidURL = "web: create leaderboard client:"
	tests := []struct {
		name string
		web  string
		env  map[string]string
		args []string
		want string
	}{
		{name: "both off by default"},
		{name: "leaderboard alone", web: "  leaderboard: true\n"},
		{name: "web only", web: "  enabled: true\n", want: invalidPort},
		{
			name: "web and leaderboard", web: "  enabled: true\n  leaderboard: true\n",
			want: invalidURL,
		},
		{
			name: "environment enables web and leaderboard",
			web:  "  enabled: false\n  leaderboard: false\n",
			env:  map[string]string{"WEB__WEB__ENABLED": "true", "WEB__WEB__LEADERBOARD": "true"},
			want: invalidURL,
		},
		{
			name: "environment disables leaderboard",
			web:  "  enabled: true\n  leaderboard: true\n",
			env:  map[string]string{"WEB__WEB__LEADERBOARD": "false"},
			want: invalidPort,
		},
		{
			name: "explicit web off overrides preview and leaderboard switches",
			web:  "  enabled: true\n  leaderboard: true\n",
			env:  map[string]string{"WEB__WEB__ENABLED": "false"},
			args: []string{"--preview", "--leaderboard"},
		},
		{
			name: "explicit leaderboard switch overrides YAML",
			web:  "  enabled: true\n  leaderboard: true\n",
			args: []string{"--leaderboard=false"},
			want: invalidPort,
		},
		{name: "legacy preview", args: []string{"--preview"}, want: invalidPort},
		{name: "legacy leaderboard", args: []string{"--leaderboard"}, want: invalidURL},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := config
			if tt.web != "" {
				input += "web:\n" + tt.web
			}
			configPath := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(configPath, []byte(input), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CONFIG_PATH", configPath)
			for key, value := range tt.env {
				t.Setenv(key, value)
			}

			root := newRootCommand(serve)
			root.SetArgs(append([]string{"serve"}, tt.args...))
			root.SilenceUsage = true
			root.SilenceErrors = true

			err := root.ExecuteContext(t.Context())
			if tt.want == "" {
				if err != nil {
					t.Fatalf("disabled web should exit successfully, got %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.want) {
				t.Fatalf("expected error containing %q, got %v", tt.want, err)
			}
		})
	}
}

func TestServeLeaderboardRoute(t *testing.T) {
	clearWebTestEnvironment(t)

	for _, tt := range []struct {
		enabled    bool
		wantStatus int
	}{
		{enabled: false, wantStatus: http.StatusNotFound},
		{enabled: true, wantStatus: http.StatusOK},
	} {
		t.Run(fmt.Sprintf("enabled=%t", tt.enabled), func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if !tt.enabled {
					t.Error("disabled leaderboard reached the upstream API")
				}
				_, _ = fmt.Fprint(w, `{"project_id":"5001","rows":[{"rank":1,"user_id":101,"score":12}]}`)
			}))
			defer upstream.Close()

			config := fmt.Sprintf(`
http_server:
  port: 8085
web:
  enabled: true
  leaderboard: %t
leaderboard_api:
  base_url: %s
  timeout: 1s
`, tt.enabled, upstream.URL)
			configPath := filepath.Join(t.TempDir(), "config.yml")
			if err := os.WriteFile(configPath, []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			t.Setenv("CONFIG_PATH", configPath)

			response := httptest.NewRecorder()
			root := newRootCommand(func(_ context.Context, server *httpserver.Server) error {
				request := httptest.NewRequest(http.MethodGet, "/leaderboard?project_id=5001", nil)
				if !tt.enabled {
					// Pass the auth middleware so the test observes the missing route.
					request.Header.Set("X-User-Info", base64.StdEncoding.EncodeToString([]byte(`{"id":42}`)))
				}
				server.GetRouter().ServeHTTP(response, request)
				return nil
			})
			root.SetArgs([]string{"serve"})
			if err := root.ExecuteContext(t.Context()); err != nil {
				t.Fatal(err)
			}
			if response.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d with body %s", tt.wantStatus, response.Code, response.Body)
			}
			if tt.enabled && !strings.Contains(response.Body.String(), "<td>101</td>") {
				t.Errorf("expected leaderboard row in response %s", response.Body)
			}
		})
	}
}
