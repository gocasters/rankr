package repository

import (
	"context"

	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/statuscode"
)

type authRepository struct {
	db *database.Database
}

func NewRepository(db *database.Database) auth.Repository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) ListPermissionsByRoleName(ctx context.Context, roleName string) ([]string, error) {
	const query = `
		SELECT p.name
		FROM role_permissions rp
		LEFT JOIN permissions p ON p.id = rp.permission_id
		WHERE rp.role = $1
		ORDER BY p.name
	`

	rows, err := r.db.Pool.Query(ctx, query, roleName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var (
		permissions []string
		hasRows     bool
	)
	for rows.Next() {
		hasRows = true
		var name *string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		if name != nil {
			permissions = append(permissions, *name)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if !hasRows {
		return nil, statuscode.ErrRoleNotFound
	}
	return permissions, nil
}
