package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WebhookRepository struct {
	db *pgxpool.Pool
}

type EventFilter struct {
	Provider    *int32
	EventType   *string
	StartTime   *time.Time
	EndTime     *time.Time
	DeliveryIDs []string
	Limit       *int
	Offset      *int
}

type EventStats struct {
	TotalEvents      int64            `json:"total_events"`
	EventsByProvider map[int32]int64  `json:"events_by_provider"`
	EventsByType     map[string]int64 `json:"events_by_type"`
	FirstEventAt     *time.Time       `json:"first_event_at"`
	LastEventAt      *time.Time       `json:"last_event_at"`
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

// FindByDeliveryID retrieves an event by provider and delivery_id
func (repo WebhookRepository) FindByDeliveryID(ctx context.Context, provider int32, deliveryID string) (*eventpb.Event, error) {
	var event eventpb.Event
	var payloadJSON []byte
	var receivedAt time.Time

	err := repo.db.QueryRow(
		ctx,
		`SELECT provider, delivery_id, event_type, payload, received_at 
		 FROM webhook_events 
		 WHERE provider = $1 AND delivery_id = $2`,
		provider, deliveryID,
	).Scan(&event.Provider, &event.Id, &event.EventName, &payloadJSON, &receivedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event with provider %d and delivery_id %s not found",
				provider, deliveryID)
		}
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
		return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
	}

	return &event, nil
}

// GetEventsByTimeRange retrieves events within a time range
func (repo WebhookRepository) GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit int) ([]*eventpb.Event, error) {
	filter := EventFilter{
		StartTime: &start,
		EndTime:   &end,
		Limit:     &limit,
	}
	return repo.FindEvents(ctx, filter)
}

// GetEventsByProvider retrieves events for a specific provider
func (repo WebhookRepository) GetEventsByProvider(ctx context.Context, provider int32, limit int) ([]*eventpb.Event, error) {
	filter := EventFilter{
		Provider: &provider,
		Limit:    &limit,
	}
	return repo.FindEvents(ctx, filter)
}

// GetEventsByType retrieves events of a specific type
func (repo WebhookRepository) GetEventsByType(ctx context.Context, eventType string, limit int) ([]*eventpb.Event, error) {
	filter := EventFilter{
		EventType: &eventType,
		Limit:     &limit,
	}
	return repo.FindEvents(ctx, filter)
}

// GetRecentEvents retrieves the most recent events
func (repo WebhookRepository) GetRecentEvents(ctx context.Context, limit int) ([]*eventpb.Event, error) {
	filter := EventFilter{
		Limit: &limit,
	}
	return repo.FindEvents(ctx, filter)
}

// FindEvents retrieves events based on filter criteria
func (repo WebhookRepository) FindEvents(ctx context.Context, filter EventFilter) ([]*eventpb.Event, error) {
	query := `SELECT provider, delivery_id, event_type, payload, received_at FROM webhook_events WHERE 1=1`
	args := make([]interface{}, 0)
	argCount := 0

	// Build dynamic query based on filter
	if filter.Provider != nil {
		argCount++
		query += fmt.Sprintf(" AND provider = $%d", argCount)
		args = append(args, *filter.Provider)
	}

	if filter.EventType != nil {
		argCount++
		query += fmt.Sprintf(" AND event_type = $%d", argCount)
		args = append(args, *filter.EventType)
	}

	if filter.StartTime != nil {
		argCount++
		query += fmt.Sprintf(" AND received_at >= $%d", argCount)
		args = append(args, *filter.StartTime)
	}

	if filter.EndTime != nil {
		argCount++
		query += fmt.Sprintf(" AND received_at <= $%d", argCount)
		args = append(args, *filter.EndTime)
	}

	if len(filter.DeliveryIDs) > 0 {
		argCount++
		query += fmt.Sprintf(" AND delivery_id = ANY($%d)", argCount)
		args = append(args, filter.DeliveryIDs)
	}

	query += " ORDER BY received_at DESC"

	if filter.Limit != nil {
		argCount++
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, *filter.Limit)
	}

	if filter.Offset != nil {
		argCount++
		query += fmt.Sprintf(" OFFSET $%d", argCount)
		args = append(args, *filter.Offset)
	}

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query events: %w", err)
	}
	defer rows.Close()

	var events []*eventpb.Event
	for rows.Next() {
		var event eventpb.Event
		var payloadJSON []byte
		var receivedAt time.Time

		err := rows.Scan(&event.Provider, &event.Id, &event.EventName, &payloadJSON, &receivedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		if err := json.Unmarshal(payloadJSON, &event.Payload); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return events, nil
}

func (repo WebhookRepository) CountEvents(ctx context.Context) (int64, error) {
	var count int64
	err := repo.db.QueryRow(ctx, "SELECT COUNT(*) FROM webhook_events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}
	return count, nil
}

// EventExists checks if an event with the given provider and delivery_id exists
func (repo WebhookRepository) EventExists(ctx context.Context, provider int32, deliveryID string) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(ctx,
		"SELECT EXISTS(SELECT 1 FROM webhook_events WHERE provider = $1 AND delivery_id = $2)",
		provider, deliveryID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check event existence: %w", err)
	}
	return exists, nil
}
