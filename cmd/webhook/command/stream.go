package command

import (
	"errors"
	"fmt"

	"github.com/gocasters/rankr/webhookapp"
	"github.com/nats-io/nats.go"
)

func ensureWebhookStream(js nats.JetStreamContext, cfg webhookapp.NATSConfig) error {
	if cfg.Stream.MaxBytes <= 0 {
		return fmt.Errorf("nats.stream.max_bytes must be positive, got %d", cfg.Stream.MaxBytes)
	}

	info, err := js.StreamInfo(cfg.Stream.Name)
	if errors.Is(err, nats.ErrStreamNotFound) {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     cfg.Stream.Name,
			Subjects: cfg.Stream.Subjects,
			Storage:  nats.FileStorage,
			MaxBytes: cfg.Stream.MaxBytes,
		})
		return err
	}
	if err != nil {
		return fmt.Errorf("inspect stream %q: %w", cfg.Stream.Name, err)
	}

	if info.Config.MaxBytes == cfg.Stream.MaxBytes {
		return nil
	}
	if info.State.Bytes > uint64(cfg.Stream.MaxBytes) {
		return fmt.Errorf("stream %q: %d stored bytes exceed the requested limit of %d",
			cfg.Stream.Name, info.State.Bytes, cfg.Stream.MaxBytes)
	}

	updated := info.Config
	updated.MaxBytes = cfg.Stream.MaxBytes
	_, err = js.UpdateStream(&updated)
	return err
}
