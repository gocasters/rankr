package repository

import (
        "context"
        "errors"
        "fmt"
        "time"
        "github.com/gocasters/rankr/pkg/database"
        "github.com/jackc/pgx/v5"
    )

type RoleRepository struct {
    db *database.Database
}

func NewRoleRepository(db *database.Database) *RoleRepository {
    return &RoleRepository{db: db}
}

// GetRoleByUserID returns the role for the given userID, or ErrRoleNotFound if no row exists
func (r *RoleRepository) GetRoleByUserID(ctx context.Context, userID string) (string, error) {
    var role string
    query := `SELECT role FROM user_roles WHERE user_id = $1`

    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

        err := r.db.Pool.QueryRow(ctx, query, userID).Scan(&role)
        if err != nil {
            if errors.Is(err, pgx.ErrNoRows) {
                return "", ErrRoleNotFound
            }
            return "", fmt.Errorf("GetRoleByUserID failed: %w", err)
        }
    return role, nil
}
var ErrRoleNotFound = errors.New("role not found")

// AssignRole 
func (r *RoleRepository) AssignRole(ctx context.Context, userID, role string) error {
    query := `
        INSERT INTO user_roles (user_id, role, created_at, updated_at)
        VALUES ($1, $2, now(), now())
        ON CONFLICT (user_id)
        DO UPDATE SET role = EXCLUDED.role, updated_at = now();
    `

    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    _, err := r.db.Pool.Exec(ctx, query, userID, role)
    if err != nil {
        return fmt.Errorf("AssignRole failed: %w", err)
    }
    return nil
}

// DeleteRole 
func (r *RoleRepository) DeleteRole(ctx context.Context, userID string) error {
    query := `DELETE FROM user_roles WHERE user_id = $1`

    ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
    defer cancel()

    _, err := r.db.Pool.Exec(ctx, query, userID)
    if err != nil {
        return fmt.Errorf("DeleteRole failed: %w", err)
    }
    return nil
}
