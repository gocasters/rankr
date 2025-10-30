package repository

import (
	"context"
	"fmt"

	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/gocasters/rankr/projectapp/service/project"
)

type ProjectRepository struct {
	database *database.Database
}

func NewProjectRepository(database *database.Database) project.Repository {
	return &ProjectRepository{database: database}
}

const (
	sqlProjectInsert = `
		INSERT INTO projects (id, name, slug, description, design_reference_url, git_repo_id, repo_provider, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING created_at, updated_at, archived_at;
	`

	sqlProjectByID = `
		SELECT id, name, slug, description, design_reference_url, git_repo_id, repo_provider, status, created_at, updated_at, archived_at
		FROM projects
		WHERE id = $1;
	`

	sqlProjectBySlug = `
		SELECT id, name, slug, description, design_reference_url, git_repo_id, repo_provider, status, created_at, updated_at, archived_at
		FROM projects
		WHERE slug = $1;
	`

	sqlProjectList = `
		SELECT id, name, slug, description, design_reference_url, git_repo_id, repo_provider, status, created_at, updated_at, archived_at
		FROM projects
		ORDER BY created_at DESC;
	`

	sqlProjectUpdate = `
		UPDATE projects
		SET name = $2,
		    slug = $3,
		    description = $4,
		    design_reference_url = $5,
		    git_repo_id = $6,
		    repo_provider = $7,
		    status = $8
		WHERE id = $1
		RETURNING created_at, updated_at, archived_at;
	`

	sqlProjectDelete = `DELETE FROM projects WHERE id = $1;`
)

func (r *ProjectRepository) Create(ctx context.Context, p *project.ProjectEntity) error {

	row := r.database.Pool.QueryRow(ctx, sqlProjectInsert,
		p.ID, p.Name, p.Slug, p.Description, p.DesignReferenceURL, p.GitRepoID, p.RepoProvider, p.Status,
	)
	if err := row.Scan(&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt); err != nil {
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: slug", constant.ErrUniqueConstraint)
		}
		return err
	}
	return nil
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*project.ProjectEntity, error) {
	var p project.ProjectEntity
	err := r.database.Pool.
		QueryRow(ctx, sqlProjectByID, id).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.DesignReferenceURL, &p.GitRepoID, &p.RepoProvider, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) FindBySlug(ctx context.Context, slug string) (*project.ProjectEntity, error) {
	var p project.ProjectEntity
	err := r.database.Pool.
		QueryRow(ctx, sqlProjectBySlug, slug).
		Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.DesignReferenceURL, &p.GitRepoID, &p.RepoProvider, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *ProjectRepository) List(ctx context.Context) ([]*project.ProjectEntity, error) {
	rows, err := r.database.Pool.Query(ctx, sqlProjectList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*project.ProjectEntity
	for rows.Next() {
		var p project.ProjectEntity
		if err := rows.Scan(&p.ID, &p.Name, &p.Slug, &p.Description, &p.DesignReferenceURL, &p.GitRepoID, &p.RepoProvider, &p.Status, &p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt); err != nil {
			return nil, err
		}
		out = append(out, &p)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *ProjectRepository) Update(ctx context.Context, p *project.ProjectEntity) error {
	row := r.database.Pool.QueryRow(ctx, sqlProjectUpdate,
		p.ID, p.Name, p.Slug, p.Description, p.DesignReferenceURL, p.GitRepoID, p.RepoProvider, p.Status,
	)
	if err := row.Scan(&p.CreatedAt, &p.UpdatedAt, &p.ArchivedAt); err != nil {
		if isNoRows(err) {
			return constant.ErrNotFound
		}
		if isUniqueViolation(err) {
			return fmt.Errorf("%w: slug", constant.ErrUniqueConstraint)
		}
		return err
	}
	return nil
}

func (r *ProjectRepository) Delete(ctx context.Context, id string) error {
	ct, err := r.database.Pool.Exec(ctx, sqlProjectDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return constant.ErrNotFound
	}
	return nil
}

func isNoRows(err error) bool {

	return err != nil && err.Error() == "no rows in result set"
}

func isUniqueViolation(err error) bool {

	type causer interface{ Error() string }
	if err == nil {
		return false
	}
	return contains(err.Error(), "duplicate key value violates unique constraint")
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool { return (len(s) > 0) && (index(s, sub) >= 0) })()
}

func index(s, sep string) int {

outer:
	for i := 0; i+len(sep) <= len(s); i++ {
		for j := 0; j < len(sep); j++ {
			if s[i+j] != sep[j] {
				continue outer
			}
		}
		return i
	}
	return -1
}
