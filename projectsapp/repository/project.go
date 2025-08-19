package repository

import (
	"context"
	"fmt"

	"github.com/gocasters/rankr/projectsapp/constants"
	"github.com/gocasters/rankr/projectsapp/helpers"
	"github.com/gocasters/rankr/projectsapp/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProjectRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewProjectRepositoryPG(pool *pgxpool.Pool) *ProjectRepositoryPG {
	return &ProjectRepositoryPG{pool: pool}
}

const (
	sqlProjectInsert = `
		INSERT INTO projects (id, name, slug, description, design_reference_url, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at, archived_at;
	`

	sqlProjectByID = `
		SELECT id, name, slug, description, design_reference_url, status, created_at, updated_at, archived_at
		FROM projects
		WHERE id = $1;
	`

	sqlProjectBySlug = `
		SELECT id, name, slug, description, design_reference_url, status, created_at, updated_at, archived_at
		FROM projects
		WHERE slug = $1;
	`

	sqlProjectList = `
		SELECT id, name, slug, description, design_reference_url, status, created_at, updated_at, archived_at
		FROM projects
		ORDER BY created_at DESC;
	`

	sqlProjectUpdate = `
		UPDATE projects
		SET name = $2,
		    slug = $3,
		    description = $4,
		    design_reference_url = $5,
		    status = $6
		WHERE id = $1
		RETURNING created_at, updated_at, archived_at;
	`

	sqlProjectDelete = `DELETE FROM projects WHERE id = $1;`
)

func (r *ProjectRepositoryPG) Create(ctx context.Context, p *types.ProjectEntity) error {

	row := r.pool.QueryRow(ctx, sqlProjectInsert,
		p.ID, p.Name, p.Slug, p.Description, p.DesignReferenceURL, p.Status,
	)
	if err := row.Scan(&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt); err != nil {
		if helpers.IsUniqueViolation(err) {
			return fmt.Errorf("%w: slug", constants.ErrUniqueConstraint)
		}
		return err
	}
	return nil
}

func (r *ProjectRepositoryPG) FindByID(ctx context.Context, id string) (*types.ProjectEntity, error) {
	var p types.ProjectEntity
	err := r.pool.
		QueryRow(ctx, sqlProjectByID, id).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.DesignReferenceURL, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if err != nil {
		if helpers.IsNoRows(err) {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepositoryPG) FindBySlug(ctx context.Context, slug string) (*types.ProjectEntity, error) {
	var p types.ProjectEntity
	err := r.pool.
		QueryRow(ctx, sqlProjectBySlug, slug).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.DesignReferenceURL, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if err != nil {
		if helpers.IsNoRows(err) {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepositoryPG) List(ctx context.Context) ([]*types.ProjectEntity, error) {
	rows, err := r.pool.Query(ctx, sqlProjectList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*types.ProjectEntity
	for rows.Next() {
		var p types.ProjectEntity
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.DesignReferenceURL, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *ProjectRepositoryPG) Update(ctx context.Context, p *types.ProjectEntity) error {
	row := r.pool.QueryRow(ctx, sqlProjectUpdate,
		p.ID, p.Name, p.Slug, p.Description, p.DesignReferenceURL, p.Status,
	)
	if err := row.Scan(&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt); err != nil {
		if helpers.IsNoRows(err) {
			return constants.ErrNotFound
		}
		if helpers.IsUniqueViolation(err) {
			return fmt.Errorf("%w: slug", constants.ErrUniqueConstraint)
		}
		return err
	}
	return nil
}

func (r *ProjectRepositoryPG) Delete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, sqlProjectDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return constants.ErrNotFound
	}
	return nil
}
