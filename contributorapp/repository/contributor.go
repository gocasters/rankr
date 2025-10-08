package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gocasters/rankr/adapter/redis"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/database"
	"github.com/gocasters/rankr/type"
	"log/slog"
	"strings"
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

func (repo *ContributorRepo) GetContributorByID(ctx context.Context, ID types.ID) (*contributor.Contributor, error) {
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
func (repo *ContributorRepo) CreateContributor(ctx context.Context, contributor contributor.Contributor) (*contributor.Contributor, error) {
	query := `
    	INSERT INTO contributors (github_id, github_username, email , privacy_mode, display_name, profile_image, bio, created_at)
    	VALUES ($1, $2, $3, $4 ,$5, $6, $7, $8)
    	RETURNING id;
    `

	var id int64
	err := repo.PostgreSQL.Pool.QueryRow(ctx, query,
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

func (repo *ContributorRepo) UpdateProfileContributor(ctx context.Context, contri contributor.Contributor) (*contributor.Contributor, error) {
	var updateContributor contributor.Contributor
	updates := make(map[string]interface{})

	if contri.GitHubID != 0 {
		updates["github_id"] = contri.GitHubID
	}

	if contri.GitHubUsername != "" {
		updates["github_username"] = contri.GitHubUsername
	}

	if contri.DisplayName != nil {
		updates["display_name"] = *contri.DisplayName
	}

	if contri.ProfileImage != nil {
		updates["profile_image"] = *contri.ProfileImage
	}

	if contri.Bio != nil {
		updates["bio"] = *contri.Bio
	}

	if contri.PrivacyMode != "" {
		updates["privacy_mode"] = contri.PrivacyMode
	}

	if len(updates) == 0 {
		return &contri, nil
	}

	sets := make([]string, 0, len(updates))
	args := make([]interface{}, 0, len(updates))
	i := 1
	for key, value := range updates {
		sets = append(sets, fmt.Sprintf("%s=$%d", key, i))
		args = append(args, value)
		i++
	}

	query := fmt.Sprintf(
		"UPDATE contributors SET %s WHERE id = $%d RETURNING id, github_id, github_username, display_name, profile_image, bio, privacy_mode, email, is_verified, two_factor_enabled, created_at, updated_at;",
		strings.Join(sets, ", "), i)

	args = append(args, contri.ID)

	if err := repo.PostgreSQL.Pool.QueryRow(ctx, query, args...).Scan(
		&updateContributor.ID,
		&updateContributor.GitHubID,
		&updateContributor.GitHubUsername,
		&updateContributor.DisplayName,
		&updateContributor.ProfileImage,
		&updateContributor.Bio,
		&updateContributor.PrivacyMode,
		&updateContributor.Email,
		&updateContributor.IsVerified,
		&updateContributor.TwoFactor,
		&updateContributor.CreatedAt,
		&updateContributor.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("failed update profile: %v", err)
	}

	return &updateContributor, nil
}
