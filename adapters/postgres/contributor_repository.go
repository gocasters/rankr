package postgres

import (
    "context"
    "database/sql"
    "time"

    "github.com/gocasters/rankr/domain/contributor"
)

type ContributorRepository struct {
    db *sql.DB
}

func NewContributorRepository(db *sql.DB) contributor.Repository {
    return &ContributorRepository{db: db}
}

func (r *ContributorRepository) Create(ctx context.Context, c *contributor.Contributor) error {
    query := `INSERT INTO contributors (id, username, email, display_name, avatar_url, github_id, is_active, created_at, updated_at) 
              VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
    _, err := r.db.ExecContext(ctx, query, c.ID, c.Username, c.Email, c.DisplayName, 
        c.AvatarURL, c.GitHubID, c.IsActive, c.CreatedAt, c.UpdatedAt)
    return err
}

func (r *ContributorRepository) GetByID(ctx context.Context, id string) (*contributor.Contributor, error) {
    query := `SELECT id, username, email, display_name, avatar_url, github_id, is_active, created_at, updated_at 
              FROM contributors WHERE id = $1`
    var c contributor.Contributor
    err := r.db.QueryRowContext(ctx, query, id).Scan(&c.ID, &c.Username, &c.Email, &c.DisplayName, 
        &c.AvatarURL, &c.GitHubID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, contributor.ErrContributorNotFound
    }
    return &c, err
}

func (r *ContributorRepository) GetByUsername(ctx context.Context, username string) (*contributor.Contributor, error) {
    query := `SELECT id, username, email, display_name, avatar_url, github_id, is_active, created_at, updated_at 
              FROM contributors WHERE username = $1`
    var c contributor.Contributor
    err := r.db.QueryRowContext(ctx, query, username).Scan(&c.ID, &c.Username, &c.Email, &c.DisplayName, 
        &c.AvatarURL, &c.GitHubID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, contributor.ErrContributorNotFound
    }
    return &c, err
}

func (r *ContributorRepository) GetByEmail(ctx context.Context, email string) (*contributor.Contributor, error) {
    query := `SELECT id, username, email, display_name, avatar_url, github_id, is_active, created_at, updated_at 
              FROM contributors WHERE email = $1`
    var c contributor.Contributor
    err := r.db.QueryRowContext(ctx, query, email).Scan(&c.ID, &c.Username, &c.Email, &c.DisplayName, 
        &c.AvatarURL, &c.GitHubID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
    if err == sql.ErrNoRows {
        return nil, contributor.ErrContributorNotFound
    }
    return &c, err
}

func (r *ContributorRepository) Update(ctx context.Context, c *contributor.Contributor) error {
    c.UpdatedAt = time.Now()
    query := `UPDATE contributors SET username = $2, email = $3, display_name = $4, avatar_url = $5, 
              github_id = $6, is_active = $7, updated_at = $8 WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, c.ID, c.Username, c.Email, c.DisplayName, 
        c.AvatarURL, c.GitHubID, c.IsActive, c.UpdatedAt)
    return err
}

func (r *ContributorRepository) Delete(ctx context.Context, id string) error {
    query := `DELETE FROM contributors WHERE id = $1`
    _, err := r.db.ExecContext(ctx, query, id)
    return err
}

func (r *ContributorRepository) List(ctx context.Context, limit, offset int) ([]*contributor.Contributor, error) {
    query := `SELECT id, username, email, display_name, avatar_url, github_id, is_active, created_at, updated_at 
              FROM contributors ORDER BY created_at DESC LIMIT $1 OFFSET $2`
    rows, err := r.db.QueryContext(ctx, query, limit, offset)
    if err != nil {
        return nil, err
    }
    defer rows.Close()

    var contributors []*contributor.Contributor
    for rows.Next() {
        var c contributor.Contributor
        err := rows.Scan(&c.ID, &c.Username, &c.Email, &c.DisplayName, 
            &c.AvatarURL, &c.GitHubID, &c.IsActive, &c.CreatedAt, &c.UpdatedAt)
        if err != nil {
            return nil, err
        }
        contributors = append(contributors, &c)
    }

    return contributors, nil
}

func (r *ContributorRepository) Exists(ctx context.Context, id string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM contributors WHERE id = $1)`
    err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
    return exists, err
}

func (r *ContributorRepository) ExistsByUsername(ctx context.Context, username string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM contributors WHERE username = $1)`
    err := r.db.QueryRowContext(ctx, query, username).Scan(&exists)
    return exists, err
}

func (r *ContributorRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM contributors WHERE email = $1)`
    err := r.db.QueryRowContext(ctx, query, email).Scan(&exists)
    return exists, err
}

func (r *ContributorRepository) Count(ctx context.Context) (int, error) {
    var count int
    query := `SELECT COUNT(*) FROM contributors`
    err := r.db.QueryRowContext(ctx, query).Scan(&count)
    return count, err
}
