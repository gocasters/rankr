package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/database"
	types "github.com/gocasters/rankr/type"
	"github.com/jackc/pgx/v5"
)

type Config struct {
	CacheEnabled bool   `koanf:"cache_enabled"`
	CachePrefix  string `koanf:"cache_prefix"`
}

type ContributorRepo struct {
	Config      Config
	Logger      *slog.Logger
	PostgresSQL *database.Database
	Cache       *redis.Adapter
}

func NewContributorRepo(config Config, db *database.Database, logger *slog.Logger) contributor.Repository {
	return &ContributorRepo{
		Config:      config,
		Logger:      logger,
		PostgresSQL: db,
	}
}

func (repo ContributorRepo) GetContributorByID(ctx context.Context, id types.ID) (*contributor.Contributor, error) {
	query := "SELECT id, github_id, github_username, email, password, is_verified, two_factor_enabled, privacy_mode, display_name, profile_image, bio, created_at, updated_at FROM contributors WHERE id=$1"
	row := repo.PostgresSQL.Pool.QueryRow(ctx, query, id)

	var contrib contributor.Contributor
	var githubID sql.NullInt64
	err := row.Scan(
		&contrib.ID,
		&githubID,
		&contrib.GitHubUsername,
		&contrib.Email,
		&contrib.Password,
		&contrib.IsVerified,
		&contrib.TwoFactor,
		&contrib.PrivacyMode,
		&contrib.DisplayName,
		&contrib.ProfileImage,
		&contrib.Bio,
		&contrib.CreatedAt,
		&contrib.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("result with id %d not found", id)
		}
		return nil, fmt.Errorf("error retrieving contributor with id: %d, error: %v", id, err)
	}

	if githubID.Valid {
		contrib.GitHubID = githubID.Int64
	}
	return &contrib, nil
}

func (repo ContributorRepo) GetContributorByGitHubUsername(ctx context.Context, username string) (*contributor.Contributor, error) {
	query := "SELECT id, github_id, github_username, email, password, is_verified, two_factor_enabled, privacy_mode, display_name, profile_image, bio, created_at, updated_at FROM contributors WHERE github_username=$1"
	row := repo.PostgresSQL.Pool.QueryRow(ctx, query, username)

	var contrib contributor.Contributor
	var githubID sql.NullInt64
	err := row.Scan(
		&contrib.ID,
		&githubID,
		&contrib.GitHubUsername,
		&contrib.Email,
		&contrib.Password,
		&contrib.IsVerified,
		&contrib.TwoFactor,
		&contrib.PrivacyMode,
		&contrib.DisplayName,
		&contrib.ProfileImage,
		&contrib.Bio,
		&contrib.CreatedAt,
		&contrib.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("result with username %s not found", username)
		}
		return nil, fmt.Errorf("error retrieving contributor with username: %s, error: %v", username, err)
	}

	if githubID.Valid {
		contrib.GitHubID = githubID.Int64
	}

	return &contrib, nil
}
func (repo ContributorRepo) CreateContributor(ctx context.Context, contributor contributor.Contributor) (*contributor.Contributor, error) {
	query := `
    	INSERT INTO contributors (github_id, github_username, email, password, privacy_mode, display_name, profile_image, bio, created_at, updated_at)
    	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
    	RETURNING id, created_at, updated_at;
    `

	createdAt := contributor.CreatedAt
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	updatedAt := contributor.UpdatedAt
	if updatedAt.IsZero() {
		updatedAt = createdAt
	}

	var id int64
	err := repo.PostgresSQL.Pool.QueryRow(ctx, query,
		nullInt64(contributor.GitHubID),
		contributor.GitHubUsername,
		contributor.Email,
		contributor.Password,
		contributor.PrivacyMode,
		contributor.DisplayName,
		contributor.ProfileImage,
		contributor.Bio,
		createdAt,
		updatedAt,
	).Scan(&id, &contributor.CreatedAt, &contributor.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create contributor: %w", err)
	}

	contributor.ID = id
	return &contributor, nil
}
func (repo ContributorRepo) UpdateProfileContributor(ctx context.Context, contri contributor.Contributor) (*contributor.Contributor, error) {
	var updated contributor.Contributor
	var githubID sql.NullInt64

	query := `
		UPDATE contributors
		SET github_id=$1,
		    github_username=$2,
		    display_name=$3,
		    profile_image=$4,
		    bio=$5,
		    privacy_mode=$6,
		    email=$7,
		    updated_at=NOW()
		WHERE id=$8
		RETURNING id, github_id, github_username, display_name, profile_image, bio, privacy_mode, email, created_at, updated_at;
	`

	err := repo.PostgresSQL.Pool.QueryRow(ctx, query,
		nullInt64(contri.GitHubID),
		contri.GitHubUsername,
		contri.DisplayName,
		contri.ProfileImage,
		contri.Bio,
		contri.PrivacyMode,
		contri.Email,
		contri.ID,
	).Scan(
		&updated.ID,
		&githubID,
		&updated.GitHubUsername,
		&updated.DisplayName,
		&updated.ProfileImage,
		&updated.Bio,
		&updated.PrivacyMode,
		&updated.Email,
		&updated.CreatedAt,
		&updated.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("no contributor found with id %d", contri.ID)
		}

		return nil, fmt.Errorf("failed to update contributor profile: %w", err)
	}

	if githubID.Valid {
		updated.GitHubID = githubID.Int64
	}

	return &updated, nil
}

func (repo ContributorRepo) UpdatePassword(ctx context.Context, id types.ID, hashedPassword string) error {
	commandTag, err := repo.PostgresSQL.Pool.Exec(ctx,
		"UPDATE contributors SET password=$1, updated_at=NOW() WHERE id=$2",
		hashedPassword, id,
	)
	if err != nil {
		return fmt.Errorf("failed to update password for contributor %d: %w", id, err)
	}

	if commandTag.RowsAffected() == 0 {
		return fmt.Errorf("no contributor found with id %d", id)
	}

	return nil
}

func nullInt64(v int64) sql.NullInt64 {
	if v == 0 {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: v, Valid: true}
}
