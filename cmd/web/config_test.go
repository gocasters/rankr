package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLoadConfigReadsYAMLAndEnv(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.yml")

	content := []byte(`
http_server:
  host: "127.0.0.1"
  port: 8085
  shutdown_context_timeout: 3s
`)

	if err := os.WriteFile(configPath, content, 0600); err != nil {
		t.Fatalf("write test config: %v", err)
	}

	t.Setenv("WEB__HTTP_SERVER__PORT", "9090")

	cfg, err := loadConfig(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	if cfg.HTTPServer.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %q", cfg.HTTPServer.Host)
	}

	if cfg.HTTPServer.Port != 9090 {
		t.Errorf("expected environment override port 9090, got %d",
			cfg.HTTPServer.Port)
	}

	if cfg.HTTPServer.ShutdownTimeout != 3*time.Second {
		t.Errorf("expected shutdown timeout 3s, got %v",
			cfg.HTTPServer.ShutdownTimeout)
	}
}
