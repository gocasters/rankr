package repository

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/database"
	types "github.com/gocasters/rankr/type"
	"github.com/jackc/pgx/v5"
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
		INSERT INTO daily_contributor_scores 
		(contributor_id, user_id, daily_score, rank, timeframe, calculated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (contributor_id, timeframe, calculated_at::date) DO UPDATE 
		SET daily_score = EXCLUDED.daily_score,
			rank = EXCLUDED.rank,
			user_id = EXCLUDED.user_id,
			status = 0,
			calculated_at = EXCLUDED.calculated_at
	`

	for _, score := range scores {
		batch.Queue(query,
			score.ContributorID,
			score.UserID,
			score.DailyScore,
			score.Rank,
			score.Timeframe,
			score.CalculatedAt,
		)
	}

	results := repo.PostgreSQL.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to insert daily score at index %d: %w", i, err)
		}
	}

	return nil
}

func (repo LeaderboardstatRepo) GetPendingDailyScores(ctx context.Context) ([]leaderboardstat.DailyContributorScore, error) {
	query := `
		SELECT id, contributor_id, user_id, daily_score, rank, timeframe, calculated_at
		FROM daily_contributor_scores 
		WHERE status = 0
		ORDER BY calculated_at
	`

	rows, err := repo.PostgreSQL.Pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query pending scores: %w", err)
	}
	defer rows.Close()

	var scores []leaderboardstat.DailyContributorScore
	for rows.Next() {
		var score leaderboardstat.DailyContributorScore
		err := rows.Scan(
			&score.ID,
			&score.ContributorID,
			&score.UserID,
			&score.DailyScore,
			&score.Rank,
			&score.Timeframe,
			&score.CalculatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan score: %w", err)
		}
		scores = append(scores, score)
	}

	return scores, nil
}
func (repo LeaderboardstatRepo) UpdateUserScores(ctx context.Context, userScores []leaderboardstat.UserScore) error {
	if len(userScores) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO user_scores (user_id, score) 
		VALUES ($1, $2)
		ON CONFLICT (user_id) 
		DO UPDATE SET score = user_scores.score + EXCLUDED.score
	`

	for _, userScore := range userScores {
		batch.Queue(query, userScore.UserID, userScore.Score)
	}

	results := repo.PostgreSQL.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to update user score at index %d: %w", i, err)
		}
	}

	return nil
}
func (repo LeaderboardstatRepo) UpdateProjectScores(ctx context.Context, projectScores []leaderboardstat.ProjectScore) error {
	if len(projectScores) == 0 {
		return nil
	}

	batch := &pgx.Batch{}

	query := `
		INSERT INTO project_scores (user_id, project_id, score) 
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, project_id) 
		DO UPDATE SET score = project_scores.score + EXCLUDED.score
	`

	for _, projectScore := range projectScores {
		batch.Queue(query, projectScore.UserID, projectScore.ProjectID, projectScore.Score)
	}

	results := repo.PostgreSQL.Pool.SendBatch(ctx, batch)
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return fmt.Errorf("failed to update project score at index %d: %w", i, err)
		}
	}

	return nil
}
func (repo LeaderboardstatRepo) MarkDailyScoresAsProcessed(ctx context.Context, scoreIDs []types.ID) error {
	if len(scoreIDs) == 0 {
		return nil
	}

	query := `
		UPDATE daily_contributor_scores 
		SET status = 1 
		WHERE id = ANY($1)
	`

	_, err := repo.PostgreSQL.Pool.Exec(ctx, query, scoreIDs)
	if err != nil {
		return fmt.Errorf("failed to mark scores as processed: %w", err)
	}

	return nil
}
