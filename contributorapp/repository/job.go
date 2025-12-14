package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/contributorapp/service/job"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/jackc/pgx/v5"
	"time"
)

type JobRepository struct {
	PostgresDB *database.Database
}

func NewJobRepository(db *database.Database) JobRepository {
	return JobRepository{PostgresDB: db}
}

func (r JobRepository) CreateJob(ctx context.Context, job job.Job) (uint, error) {
	query := `INSERT INTO jobs (idempotency_key, file_path, file_name, file_hash, status, total_records, success_count, fail_count, created_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
              RETURNING id;
             `

	var id int64
	err := r.PostgresDB.Pool.QueryRow(ctx, query,
		job.IdempotencyKey,
		job.FilePath,
		job.FileName,
		job.FileHash,
		job.Status,
		job.TotalRecords,
		job.SuccessCount,
		job.FailCount,
		job.CreatedAt,
	).Scan(&id)

	if err != nil {
		return 0, fmt.Errorf("failed to create job: %w", err)
	}

	return uint(id), nil
}

func (r JobRepository) GetJobByIdempotencyKey(ctx context.Context, key string) (*job.Job, error) {
	query := `SELECT id, idempotency_key, file_path, file_name, file_hash, status, total_records, success_count, fail_count, created_at, updated_at
              FROM jobs
              WHERE idempotency_key=$1;
              `

	row := r.PostgresDB.Pool.QueryRow(ctx, query, key)

	j := new(job.Job)

	err := row.Scan(
		&j.ID,
		&j.IdempotencyKey,
		&j.FilePath,
		&j.FileName,
		&j.FileHash,
		&j.Status,
		&j.TotalRecords,
		&j.SuccessCount,
		&j.FailCount,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, job.ErrJobNotExists
		}

		return nil, err
	}

	return j, nil
}

func (r JobRepository) GetJobByFileHash(ctx context.Context, fileHash string) (*job.Job, error) {
	query := `SELECT id, idempotency_key, file_path, file_name, file_hash, status, total_records, success_count, fail_count, created_at, updated_at
              FROM jobs
              WHERE file_hash=$1;
              `

	row := r.PostgresDB.Pool.QueryRow(ctx, query, fileHash)

	j := new(job.Job)

	err := row.Scan(
		&j.ID,
		&j.IdempotencyKey,
		&j.FilePath,
		&j.FileName,
		&j.FileHash,
		&j.Status,
		&j.TotalRecords,
		&j.SuccessCount,
		&j.FailCount,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, job.ErrJobNotExists
		}

		return nil, err
	}

	return j, nil
}

func (r JobRepository) GetJobByID(ctx context.Context, ID uint) (job.Job, error) {
	query := `SELECT id, idempotency_key, file_path, file_name, file_hash, status, total_records, success_count, fail_count, created_at, updated_at
              FROM jobs
              WHERE id=$1;
              `

	row := r.PostgresDB.Pool.QueryRow(ctx, query, ID)

	var j job.Job

	err := row.Scan(
		&j.ID,
		&j.IdempotencyKey,
		&j.FilePath,
		&j.FileName,
		&j.FileHash,
		&j.Status,
		&j.TotalRecords,
		&j.SuccessCount,
		&j.FailCount,
		&j.CreatedAt,
		&j.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return job.Job{}, job.ErrJobNotExists
		}

		return job.Job{}, err
	}

	return j, nil
}

func (r JobRepository) UpdateStatus(ctx context.Context, jobID uint, status job.Status) error {
	query := `UPDATE jobs
			 SET status=$1, updated_at=$2
			 WHERE id=$3;
             `

	res, err := r.PostgresDB.Pool.Exec(ctx, query, status, time.Now(), jobID)
	if err != nil {
		return err
	}

	if rowAffect := res.RowsAffected(); rowAffect == 0 {
		return job.ErrJobNotExists
	}

	return nil
}

func (r JobRepository) UpdateJob(ctx context.Context, j job.Job) error {
	query := `UPDATE jobs
              SET status=$1,
                  total_records=$2,
                  success_count=$3,
                  fail_count=$4,
                  updated_at=$5
              WHERE id=$6;
              `
	res, err := r.PostgresDB.Pool.Exec(ctx, query,
		j.Status,
		j.TotalRecords,
		j.SuccessCount,
		j.FailCount,
		time.Now(),
		j.ID,
	)

	if err != nil {
		return fmt.Errorf("failed to update job: %w", err)
	}

	if rowAffect := res.RowsAffected(); rowAffect == 0 {
		return job.ErrJobNotExists
	}

	return nil
}
