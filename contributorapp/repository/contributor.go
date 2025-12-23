package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/database"
	types "github.com/gocasters/rankr/type"
	"github.com/jackc/pgx/v5"
	"log/slog"
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
	query := "SELECT id, github_id, github_username, email, is_verified, two_factor_enabled, privacy_mode, display_name, profile_image, bio, created_at FROM contributors WHERE id=$1"
	row := repo.PostgresSQL.Pool.QueryRow(ctx, query, id)

	var contrib contributor.Contributor
	err := row.Scan(
		&contrib.ID,
		&contrib.GitHubID,
		&contrib.GitHubUsername,
		&contrib.Email,
		&contrib.IsVerified,
		&contrib.TwoFactor,
		&contrib.PrivacyMode,
		&contrib.DisplayName,
		&contrib.ProfileImage,
		&contrib.Bio,
		&contrib.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("result with id %d not found", id)
		}
		return nil, fmt.Errorf("error retrieving contributor with id: %d, error: %v", id, err)
	}

	return &contrib, nil
}
func (repo ContributorRepo) CreateContributor(ctx context.Context, contributor contributor.Contributor) (*contributor.Contributor, error) {
	query := `
    	INSERT INTO contributors (github_id, github_username, email , privacy_mode, display_name, profile_image, bio, created_at)
    	VALUES ($1, $2, $3, $4 ,$5, $6, $7, $8)
    	RETURNING id;
    `

	var id int64
	err := repo.PostgresSQL.Pool.QueryRow(ctx, query,
		contributor.GitHubID,
		contributor.GitHubUsername,
		contributor.Email,
		contributor.PrivacyMode,
		contributor.DisplayName,
		contributor.ProfileImage,
		contributor.Bio,
		contributor.CreatedAt,
	).Scan(&id)

	if err != nil {
		return nil, fmt.Errorf("failed to create contributor: %w", err)
	}

	contributor.ID = id
	return &contributor, nil
}
func (repo ContributorRepo) UpdateProfileContributor(ctx context.Context, contri contributor.Contributor) (*contributor.Contributor, error) {
	var updated contributor.Contributor

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
		contri.GitHubID,
		contri.GitHubUsername,
		contri.DisplayName,
		contri.ProfileImage,
		contri.Bio,
		contri.PrivacyMode,
		contri.Email,
		contri.ID,
	).Scan(
		&updated.ID,
		&updated.GitHubID,
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

	return &updated, nil
}

func (repo ContributorRepo) GetContributorByGitHubUsername(ctx context.Context, githubUsername string) (int64, bool, error) {
	query := `SELECT github_username FROM contributors WHERE github_username=$1 RETURNING id;`

	var id int64

	err := repo.PostgresSQL.Pool.QueryRow(ctx, query, githubUsername).Scan(&id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, false, nil
		}

		return 0, false, fmt.Errorf("failed to check contributor by github_username: %w", err)
	}

	return id, true, nil
}

func (repo ContributorRepo) FindByVCSUsernames(ctx context.Context, provider contributor.VcsProvider, usernames []string) ([]*contributor.Contributor, error) {
	if len(usernames) == 0 {
		return []*contributor.Contributor{}, nil
	}

	query := `
		SELECT id, github_id, github_username, COALESCE(email, ''), is_verified, two_factor_enabled,
		       privacy_mode, COALESCE(display_name, ''), COALESCE(profile_image, ''), COALESCE(bio, ''), created_at
		FROM contributors
		WHERE github_username = ANY($1)
	`

	rows, err := repo.PostgresSQL.Pool.Query(ctx, query, usernames)
	if err != nil {
		return nil, fmt.Errorf("failed to find contributors by usernames: %w", err)
	}
	defer rows.Close()

	var contributors []*contributor.Contributor
	for rows.Next() {
		var c contributor.Contributor
		err := rows.Scan(
			&c.ID,
			&c.GitHubID,
			&c.GitHubUsername,
			&c.Email,
			&c.IsVerified,
			&c.TwoFactor,
			&c.PrivacyMode,
			&c.DisplayName,
			&c.ProfileImage,
			&c.Bio,
			&c.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan contributor: %w", err)
		}
		contributors = append(contributors, &c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating contributors: %w", err)
	}

	return contributors, nil
}
