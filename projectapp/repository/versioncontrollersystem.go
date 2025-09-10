package repository

import (
	"context"

	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/gocasters/rankr/projectapp/service/versioncontrollersystemproject"
)

type VersionControllerSystemProjectRepository struct {
	database *database.Database
}

func NewVersionControllerSystemProjectRepository(database *database.Database) versioncontrollersystemproject.Repository {
	return &VersionControllerSystemProjectRepository{database: database}
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

func (r *VersionControllerSystemProjectRepository) Create(ctx context.Context, versionControllerSystemProjectEntity *versioncontrollersystemproject.VersionControllerSystemProjectEntity) (*versioncontrollersystemproject.VersionControllerSystemProjectEntity, error) {
	row := r.database.Pool.QueryRow(ctx, sqlVcsInsert,
		versionControllerSystemProjectEntity.ID, versionControllerSystemProjectEntity.ProjectID, versionControllerSystemProjectEntity.Provider, versionControllerSystemProjectEntity.ProviderRepoID, versionControllerSystemProjectEntity.Owner, versionControllerSystemProjectEntity.Name, versionControllerSystemProjectEntity.RemoteURL,
		versionControllerSystemProjectEntity.DefaultBranch, versionControllerSystemProjectEntity.Visibility, versionControllerSystemProjectEntity.InstallationID, versionControllerSystemProjectEntity.LastSyncedAt,
	)

	if err := row.Scan(&versionControllerSystemProjectEntity.CreatedAt, &versionControllerSystemProjectEntity.UpdatedAt); err != nil {
		if isUniqueViolation(err) {
			return nil, constant.ErrUniqueConstraint
		}
		return nil, err
	}
	return versionControllerSystemProjectEntity, nil
}

func (r *VersionControllerSystemProjectRepository) FindByID(ctx context.Context, id string) (*versioncontrollersystemproject.VersionControllerSystemProjectEntity, error) {
	var v versioncontrollersystemproject.VersionControllerSystemProjectEntity
	err := r.database.Pool.QueryRow(ctx, sqlVcsByID, id).
		Scan(&v.ID, &v.ProjectID, &v.Provider, &v.ProviderRepoID, &v.Owner, &v.Name, &v.RemoteURL,
			&v.DefaultBranch, &v.Visibility, &v.InstallationID, &v.LastSyncedAt,
			&v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *VersionControllerSystemProjectRepository) FindByProjectID(ctx context.Context, projectID string) ([]*versioncontrollersystemproject.VersionControllerSystemProjectEntity, error) {
	rows, err := r.database.Pool.Query(ctx, sqlVcsByProject, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*versioncontrollersystemproject.VersionControllerSystemProjectEntity
	for rows.Next() {
		var v versioncontrollersystemproject.VersionControllerSystemProjectEntity
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

func (r *VersionControllerSystemProjectRepository) FindByProviderID(ctx context.Context, provider constant.VcsProvider, providerRepoID string, projectID string) (*versioncontrollersystemproject.VersionControllerSystemProjectEntity, error) {
	var v versioncontrollersystemproject.VersionControllerSystemProjectEntity
	err := r.database.Pool.QueryRow(ctx, sqlVcsByProviderID, provider, providerRepoID, projectID).
		Scan(&v.ID, &v.ProjectID, &v.Provider, &v.ProviderRepoID, &v.Owner, &v.Name, &v.RemoteURL,
			&v.DefaultBranch, &v.Visibility, &v.InstallationID, &v.LastSyncedAt,
			&v.CreatedAt, &v.UpdatedAt)
	if err != nil {
		if isNoRows(err) {
			return nil, constant.ErrNotFound
		}
		return nil, err
	}
	return &v, nil
}

func (r *VersionControllerSystemProjectRepository) Update(ctx context.Context, v *versioncontrollersystemproject.VersionControllerSystemProjectEntity) error {
	row := r.database.Pool.QueryRow(ctx, sqlVcsUpdate,
		v.ID, v.Owner, v.Name, v.RemoteURL, v.DefaultBranch, v.Visibility, v.InstallationID, v.LastSyncedAt,
	)
	if err := row.Scan(&v.CreatedAt, &v.UpdatedAt); err != nil {
		if isNoRows(err) {
			return constant.ErrNotFound
		}
		return err
	}
	return nil
}

func (r *VersionControllerSystemProjectRepository) Delete(ctx context.Context, id string) error {
	ct, err := r.database.Pool.Exec(ctx, sqlVcsDelete, id)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return constant.ErrNotFound
	}
	return nil
}

func (r *VersionControllerSystemProjectRepository) List(ctx context.Context) ([]*versioncontrollersystemproject.VersionControllerSystemProjectEntity, error) {
	rows, err := r.database.Pool.Query(ctx, sqlVcsRepositoriesList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []*versioncontrollersystemproject.VersionControllerSystemProjectEntity
	for rows.Next() {
		var v versioncontrollersystemproject.VersionControllerSystemProjectEntity
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
