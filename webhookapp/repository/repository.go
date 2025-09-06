package repository

import (
	"context"
	"encoding/json"
	"fmt"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/jackc/pgx/v5/pgxpool"
	"time"
)

type WebhookRepository struct {
	db *pgxpool.Pool
}

func NewWebhookRepository(db *pgxpool.Pool) WebhookRepository {
	return WebhookRepository{
		db: db,
	}
}

func (repo WebhookRepository) Save(ctx context.Context, event *eventpb.Event) error {
	payload, err := json.Marshal(event.Payload)
	if err != nil {
		return err
	}

	tag, exErr := repo.db.Exec(
		ctx,
		`
		INSERT INTO webhook_events (provider, delivery_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, delivery_id) DO NOTHING
	`,
		event.Provider,
		event.Id,
		event.EventName,
		payload,
		time.Now(),
	)

	if exErr != nil {
		return exErr
	}

	if tag.RowsAffected() == 0 {
		return fmt.Errorf("no rows affected")
	}

	return nil
}
