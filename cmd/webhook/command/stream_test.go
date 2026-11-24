package command

import (
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/config"
	"github.com/gocasters/rankr/webhookapp"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/require"
)

func TestEnsureWebhookStreamCreatesWithMaxBytes(t *testing.T) {
	// Start a temporary NATS server with a random port and a test-specific data directory.
	srv, err := server.NewServer(&server.Options{
		Host:      "127.0.0.1",
		Port:      -1,
		JetStream: true,
		StoreDir:  t.TempDir(),
	})
	require.NoError(t, err)

	go srv.Start()
	t.Cleanup(func() {
		srv.Shutdown()
		srv.WaitForShutdown()
	})
	require.True(t, srv.ReadyForConnections(5*time.Second))

	conn, err := nats.Connect(srv.ClientURL())
	require.NoError(t, err)
	t.Cleanup(conn.Close)

	js, err := conn.JetStream()
	require.NoError(t, err)

	// Use the same configuration the webhook passes at startup.
	var cfg webhookapp.NATSConfig
	cfg.Stream.Name = "EVENTS_STREAM"
	cfg.Stream.Subjects = []string{"rankr_raw_events"}
	cfg.Stream.MaxBytes = 256 * 1024 * 1024

	require.NoError(t, ensureWebhookStream(js, cfg))

	// Read the stream configuration back from JetStream.
	info, err := js.StreamInfo(cfg.Stream.Name)
	require.NoError(t, err)
	require.Equal(t, cfg.Stream.MaxBytes, info.Config.MaxBytes)

	t.Run("updates an existing stream and keeps its message", func(t *testing.T) {
		// Simulate an existing Stage stream with no byte limit.
		oldConfig := info.Config
		oldConfig.MaxBytes = -1
		_, err := js.UpdateStream(&oldConfig)
		require.NoError(t, err)

		_, err = js.Publish("rankr_raw_events", []byte("event before restart"))
		require.NoError(t, err)

		// Simulate a webhook restart with the configured byte limit.
		require.NoError(t, ensureWebhookStream(js, cfg))

		after, err := js.StreamInfo(cfg.Stream.Name)
		require.NoError(t, err)
		require.Equal(t, cfg.Stream.MaxBytes, after.Config.MaxBytes)
		require.EqualValues(t, 1, after.State.Msgs)
	})

	t.Run("rejects a lower limit than existing stored data", func(t *testing.T) {
		_, err := js.Publish("rankr_raw_events", []byte("existing event data"))
		require.NoError(t, err)

		_, err = js.Publish("rankr_raw_events", []byte("another event"))
		require.NoError(t, err)

		originalInfo, err := js.StreamInfo(cfg.Stream.Name)
		require.NoError(t, err)

		lowerLimit := int64(1)
		cfgWithLowerLimit := cfg
		cfgWithLowerLimit.Stream.MaxBytes = lowerLimit

		// Ensure the existing data exceeds the requested limit.
		require.Greater(t, originalInfo.State.Bytes, uint64(1))

		err = ensureWebhookStream(js, cfgWithLowerLimit)
		require.Error(t, err)
		require.Contains(t, err.Error(), "stored bytes exceed")

		after, err := js.StreamInfo(cfg.Stream.Name)
		require.NoError(t, err)
		require.Equal(t, originalInfo.Config.MaxBytes, after.Config.MaxBytes)
		require.Equal(t, originalInfo.State.Msgs, after.State.Msgs)
		require.Equal(t, originalInfo.State.Bytes, after.State.Bytes)
	})
}

func TestWebhookStreamLimitDefaults(t *testing.T) {
	tests := []struct {
		name string
		path string
		want int64
	}{
		{
			name: "local",
			path: "../../../deploy/webhook/development/config.local.yml",
			want: 64 * 1024 * 1024,
		},
		{
			name: "development",
			path: "../../../deploy/webhook/development/config.yml",
			want: 64 * 1024 * 1024,
		},
		{
			name: "production",
			path: "../../../deploy/webhook/production/config.yml",
			want: 1024 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var cfg webhookapp.Config
			err := config.Load(config.Options{
				// Use a separate prefix to test YAML defaults in isolation.
				Prefix:       "RANKR_YAML_DEFAULTS_TEST_",
				YamlFilePath: tt.path,
			}, &cfg)
			if err != nil {
				t.Fatal(err)
			}

			if got := cfg.NATSConfig.Stream.MaxBytes; got != tt.want {
				t.Fatalf("stream MaxBytes = %d, want %d", got, tt.want)
			}
		})
	}
}
