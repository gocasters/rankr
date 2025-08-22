package repository

import (
	"context"

	"github.com/gocasters/rankr/projectsapp/constants"
	"github.com/gocasters/rankr/projectsapp/helpers"
	"github.com/gocasters/rankr/projectsapp/types"
	"github.com/jackc/pgx/v5/pgxpool"
)

type VcsRepoRepositoryPG struct {
	pool *pgxpool.Pool
}

func NewVcsRepoRepositoryPG(pool *pgxpool.Pool) *VcsRepoRepositoryPG {
	return &VcsRepoRepositoryPG{pool: pool}
}

const (
	sqlVcsInsert = `
		INSERT INTO vcs_repos (
			id, project_id, provider, provider_repo_id, owner, name, remote_url,
			default_branch, visibility, installation_id, last_synced_at
		)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING created_at, updated_at;
	`

	sqlVcsByID = `
		SELECT id, project_id, provider, provider_repo_id, owner, name, remote_url,
		       default_branch, visibility, installation_id, last_synced_at,
		       created_at, updated_at
		FROM vcs_repos
		WHERE id = $1;
	`

	sqlVcsByProject = `
		SELECT id, project_id, provider, provider_repo_id, owner, name, remote_url,
		       default_branch, visibility, installation_id, last_synced_at,
		       created_at, updated_at
		FROM vcs_repos
		WHERE project_id = $1
		ORDER BY created_at DESC;
	`

	sqlVcsByProviderID = `
		SELECT id, project_id, provider, provider_repo_id, owner, name, remote_url,
		       default_branch, visibility, installation_id, last_synced_at,
		       created_at, updated_at
		FROM vcs_repos
		WHERE provider = $1 AND provider_repo_id = $2 AND project_id = $3;
	`

	sqlVcsUpdate = `
		UPDATE vcs_repos
		SET owner = $2,
		    name = $3,
		    remote_url = $4,
		    default_branch = $5,
		    visibility = $6,
		    installation_id = $7,
		    last_synced_at = $8
		WHERE id = $1
		RETURNING created_at, updated_at;
	`

	sqlVcsDelete = `DELETE FROM vcs_repos WHERE id = $1;`

	sqlVcsRepositoriesList = `
		SELECT id, project_id, provider, provider_repo_id, owner, name, remote_url,
		       default_branch, visibility, installation_id, last_synced_at,
		       created_at, updated_at
		FROM vcs_repos
		ORDER BY created_at DESC;`
)

func (r *VcsRepoRepositoryPG) Create(ctx context.Context, v *types.VcsRepoEntity) error {
	row := r.pool.QueryRow(ctx, sqlVcsInsert,
		v.ID, v.ProjectID, v.Provider, v.ProviderRepoID, v.Owner, v.Name, v.RemoteURL,
		v.DefaultBranch, v.Visibility, v.InstallationID, v.LastSyncedAt,
	)
	if err := row.Scan(&v.CreatedAt, &v.UpdatedAt); err != nil {
		if helpers.IsUniqueViolation(err) {
			//unique(project_id, provider, provider_repo_id)
			return constants.ErrUniqueConstraint
		}
		return err
	}
	return nil
}

func (r *VcsRepoRepositoryPG) FindByID(ctx context.Context, id string) (*types.VcsRepoEntity, error) {
	var v types.VcsRepoEntity
	err := r.pool.QueryRow(ctx, sqlVcsByID, id).
		Scan(&v.ID, &v.ProjectID, &v.Provider, &v.ProviderRepoID, &v.Owner, &v.Name, &v.RemoteURL,
			&v.DefaultBranch, &v.Visibility, &v.InstallationID, &v.LastSyncedAt,
			&v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if helpers.IsNoRows(err) {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *VcsRepoRepositoryPG) FindByProjectID(ctx context.Context, projectID string) ([]*types.VcsRepoEntity, error) {
	rows, err := r.pool.Query(ctx, sqlVcsByProject, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*types.VcsRepoEntity
	for rows.Next() {
		var v types.VcsRepoEntity
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Provider, &v.ProviderRepoID, &v.Owner, &v.Name, &v.RemoteURL,
			&v.DefaultBranch, &v.Visibility, &v.InstallationID, &v.LastSyncedAt,
			&v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil
}

func (r *VcsRepoRepositoryPG) FindByProviderID(ctx context.Context, provider types.VcsProvider, providerRepoID string, projectID string) (*types.VcsRepoEntity, error) {
	var v types.VcsRepoEntity
	err := r.pool.QueryRow(ctx, sqlVcsByProviderID, provider, providerRepoID, projectID).
		Scan(&v.ID, &v.ProjectID, &v.Provider, &v.ProviderRepoID, &v.Owner, &v.Name, &v.RemoteURL,
			&v.DefaultBranch, &v.Visibility, &v.InstallationID, &v.LastSyncedAt,
			&v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if helpers.IsNoRows(err) {
			return nil, constants.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *VcsRepoRepositoryPG) Update(ctx context.Context, v *types.VcsRepoEntity) error {
	row := r.pool.QueryRow(ctx, sqlVcsUpdate,
		v.ID, v.Owner, v.Name, v.RemoteURL, v.DefaultBranch, v.Visibility, v.InstallationID, v.LastSyncedAt,
	)
	if err := row.Scan(&v.CreatedAt, &v.UpdatedAt); err != nil {
		if helpers.IsNoRows(err) {
			return constants.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *VcsRepoRepositoryPG) Delete(ctx context.Context, id string) error {
	ct, err := r.pool.Exec(ctx, sqlVcsDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return constants.ErrNotFound
	}
	return nil
}

func (r *VcsRepoRepositoryPG) List(ctx context.Context) ([]*types.VcsRepoEntity, error) {
	rows, err := r.pool.Query(ctx, sqlVcsRepositoriesList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*types.VcsRepoEntity
	for rows.Next() {
		var v types.VcsRepoEntity
		if err := rows.Scan(&v.ID, &v.ProjectID, &v.Provider, &v.ProviderRepoID, &v.Owner, &v.Name, &v.RemoteURL,
			&v.DefaultBranch, &v.Visibility, &v.InstallationID, &v.LastSyncedAt,
			&v.CreatedAt, &v.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, &v)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return out, nil

}
