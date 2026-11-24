package command

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/webhookapp"
)

func TestStageWebhookStreamLimit(t *testing.T) {
	data, err := os.ReadFile("../../../deploy/stage/docker-compose.yml")
	if err != nil {
		t.Fatal(err)
	}

	var compose struct {
		Services map[string]struct {
			Environment map[string]string `yaml:"environment"`
		} `yaml:"services"`
	}
	if err := yaml.Unmarshal(data, &compose); err != nil {
		t.Fatal(err)
	}

	const key = "WEBHOOK_NATS__STREAM__MAX_BYTES"
	const want = "268435456" // 256 MiB

	got, ok := compose.Services["webhook"].Environment[key]
	if !ok {
		t.Fatalf("stage webhook is missing %s", key)
	}
	if got != want {
		t.Fatalf("%s = %q, want %q", key, got, want)
	}
}

func TestWebhookLoadsStreamLimitFromEnv(t *testing.T) {
	t.Setenv("WEBHOOK_NATS__STREAM__MAX_BYTES", "268435456")

	var cfg webhookapp.Config
	err := config.Load(config.Options{
		Prefix:       "WEBHOOK_",
		YamlFilePath: "../../../deploy/webhook/production/config.yml",
	}, &cfg)
	if err != nil {
		t.Fatal(err)
	}

	const want int64 = 256 * 1024 * 1024
	if got := cfg.NATSConfig.Stream.MaxBytes; got != want {
		t.Fatalf("stream MaxBytes = %d, want %d", got, want)
	}
}
