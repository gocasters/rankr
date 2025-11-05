package repository

import (
	"context"
	"time"

	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/pkg/database"
	types "github.com/gocasters/rankr/type"
	"github.com/jackc/pgx/v5"
)

type Config struct {
	CacheEnabled bool   `koanf:"cache_enabled"`
	CachePrefix  string `koanf:"cache_prefix"`
}

type grantRepository struct {
	db     *database.Database
	config Config
}

func NewAuthRepo(cfg Config, db *database.Database) auth.Repository {
	return NewGrantRepository(cfg, db)
}

func NewGrantRepository(cfg Config, db *database.Database) auth.Repository {
	return &grantRepository{
		db:     db,
		config: cfg,
	}
}

func (r *grantRepository) Create(ctx context.Context, g auth.Grant) (types.ID, error) {
	const query = `
		INSERT INTO grants (subject, object, action, field, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	var rawID uint64
	now := time.Now()
	if g.CreatedAt.IsZero() {
		g.CreatedAt = now
	}
	if g.UpdatedAt.IsZero() {
		g.UpdatedAt = now
	}

	err := r.db.Pool.QueryRow(ctx, query,
		g.Subject,
		g.Object,
		g.Action,
		g.Field,
		g.CreatedAt,
		g.UpdatedAt,
	).Scan(&rawID)
	return types.ID(rawID), err
}

func (r *grantRepository) Update(ctx context.Context, g auth.Grant) error {
	const query = `
		UPDATE grants
		SET subject = $1,
		    object = $2,
		    action = $3,
		    field = $4,
		    updated_at = $5
		WHERE id = $6
	`

	if g.UpdatedAt.IsZero() {
		g.UpdatedAt = time.Now()
	}

	_, err := r.db.Pool.Exec(ctx, query,
		g.Subject,
		g.Object,
		g.Action,
		g.Field,
		g.UpdatedAt,
		uint64(g.ID),
	)
	return err
}

func (r *grantRepository) Get(ctx context.Context, id types.ID) (auth.Grant, error) {
	const query = `
		SELECT id, subject, object, action, field, created_at, updated_at
		FROM grants
		WHERE id = $1
	`

	var (
		rawID  uint64
		fields []string
		g      auth.Grant
	)

	err := r.db.Pool.QueryRow(ctx, query, uint64(id)).Scan(
		&rawID,
		&g.Subject,
		&g.Object,
		&g.Action,
		&fields,
		&g.CreatedAt,
		&g.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return auth.Grant{}, auth.ErrGrantNotFound
		}
		return auth.Grant{}, err
	}

	g.ID = types.ID(rawID)
	g.Field = fields
	return g, nil
}

func (r *grantRepository) Delete(ctx context.Context, id types.ID) error {
	const query = `DELETE FROM grants WHERE id = $1`

	tag, err := r.db.Pool.Exec(ctx, query, uint64(id))
	if err != nil {
		return err
	}

	if tag.RowsAffected() == 0 {
		return auth.ErrGrantNotFound
	}
	return nil
}

func (r *grantRepository) List(ctx context.Context) ([]auth.Grant, error) {
	const query = `
		SELECT id, subject, object, action, field, created_at, updated_at
		FROM grants
		ORDER BY updated_at DESC
	`

	rows, err := r.db.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var grants []auth.Grant
	for rows.Next() {
		var (
			rawID  uint64
			fields []string
			g      auth.Grant
		)

		if err := rows.Scan(
			&rawID,
			&g.Subject,
			&g.Object,
			&g.Action,
			&fields,
			&g.CreatedAt,
			&g.UpdatedAt,
		); err != nil {
			return nil, err
		}

		g.ID = types.ID(rawID)
		g.Field = fields
		grants = append(grants, g)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return grants, nil
}
