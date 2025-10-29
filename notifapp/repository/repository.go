package repository

import (
	"github.com/gocasters/rankr/pkg/database"
)

// TODO: Implement database repository
// - Create notifications table: id, user_id, message, type, status, created_at, read_at
// - Add indexes on user_id and (user_id, status) for performance
// - Map database errors to domain errors (ErrNotFound, etc.)
// - Always include user_id in WHERE clauses for security
// - Use parameterized queries to prevent SQL injection
// - Consider pagination for List method with large datasets
// - Add query timeouts using context
// - Implement soft delete strategy (add deleted_at column)
// - Write integration tests with database

func New(db *database.Database) Repository {
	return Repository{db: db}
}

type Repository struct {
	db *database.Database
}
