package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/database"
	"log/slog"
)

type Config struct {
	CacheEnabled bool   `koanf:"cache_enabled"`
	CachePrefix  string `koanf:"cache_prefix"`
}

type ContributorRepo struct {
	Config     Config
	Logger     *slog.Logger
	PostgreSQL *database.Database
	Cache      *redis.Adapter
}

func NewContributorRepo(config Config, db *database.Database, logger *slog.Logger) contributor.Repository {
	return &ContributorRepo{
		Config:     config,
		Logger:     logger,
		PostgreSQL: db,
	}
}

func (repo ContributorRepo) GetContributorByID(ctx context.Context, ID types.ID) (*contributor.Contributor, error) {
	query := "SELECT id, github_id, github_username, email, is_verified, two_factor_enabled, privacy_mode, display_name, profile_image, bio, created_at FROM contributors WHERE id=$1"
	row := repo.PostgreSQL.Pool.QueryRow(ctx, query, ID)

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
			return nil, fmt.Errorf("result with id %d not found", ID)
		}
		return nil, fmt.Errorf("error retrieving contributor with id: %d, error: %v", ID, err)
	}

	return &contrib, nil
}
