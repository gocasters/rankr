package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/logger"
	"strings"
	"time"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/protobuf/proto"
)

var ErrDuplicateEvent = errors.New("duplicate webhook event")

type EventStats struct {
	TotalEvents      int64            `json:"total_events"`
	EventsByProvider map[int32]int64  `json:"events_by_provider"`
	EventsByType     map[string]int64 `json:"events_by_type"`
	FirstEventAt     *time.Time       `json:"first_event_at"`
	LastEventAt      *time.Time       `json:"last_event_at"`
}

type EventFilter struct {
	Provider    *int32
	EventType   *int32 // store protobuf enum as int32
	StartTime   *time.Time
	EndTime     *time.Time
	DeliveryIDs []string
	Limit       *int
	Offset      *int
}

type WebhookRepository struct {
	db *pgxpool.Pool
}

// NewWebhookRepository returns a WebhookRepository backed by the provided pgx connection pool.
func NewWebhookRepository(db *pgxpool.Pool) WebhookRepository {
	return WebhookRepository{
		db: db,
	}
}

// Save stores a webhook event
func (repo WebhookRepository) Save(ctx context.Context, event *eventpb.Event) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	tag, err := repo.db.Exec(
		ctx,
		`INSERT INTO webhook_events (provider, delivery_id, event_type, payload, received_at)
	 VALUES ($1, $2, $3, $4, $5)
	 ON CONFLICT (provider, delivery_id) DO NOTHING`,
		event.Provider,
		event.Id,
		event.EventName,
		payload,
		time.Now(),
	)

	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrDuplicateEvent
	}
	return nil
}

// FindByDeliveryID retrieves an event by provider and delivery_id
func (repo WebhookRepository) FindByDeliveryID(ctx context.Context, provider int32, deliveryID string) (*eventpb.Event, error) {
	var payloadBytes []byte
	err := repo.db.QueryRow(
		ctx,
		`SELECT payload FROM webhook_events WHERE provider=$1 AND delivery_id=$2`,
		provider, deliveryID,
	).Scan(&payloadBytes)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("event not found")
		}
		return nil, fmt.Errorf("failed to query event: %w", err)
	}

	var event eventpb.Event
	if err := proto.Unmarshal(payloadBytes, &event); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event: %w", err)
	}
	return &event, nil
}

// FindEvents retrieves events based on filters
func (repo WebhookRepository) FindEvents(ctx context.Context, filter EventFilter) ([]*eventpb.Event, error) {
	query := `SELECT payload FROM webhook_events WHERE 1=1`
	args := make([]interface{}, 0)
	argCount := 0

	if filter.Provider != nil {
		argCount++
		query += fmt.Sprintf(" AND provider=$%d", argCount)
		args = append(args, *filter.Provider)
	}
	if filter.EventType != nil {
		argCount++
		query += fmt.Sprintf(" AND event_type=$%d", argCount)
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
		var payloadBytes []byte
		if err := rows.Scan(&payloadBytes); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		var event eventpb.Event
		if err := proto.Unmarshal(payloadBytes, &event); err != nil {
			return nil, fmt.Errorf("failed to unmarshal payload: %w", err)
		}

		events = append(events, &event)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return events, nil
}

// CountEvents returns total number of events
func (repo WebhookRepository) CountEvents(ctx context.Context) (int64, error) {
	var count int64
	err := repo.db.QueryRow(ctx, "SELECT COUNT(*) FROM webhook_events").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count events: %w", err)
	}
	return count, nil
}

