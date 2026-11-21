package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestServeWebFlags(t *testing.T) {
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

			root := newRootCommand()
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
