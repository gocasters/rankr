package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gocasters/rankr/contributorapp/service/job"
	"github.com/gocasters/rankr/pkg/database"
	"time"
)

type FailRecordRepository struct {
	PostgresDB *database.Database
}

func NewFailRecordRepository(db *database.Database) FailRecordRepository {
	return FailRecordRepository{PostgresDB: db}
}

func (r FailRecordRepository) Create(ctx context.Context, fr job.FailRecord) error {

	query := `
		INSERT INTO fail_records (
			job_id,
			record_number,
			reason,
			raw_data,
			retry_count,
		    last_error,
		    error_type,
			created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`

	rawData, err := json.Marshal(fr.RawData)
	if err != nil {
		return fmt.Errorf("failed to marshal raw data: %w", err)
	}

	_, err = r.PostgresDB.Pool.Exec(ctx, query,
		fr.JobID,
		fr.RecordNumber,
		fr.Reason,
		rawData,
		fr.RetryCount,
		fr.LastError,
		fr.ErrType,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to create fail record: %w", err)
	}

	return nil
}

func (r FailRecordRepository) IncrementRetry(ctx context.Context, id uint, lastError string) error {

	query := `
		UPDATE fail_records
		SET retry_count = retry_count + 1,
		    last_error = $1,
		    updated_at = now()
		WHERE id = $2;
	`

	res, err := r.PostgresDB.Pool.Exec(ctx, query, lastError, id)
	if err != nil {
		return fmt.Errorf("failed to increment retry: %w", err)
	}

	if res.RowsAffected() == 0 {
		return job.ErrFailRecordNotFound
	}

	return nil
}

func (r FailRecordRepository) GetByJobID(ctx context.Context, jobID uint) ([]job.FailRecord, error) {

	query := `
		SELECT
			id,
			job_id,
			record_number,
			reason,
			raw_data,
			retry_count,
			last_error,
			error_type,
			created_at,
			updated_at
		FROM fail_records
		WHERE job_id = $1
		ORDER BY record_number;
	`

	rows, err := r.PostgresDB.Pool.Query(ctx, query, jobID)
	if err != nil {
		return nil, fmt.Errorf("failed to get fail records: %w", err)
	}
	defer rows.Close()

	var records []job.FailRecord

	for rows.Next() {
		var fr job.FailRecord
		var rawData []byte

		if err := rows.Scan(
			&fr.ID,
			&fr.JobID,
			&fr.RecordNumber,
			&fr.Reason,
			&rawData,
			&fr.RetryCount,
			&fr.LastError,
			&fr.CreatedAt,
			&fr.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(rawData, &fr.RawData); err != nil {
			return nil, fmt.Errorf("failed to unmarshal raw data: %w", err)
		}

		records = append(records, fr)
	}

	return records, nil
}
