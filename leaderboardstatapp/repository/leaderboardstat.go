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

func (repo LeaderboardstatRepo) GetContributorTotalRank(ctx context.Context, ID types.ID) (uint, error) {
	query := "SELECT RANK() OVER (ORDER BY SUM(score) DESC) as global_rank FROM scores GROUP BY contributor_id HAVING contributor_id = $1;"
	row := repo.PostgreSQL.Pool.QueryRow(ctx, query, ID)

	var rank uint
	err := row.Scan(&rank)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, nil
		}

		return 0, fmt.Errorf("error retrieving Rank for contributor id %d: %v", ID, err)
	}

	return rank, nil
}

func (repo LeaderboardstatRepo) GetContributorProjectScores(ctx context.Context, contributorID types.ID) (map[types.ID]float64, error) {
	query := "SELECT project_id, SUM(score) as total_score FROM scores WHERE contributor_id = $1 GROUP BY project_id"
	rows, err := repo.PostgreSQL.Pool.Query(ctx, query, contributorID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving project scores for contributor id %d: %v", contributorID, err)
	}
	defer rows.Close()

	projectScores := make(map[types.ID]float64)

	for rows.Next() {
		var projectID types.ID
		var totalScore float64

		err := rows.Scan(&projectID, &totalScore)
		if err != nil {
			return nil, fmt.Errorf("error scanning project scores for contributor id %d: %v", contributorID, err)
		}

		projectScores[projectID] = totalScore
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating project scores for contributor id %d: %v", contributorID, err)
	}

	return projectScores, nil
}

func (repo LeaderboardstatRepo) GetContributorScoreHistory(ctx context.Context, contributorID types.ID) ([]leaderboardstat.ScoreRecord, error) {
	query := "SELECT id, contributor_id, project_id, activity, score, earned_at FROM scores WHERE contributor_id = $1 ORDER BY project_id DESC;"
	rows, err := repo.PostgreSQL.Pool.Query(ctx, query, contributorID)
	if err != nil {
		return nil, fmt.Errorf("error retrieving project scores for contributor id %d: %v", contributorID, err)
	}
	defer rows.Close()

	var scoreHistory []leaderboardstat.ScoreRecord
	for rows.Next() {
		var record leaderboardstat.ScoreRecord

		err := rows.Scan(&record.ID, &record.ContributorID, &record.ProjectID, &record.Activity, &record.Score, &record.EarnedAt)
		if err != nil {
			return nil, fmt.Errorf("error scanning score history: %v", err)
		}
		fmt.Println("...8...>>>", record)
		scoreHistory = append(scoreHistory, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating score history: %v", err)
	}

	return scoreHistory, nil
}
