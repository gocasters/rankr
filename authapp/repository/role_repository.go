package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/pkg/statuscode"
	types "github.com/gocasters/rankr/type"
	"github.com/jackc/pgx/v5"
)

type authRepository struct {
	db *database.Database
}

func NewRepository(db *database.Database) auth.Repository {
	return &authRepository{
		db: db,
	}
}

func (r *authRepository) Create(ctx context.Context, role auth.Role) (types.ID, error) {
	const query = `
		INSERT INTO roles (name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	if role.CreatedAt.IsZero() {
		role.CreatedAt = now
	}
	if role.UpdatedAt.IsZero() {
		role.UpdatedAt = now
	}

	var id uint64
	if err := r.db.Pool.QueryRow(ctx, query,
		role.Name,
		role.Description,
		role.CreatedAt,
		role.UpdatedAt,
	).Scan(&id); err != nil {
		return 0, err
	}

	return types.ID(id), nil
}

func (r *authRepository) Get(ctx context.Context, roleID types.ID) (auth.Role, error) {
	const query = `
		SELECT r.id,
		       r.name,
		       r.description,
		       r.created_at,
		       r.updated_at,
		       COALESCE(
				   json_agg(
					   json_build_object(
						   'id', p.id,
						   'name', p.name,
						   'description', p.description,
						   'created_at', p.created_at,
						   'updated_at', p.updated_at
					   )
					   ORDER BY p.name
				   ) FILTER (WHERE p.id IS NOT NULL),
				   '[]'
			   ) AS permissions
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		LEFT JOIN permissions p ON p.id = rp.permission_id
		WHERE r.id = $1
		GROUP BY r.id
	`

	var (
		role          auth.Role
		rawID         uint64
		permissionsJS []byte
	)

	if err := r.db.Pool.QueryRow(ctx, query, uint64(roleID)).Scan(
		&rawID,
		&role.Name,
		&role.Description,
		&role.CreatedAt,
		&role.UpdatedAt,
		&permissionsJS,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return auth.Role{}, statuscode.ErrRoleNotFound
		}
		return auth.Role{}, err
	}

	role.ID = types.ID(rawID)
	if len(permissionsJS) > 0 {
		if err := json.Unmarshal(permissionsJS, &role.Permissions); err != nil {
			return auth.Role{}, fmt.Errorf("failed to decode permissions: %w", err)
		}
	}

	return role, nil
}

func (r *authRepository) Update(ctx context.Context, role auth.Role) error {
	const query = `
		UPDATE roles
		SET name = COALESCE(NULLIF($2, ''), name),
		    description = COALESCE(NULLIF($3, ''), description),
		    updated_at = $4
		WHERE id = $1
	`

	tag, err := r.db.Pool.Exec(ctx, query,
		uint64(role.ID),
		role.Name,
		role.Description,
		role.UpdatedAt,
	)
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return statuscode.ErrRoleNotFound
	}

	return nil
}

func (r *authRepository) Delete(ctx context.Context, roleID types.ID) error {
	const query = `DELETE FROM roles WHERE id = $1`

	tag, err := r.db.Pool.Exec(ctx, query, uint64(roleID))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return statuscode.ErrRoleNotFound
	}
	return nil
}

func (r *authRepository) List(ctx context.Context, page, pageSize int) ([]auth.Role, error) {
	const query = `
		SELECT r.id,
		       r.name,
		       r.description,
		       r.created_at,
		       r.updated_at,
		       COALESCE(
				   json_agg(
					   json_build_object(
						   'id', p.id,
						   'name', p.name,
						   'description', p.description,
						   'created_at', p.created_at,
						   'updated_at', p.updated_at
					   )
					   ORDER BY p.name
				   ) FILTER (WHERE p.id IS NOT NULL),
				   '[]'
			   ) AS permissions
		FROM roles r
		LEFT JOIN role_permissions rp ON rp.role_id = r.id
		LEFT JOIN permissions p ON p.id = rp.permission_id
		GROUP BY r.id
		ORDER BY r.updated_at DESC
		LIMIT $1 OFFSET $2
	`

	offset := (page - 1) * pageSize
	rows, err := r.db.Pool.Query(ctx, query, pageSize, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []auth.Role
	for rows.Next() {
		var (
			role          auth.Role
			rawID         uint64
			permissionsJS []byte
		)

		if err := rows.Scan(
			&rawID,
			&role.Name,
			&role.Description,
			&role.CreatedAt,
			&role.UpdatedAt,
			&permissionsJS,
		); err != nil {
			return nil, err
		}

		role.ID = types.ID(rawID)
		if len(permissionsJS) > 0 {
			if err := json.Unmarshal(permissionsJS, &role.Permissions); err != nil {
				return nil, fmt.Errorf("failed to decode permissions: %w", err)
			}
		}

		roles = append(roles, role)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

func (r *authRepository) AddPermission(ctx context.Context, roleID, permissionID types.ID) error {
	if err := r.ensureRoleExists(ctx, roleID); err != nil {
		return err
	}
	if err := r.ensurePermissionExists(ctx, permissionID); err != nil {
		return err
	}

	const query = `
		INSERT INTO role_permissions (role_id, permission_id, created_at)
		VALUES ($1, $2, now())
		ON CONFLICT (role_id, permission_id) DO NOTHING
	`

	_, err := r.db.Pool.Exec(ctx, query, uint64(roleID), uint64(permissionID))
	return err
}

func (r *authRepository) RemovePermission(ctx context.Context, roleID, permissionID types.ID) error {
	if err := r.ensureRoleExists(ctx, roleID); err != nil {
		return err
	}
	if err := r.ensurePermissionExists(ctx, permissionID); err != nil {
		return err
	}

	const query = `
		DELETE FROM role_permissions
		WHERE role_id = $1 AND permission_id = $2
	`

	_, err := r.db.Pool.Exec(ctx, query, uint64(roleID), uint64(permissionID))
	return err
}

func (r *authRepository) ensureRoleExists(ctx context.Context, roleID types.ID) error {
	const query = `SELECT 1 FROM roles WHERE id = $1`
	if err := r.db.Pool.QueryRow(ctx, query, uint64(roleID)).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statuscode.ErrRoleNotFound
		}
		return err
	}
	return nil
}

func (r *authRepository) ensurePermissionExists(ctx context.Context, permissionID types.ID) error {
	const query = `SELECT 1 FROM permissions WHERE id = $1`
	if err := r.db.Pool.QueryRow(ctx, query, uint64(permissionID)).Scan(new(int)); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return statuscode.ErrPermissionNotFound
		}
		return err
	}
	return nil
}
