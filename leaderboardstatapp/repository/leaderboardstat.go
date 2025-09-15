package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/database"
	types "github.com/gocasters/rankr/type"
)

type Config struct {
}
type LeaderboardstatRepo struct {
	Config     Config
	PostgreSQL *database.Database
}

func NewLeaderboardstatRepo(config Config, db *database.Database) leaderboardstat.Repository {
	return &LeaderboardstatRepo{
		Config:     config,
		PostgreSQL: db,
	}
}

func (repo LeaderboardstatRepo) GetContributorTotalScore(ctx context.Context, ID types.ID) (float64, error) {
	query := "SELECT COALESCE(SUM(score), 0) FROM scores WHERE contributor_id = $1"
	row := repo.PostgreSQL.Pool.QueryRow(ctx, query, ID)

	var totalScore float64
	err := row.Scan(&totalScore)
	if err != nil {
		if err == sql.ErrNoRows {
			// No scores found for this contributor, return 0
			return 0, nil
		}

		return 0, fmt.Errorf("error retrieving total score for contributor id %d: %v", ID, err)
	}

	return totalScore, nil
}
