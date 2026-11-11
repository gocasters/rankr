package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLeaderboardCommandRejectsInvalidAPIURL(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")
	content := []byte(`
http_server:
  host: "127.0.0.1"
  port: 8085
leaderboard_api:
  base_url: "not-a-url"
  timeout: 1s
`)

	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("CONFIG_PATH", configPath)

	root := newRootCommand()
	root.SetArgs([]string{"serve", "--leaderboard"})
	root.SilenceUsage = true
	root.SilenceErrors = true

	err := root.ExecuteContext(t.Context())
	if err == nil ||
		!strings.Contains(err.Error(), "absolute HTTP or HTTPS URL") {
		t.Fatalf("expected invalid leaderboard API URL error, got: %v", err)
	}
}
