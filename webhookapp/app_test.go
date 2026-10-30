package webhookapp

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"testing"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/httpserver"
)

func TestSetupRejectsEmptyGitHubSecret(t *testing.T) {
	config := Config{HTTPServer: httpserver.Config{Port: 8080}}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("Setup() did not panic for an empty GitHub webhook secret")
		}
		if got := fmt.Sprint(recovered); !strings.Contains(got, "GitHub webhook secret") {
			t.Fatalf("Setup() panic = %q, want missing GitHub webhook secret", got)
		}
	}()

	Setup(config, &database.Database{}, nil, &redis.Adapter{})
}

func TestSetupRejectsNegativeGitHubPayloadLimit(t *testing.T) {
	config := Config{
		HTTPServer: httpserver.Config{Port: 8080},
		GithubConfig: GithubConfig{
			Secret:           "test-secret",
			MaxPayloadSizeMB: -1,
		},
	}

	defer func() {
		recovered := recover()
		if recovered == nil {
			t.Fatal("Setup() did not panic for a negative GitHub payload limit")
		}
		const want = "the GitHub webhook max payload size must not be negative"
		if got := fmt.Sprint(recovered); got != want {
			t.Fatalf("Setup() panic = %q, want %q", got, want)
		}
	}()

	Setup(config, &database.Database{}, nil, &redis.Adapter{})
}

func TestSetupConvertsGitHubPayloadLimitFromMiBToBytes(t *testing.T) {
	secret := []byte("test-secret")
	config := Config{
		HTTPServer: httpserver.Config{Port: 8080},
		GithubConfig: GithubConfig{
			Secret:           string(secret),
			MaxPayloadSizeMB: 1,
		},
	}
	app := Setup(config, &database.Database{}, nil, &redis.Adapter{})

	t.Run("accepts exactly one MiB", func(t *testing.T) {
		payload := []byte(strings.Repeat("a", 1<<20))
		_, err := app.HTTPServer.Service.VerifyEvent(strings.NewReader(string(payload)), "12345", "push", "delivery-123", appTestSignature(secret, payload))
		if err != nil {
			t.Fatalf("VerifyEvent() error = %v for a 1 MiB payload", err)
		}
	})

	t.Run("rejects more than one MiB", func(t *testing.T) {
		payload := []byte(strings.Repeat("a", (1<<20)+1))
		_, err := app.HTTPServer.Service.VerifyEvent(strings.NewReader(string(payload)), "12345", "push", "delivery-123", appTestSignature(secret, payload))
		if err == nil || !strings.Contains(err.Error(), "payload") {
			t.Fatalf("VerifyEvent() error = %v, want payload limit error", err)
		}
	})
}

func appTestSignature(secret, payload []byte) string {
	mac := hmac.New(sha256.New, secret)
	_, _ = mac.Write(payload)
	return "sha256=" + hex.EncodeToString(mac.Sum(nil))
}
