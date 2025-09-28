package lostevent

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type LostWebhookRepository struct {
	db *pgxpool.Pool
}

// NewLostWebhookRepository creates a new lost webhook repository
func NewLostWebhookRepository(db *pgxpool.Pool) *LostWebhookRepository {
	return &LostWebhookRepository{db: db}
}

func (repo *LostWebhookRepository) Save(ctx context.Context, lostIDs *[]string) error {
	//save lost ID to db
	// TODO implement this
	return nil
}

// SaveBatch pgx.Batch for even better performance
func (repo *LostWebhookRepository) SaveBatch(ctx context.Context, provider int32, owner, repoName string, hookID int64, lostIDs []string) error {
	if len(lostIDs) == 0 {
		return nil
	}

	batch := &pgx.Batch{}
	query := `
		INSERT INTO lost_webhook_events (provider, owner, repo, hook_id, delivery_id) 
		VALUES ($1, $2, $3, $4, $5) 
		ON CONFLICT (provider, owner, repoName, hookID, delivery_id) DO NOTHING`

	for _, lostID := range lostIDs {
		batch.Queue(query, provider, owner, repoName, hookID, lostID)
	}

	results := repo.db.SendBatch(ctx, batch)
	defer results.Close()

	// Process all results
	for i := 0; i < len(lostIDs); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to save lost webhook event %d: %w", i, err)
		}
	}

	return nil
}

// GetAllEventsByProvider retrieves all lost delivery IDs for a specific provider
func (repo *LostWebhookRepository) GetAllEventsByProvider(ctx context.Context, provider int32) ([]*LostEventRow, error) {
	query := `
		SELECT provider, owner, repo, hook_id, delivery_id 
		FROM lost_webhook_events 
		WHERE provider = $1`

	rows, err := repo.db.Query(ctx, query, provider)
	if err != nil {
		return nil, fmt.Errorf("failed to query lost webhook events: %w", err)
	}
	defer rows.Close()

	var events []*LostEventRow
	for rows.Next() {
		var row LostEventRow
		if err := rows.Scan(&row.ID, &row.Provider, &row.Owner, &row.Repo, &row.HookID, &row.DeliveryID); err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}
		events = append(events, &row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return events, nil
}

// DeleteByID delete row by ID
func (repo *LostWebhookRepository) DeleteByID(ctx context.Context, provider int32, lostID string) error {
	query := `
		DELETE FROM lost_webhook_events 
		WHERE provider = $1 AND delivery_id = $2`

	result, err := repo.db.Exec(ctx, query, provider, lostID)
	if err != nil {
		return fmt.Errorf("failed to delete lost webhook event: %w", err)
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("lost webhook event not found: provider=%d, delivery_id=%s", provider, lostID)
	}

	return nil
}
