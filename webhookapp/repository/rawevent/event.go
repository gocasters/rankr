package rawevent

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/webhookapp/service"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrDuplicateEvent = errors.New("duplicate webhook event")

// RawEventFilter defines filters for querying raw events
type RawEventFilter struct {
	Provider    *int32
	Owner       *string
	RepoName    *string
	HookID      *int64
	StartTime   *time.Time
	EndTime     *time.Time
	DeliveryIDs []string
	Limit       *int
	Offset      *int
}

// RawWebhookRepository handles raw webhook event persistence
type RawWebhookRepository struct {
	db *pgxpool.Pool
}

// NewRawWebhookRepository creates a new raw webhook repository
func NewRawWebhookRepository(db *pgxpool.Pool) *RawWebhookRepository {
	return &RawWebhookRepository{db: db}
}

// Save saves a raw webhook event to the database
func (repo RawWebhookRepository) Save(ctx context.Context, event *service.RawEvent) error {
	query := `
		INSERT INTO raw_webhook_events (provider, hook_id, owner, repo, event_type, delivery_id, payload_json) 
		VALUES ($1, $2, $3,$4, $5, $6, $7)`

	result, err := repo.db.Exec(ctx, query, event.Provider, event.HookID, event.Owner, event.Repo,
		event.EventType, event.DeliveryID, event.Payload)
	if err != nil {
		return fmt.Errorf("failed to save raw webhook event: %w", err)
	}

	if result.RowsAffected() == 0 {
		return fmt.Errorf(ErrDuplicateEvent.Error()+"provider %d , delivery_id %s", event.Provider, event.DeliveryID)
	}

	return nil
}

// FindByDeliveryID retrieves a raw event by provider and delivery_id
func (repo RawWebhookRepository) FindByDeliveryID(ctx context.Context, provider int32, deliveryID string) (*WebhookEventRow, error) {
	var row WebhookEventRow
	err := repo.db.QueryRow(
		ctx,
		`SELECT id, provider, owner, repo, hook_id, delivery_id, payload_json, received_at 
         FROM raw_webhook_events 
         WHERE provider=$1 AND delivery_id=$2`,
		provider, deliveryID,
	).Scan(&row.ID, &row.Provider, &row.Owner, &row.Repo, &row.HookID, &row.DeliveryID, &row.PayloadJSON, &row.ReceivedAt)

	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("raw event not found")
		}
		return nil, fmt.Errorf("failed to query raw event: %w", err)
	}

	return &row, nil
}

// FindEvents retrieves raw events based on filters
func (repo RawWebhookRepository) FindEvents(ctx context.Context, filter RawEventFilter) ([]*WebhookEventRow, error) {
	query := `SELECT id, provider, owner, repo, hook_id, delivery_id, payload_json, received_at FROM raw_webhook_events WHERE 1=1`
	args := make([]interface{}, 0)
	argCount := 0

	if filter.Provider != nil {
		argCount++
		query += fmt.Sprintf(" AND provider=$%d", argCount)
		args = append(args, *filter.Provider)
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
		return nil, fmt.Errorf("failed to query raw events: %w", err)
	}
	defer rows.Close()

	var events []*WebhookEventRow
	for rows.Next() {
		var row WebhookEventRow
		if err := rows.Scan(&row.ID, &row.Provider, &row.Owner, &row.Repo, &row.HookID, &row.DeliveryID, &row.PayloadJSON, &row.ReceivedAt); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		events = append(events, &row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return events, nil
}

// CountEvents returns total number of raw events
func (repo RawWebhookRepository) CountEvents(ctx context.Context) (int64, error) {
	var count int64
	err := repo.db.QueryRow(ctx, "SELECT COUNT(*) FROM raw_webhook_events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count raw events: %w", err)
	}
	return count, nil
}

// EventExists checks if a raw event exists by provider and delivery_id
func (repo RawWebhookRepository) EventExists(ctx context.Context, provider int32, deliveryID string) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM raw_webhook_events WHERE provider=$1 AND delivery_id=$2)",
		provider, deliveryID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check raw event existence: %w", err)
	}
	return exists, nil
}

// GetEventsByProvider retrieves raw events for a specific provider
func (repo RawWebhookRepository) GetEventsByProvider(ctx context.Context, provider int32, limit *int) ([]*WebhookEventRow, error) {
	filter := RawEventFilter{
		Provider: &provider,
		Limit:    limit,
	}
	return repo.FindEvents(ctx, filter)
}

// GetEventsByTimeRange retrieves raw events within a time range
func (repo RawWebhookRepository) GetEventsByTimeRange(ctx context.Context, start, end time.Time, limit *int) ([]*WebhookEventRow, error) {
	filter := RawEventFilter{
		StartTime: &start,
		EndTime:   &end,
		Limit:     limit,
	}
	return repo.FindEvents(ctx, filter)
}

// GetRecentEvents retrieves the most recent raw events
func (repo RawWebhookRepository) GetRecentEvents(ctx context.Context, limit *int) ([]*WebhookEventRow, error) {
	filter := RawEventFilter{
		Limit: limit,
	}
	return repo.FindEvents(ctx, filter)
}

// GetEventsWithProviderAndTimeRange retrieves raw events for a provider within a time range
func (repo RawWebhookRepository) GetEventsByProviderOwnerRepoWithTimeRange(ctx context.Context, provider int32, owner, repoName string, hookID int64, start, end time.Time, limit *int) ([]*WebhookEventRow, error) {
	filter := RawEventFilter{
		Provider:  &provider,
		Owner:     &owner,
		RepoName:  &repoName,
		HookID:    &hookID,
		StartTime: &start,
		EndTime:   &end,
		Limit:     limit,
	}
	return repo.FindEvents(ctx, filter)
}

// FindExistingDeliveryIDs checks a slice of delivery IDs and returns a map
// of the ones that already exist in the database. This is more efficient
// than checking one by one.
func (repo RawWebhookRepository) FindExistingDeliveryIDs(ctx context.Context, deliveryIDs []string) (map[string]bool, error) {
	if len(deliveryIDs) == 0 {
		return make(map[string]bool), nil
	}

	query := `SELECT delivery_id FROM raw_webhook_events WHERE delivery_id = ANY($1)`
	rows, err := repo.db.Query(ctx, query, deliveryIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to query existing delivery IDs: %w", err)
	}
	defer rows.Close()

	existingIDs := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("failed to scan delivery ID: %w", err)
		}
		existingIDs[id] = true
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error in FindExistingDeliveryIDs: %w", err)
	}

	return existingIDs, nil
}
