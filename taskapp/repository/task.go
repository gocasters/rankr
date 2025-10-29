package repository

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/taskapp/service/task"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/lib/pq"
)

type Config struct {
	CacheEnabled bool   `koanf:"cache_enabled"`
	CachePrefix  string `koanf:"cache_prefix"`
}

type TaskRepo struct {
	Config     Config
	Logger     *slog.Logger
	PostgreSQL *database.Database
	Cache      *redis.Adapter
}

func NewTaskRepo(config Config, db *database.Database, logger *slog.Logger) task.Repository {
	return &TaskRepo{
		Config:     config,
		Logger:     logger,
		PostgreSQL: db,
	}
}

func (r *TaskRepo) CreateTask(ctx context.Context, param task.CreateTaskParam) error {
	query := `
		INSERT INTO tasks (version_control_system_id, issue_number, title, state, repository_name, labels, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $7)
		ON CONFLICT DO NOTHING
	`

	_, err := r.PostgreSQL.Pool.Exec(ctx, query,
		param.VersionControlSystemId,
		param.IssueNumber,
		param.Title,
		param.State,
		param.RepositoryName,
		pq.Array(param.Labels),
		param.CreatedAt,
	)

	if err != nil {

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {

			switch pgErr.Code[0:2] {
			case "08", "53", "57", "58":
				r.Logger.Error("Database connection/system error",
					slog.String("error", err.Error()),
					slog.String("pg_code", pgErr.Code),
				)
				return task.NewRetriableError(err, "database connection or system error")
			case "23":

				r.Logger.Warn("Integrity constraint violation not handled by ON CONFLICT",
					slog.String("error", err.Error()),
					slog.String("pg_code", pgErr.Code),
				)
				return task.ErrTaskAlreadyExists
			default:
				r.Logger.Error("Database error",
					slog.String("error", err.Error()),
					slog.String("pg_code", pgErr.Code),
				)
				return task.NewRetriableError(err, "database error")
			}
		}

		r.Logger.Error("Failed to create task",
			slog.String("error", err.Error()),
		)
		return task.NewRetriableError(err, "failed to create task")
	}

	return nil
}

func (r *TaskRepo) UpdateTaskByIssueNumber(ctx context.Context, param task.UpdateTaskParam) error {
	query := `
		UPDATE tasks
		SET state = $1, closed_at = $2, updated_at = NOW()
		WHERE issue_number = $3 AND repository_name = $4
	`

	var closedAt interface{}
	if param.ClosedAt != nil {
		closedAt = *param.ClosedAt
	} else {
		closedAt = nil
	}

	result, err := r.PostgreSQL.Pool.Exec(ctx, query,
		param.State,
		closedAt,
		param.IssueNumber,
		param.RepositoryName,
	)

	if err != nil {
		r.Logger.Error("Failed to update task", slog.String("error", err.Error()))
		return task.NewRetriableError(err, "failed to update task in database")
	}

	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		r.Logger.Warn("No task found to update",
			slog.Int("issue_number", param.IssueNumber),
			slog.String("repository", param.RepositoryName),
		)

		return task.ErrTaskNotFound
	}

	return nil
}

func (r *TaskRepo) GetTaskByIssueNumber(ctx context.Context, issueNumber int, repositoryName string) (*task.Task, error) {
	query := `
		SELECT id, version_control_system_id, issue_number, title, state, repository_name, labels, created_at, updated_at, closed_at
		FROM tasks
		WHERE issue_number = $1 AND repository_name = $2
	`

	var t task.Task
	var closedAt sql.NullTime
	var labels []string

	err := r.PostgreSQL.Pool.QueryRow(ctx, query, issueNumber, repositoryName).Scan(
		&t.ID,
		&t.VersionControlSystemId,
		&t.IssueNumber,
		&t.Title,
		&t.State,
		&t.RepositoryName,
		pq.Array(&labels),
		&t.CreatedAt,
		&t.UpdatedAt,
		&closedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, task.ErrTaskNotFound
		}
		r.Logger.Error("Failed to get task", slog.String("error", err.Error()))
		return nil, task.NewRetriableError(err, "failed to get task from database")
	}

	if closedAt.Valid {
		t.ClosedAt = &closedAt.Time
	}

	t.Labels = make([]task.Label, len(labels))
	for i, labelName := range labels {
		t.Labels[i] = task.Label{
			Name: labelName,
		}
	}

	return &t, nil
}