// EventExists checks if an event exists by provider and delivery_id
func (repo WebhookRepository) EventExists(ctx context.Context, provider int32, deliveryID string) (bool, error) {
	var exists bool
	err := repo.db.QueryRow(
		ctx,
		"SELECT EXISTS(SELECT 1 FROM webhook_events WHERE provider=$1 AND delivery_id=$2)",
		provider, deliveryID,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check event existence: %w", err)
	}
	return exists, nil
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
func (repo WebhookRepository) GetEventsByType(ctx context.Context, eventType int32, limit int) ([]*eventpb.Event, error) {
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

func (repo *WebhookRepository) BulkInsertPostgresSQL(ctx context.Context, events []string) ([]*eventpb.Event, error) {
	if len(events) == 0 {
		logger.L().Debug("No events to process")
		return nil, nil
	}

	tx, err := repo.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		logger.L().Error("Failed to begin transaction", "error", err.Error())
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
			logger.L().Warn("Rollback error", "error", err.Error())
		}
	}()

	// Create a new batch for this operation
	batch := &pgx.Batch{}
	eventMap := make([]*eventpb.Event, 0, len(events))

	sqlQuery := `
		INSERT INTO webhook_events 
		(provider, delivery_id, event_type, payload, received_at)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (provider, delivery_id) DO NOTHING
	`

	for i, raw := range events {
		var event eventpb.Event
		if err := proto.Unmarshal([]byte(raw), &event); err != nil {
			logger.L().Warn("Failed to unmarshal event, skipping",
				"index", i, "error", err.Error())
			continue
		}

		// Convert enum values to strings

		logger.L().Debug("Adding event to batch",
			"index", i,
			"provider", event.Provider,
			"delivery_id", event.Id,
			"event_type", event.EventName)

		batch.Queue(sqlQuery,
			event.Provider,
			event.Id,
			event.EventName,
			event.Payload,
			time.Now(),
		)

		eventMap = append(eventMap, &event)
	}

	if batch.Len() == 0 {
		logger.L().Warn("No valid events to insert after unmarshaling")
		return nil, nil
	}

	logger.L().Debug("Sending batch to database", "batch_size", batch.Len())

	// Send batch and process results
	results := tx.SendBatch(ctx, batch)

	// CRITICAL: Process all results AND close the results before committing
	var inserted []*eventpb.Event
	var failedCount int

	for i := 0; i < batch.Len(); i++ {
		cmdTag, err := results.Exec()
		if err != nil {
			failedCount++
			logger.L().Error("Failed to execute batch insert",
				"index", i,
				"error", err.Error())
			continue
		}

		rowsAffected := cmdTag.RowsAffected()
		if rowsAffected > 0 {
			inserted = append(inserted, eventMap[i])
			logger.L().Debug("Event inserted successfully",
				"index", i,
				"rows_affected", rowsAffected)
		} else {
			logger.L().Debug("No rows affected - duplicate or conflict",
				"index", i)
		}
	}

	// MUST close results before committing
	if err := results.Close(); err != nil {
		logger.L().Error("Failed to close batch results", "error", err.Error())
		return nil, fmt.Errorf("failed to close batch results: %w", err)
	}

	// Now commit the transaction
	if err := tx.Commit(ctx); err != nil {
		logger.L().Error("Failed to commit transaction", "error", err.Error())
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	logger.L().Info("Bulk insert completed",
		"total_events", len(events),
		"successful_inserts", len(inserted),
		"failed_inserts", failedCount)

	return inserted, nil
}

func (repo *WebhookRepository) GetLostDeliveries(ctx context.Context, provider eventpb.EventProvider, deliveries []string) ([]string, error) {
	if len(deliveries) == 0 {
		return []string{}, nil
	}

	// Convert provider enum to int32 for database query
	providerID := int32(provider)

	// Create a temporary table structure using VALUES
	placeholders := make([]string, len(deliveries))
	args := make([]interface{}, len(deliveries)+1)

	args[0] = providerID
	for i, deliveryID := range deliveries {
		placeholders[i] = fmt.Sprintf("($%d)", i+2)
		args[i+1] = deliveryID
	}

	query := fmt.Sprintf(`
		WITH expected_deliveries(delivery_id) AS (
			VALUES %s
		)
		SELECT ed.delivery_id
		FROM expected_deliveries ed
		WHERE NOT EXISTS (
			SELECT 1 
			FROM webhook_events we 
			WHERE we.provider = $1 
			AND we.delivery_id = ed.delivery_id
		)
	`, strings.Join(placeholders, ","))

	rows, err := repo.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query missing deliveries: %w", err)
	}
	defer rows.Close()

	var missingDeliveries []string
	for rows.Next() {
		var deliveryID string
		if err := rows.Scan(&deliveryID); err != nil {
			return nil, fmt.Errorf("failed to scan delivery ID: %w", err)
		}
		missingDeliveries = append(missingDeliveries, deliveryID)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return missingDeliveries, nil
}

func (repo *WebhookRepository) SaveHistoricalEvent(
	ctx context.Context,
	event *eventpb.Event,
	resourceType string,
	resourceID int64,
) error {
	payload, err := proto.Marshal(event)
	if err != nil {
		return fmt.Errorf("failed to marshal event: %w", err)
	}

	tag, err := repo.db.Exec(
		ctx,
		`INSERT INTO webhook_events
		(provider, source, resource_type, resource_id, event_type, payload, received_at, delivery_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NULL)
		ON CONFLICT (provider, resource_type, resource_id, event_type)
		WHERE source = 'historical'
		DO NOTHING`,
		event.Provider,
		"historical",
		resourceType,
		resourceID,
		event.EventName,
		payload,
		time.Now(),
	)

	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return ErrDuplicateEvent
	}

	return nil
}
