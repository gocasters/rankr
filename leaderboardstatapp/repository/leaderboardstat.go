package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/database"
	types "github.com/gocasters/rankr/type"
	"github.com/jackc/pgx/v5"
	"time"
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
		scoreHistory = append(scoreHistory, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating score history: %v", err)
	}

	return scoreHistory, nil
}

func (repo LeaderboardstatRepo) StoreDailyContributorScores(ctx context.Context, scores []leaderboardstat.DailyContributorScore) error {
	if len(scores) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO scores (contributor_id, project_id, score, rank, earned_at, status)
		VALUES ($1, $2, $3, $4, $5, 0)
	`

	for _, score := range scores {
		batch.Queue(query,
			score.ContributorID,
			score.ProjectID,
			score.Score,
			score.Rank,
			time.Now(),
		)
	}

	results := repo.PostgreSQL.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to insert daily score at batch index %d: %w", i, err)
		}
	}

	return nil
}

func (repo LeaderboardstatRepo) GetPendingDailyScores(ctx context.Context) ([]leaderboardstat.DailyContributorScore, error) {
	query := `
		SELECT id, contributor_id, project_id, score, rank
		FROM scores
		WHERE status = 0
	`

	rows, err := repo.PostgreSQL.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending scores: %w", err)
	}
	defer rows.Close()

	var scores []leaderboardstat.DailyContributorScore

	for rows.Next() {
		var s leaderboardstat.DailyContributorScore

		if err := rows.Scan(
			&s.ID,
			&s.ContributorID,
			&s.ProjectID,
			&s.Score,
			&s.Rank,
		); err != nil {
			return nil, fmt.Errorf("failed to scan pending score: %w", err)
		}

		scores = append(scores, s)
	}

	return scores, nil
}

func (repo LeaderboardstatRepo) UpdateUserProjectScores(ctx context.Context, ups []leaderboardstat.UserProjectScore) error {
	if len(ups) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO user_project_scores (contributor_id, project_id, score, timeframe, time_value, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (contributor_id, project_id, timeframe, time_value)
		DO UPDATE SET 
			score = user_project_scores.score + EXCLUDED.score,
			updated_at = EXCLUDED.updated_at
	`

	for _, up := range ups {
		batch.Queue(query,
			up.ContributorID,
			up.ProjectID,
			up.Score,
			up.Timeframe,
			up.TimeValue,
			time.Now(),
		)
	}

	results := repo.PostgreSQL.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to update user_project_score at index %d: %w", i, err)
		}
	}

	return nil
}

func (repo LeaderboardstatRepo) UpdateGlobalScores(ctx context.Context, scores []leaderboardstat.UserProjectScore) error {
	if len(scores) == 0 {
		return nil
	}

	// Aggregate totals
	totals := make(map[types.ID]float64)
	for _, s := range scores {
		totals[s.ContributorID] += s.Score
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO user_project_scores (contributor_id, project_id, score, timeframe, time_value, updated_at)
		VALUES ($1, 0, $2, 'global', null, $3)
		ON CONFLICT (contributor_id, project_id, timeframe, time_value)
		DO UPDATE SET 
			score = user_project_scores.score + EXCLUDED.score,
			updated_at = EXCLUDED.updated_at
	`

	for contributorID, total := range totals {
		batch.Queue(query, contributorID, total, time.Now())
	}

	results := repo.PostgreSQL.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		if _, err := results.Exec(); err != nil {
			return fmt.Errorf("failed to update global score at index %d: %w", i, err)
		}
	}

	return nil
}

func (repo LeaderboardstatRepo) MarkDailyScoresAsProcessed(ctx context.Context, scoreIDs []types.ID) error {
	if len(scoreIDs) == 0 {
		return nil
	}

	query := `
		UPDATE scores 
		SET status = 1 
		WHERE id = ANY($1)
	`

	_, err := repo.PostgreSQL.Pool.Exec(ctx, query, scoreIDs)
	if err != nil {
		return fmt.Errorf("failed to mark scores as processed: %w", err)
	}

	return nil
}
