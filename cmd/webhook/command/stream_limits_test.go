package command

import (
	"os"
	"testing"

	"gopkg.in/yaml.v3"
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
