package leaderboardstat

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	contributorclient "github.com/gocasters/rankr/adapter/contributor"
	lbscoringclient "github.com/gocasters/rankr/adapter/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The service calls logger.L(), which panics if the logger is not set up.
func TestMain(m *testing.M) {
	if err := logger.Init(logger.Config{Level: "error"}); err != nil {
		panic(err)
	}
	os.Exit(m.Run())
}

type contributorStatsRepo struct {
	Repository
	history []ScoreRecord
}

func (r contributorStatsRepo) GetContributorTotalScore(context.Context, types.ID) (float64, error) {
	return 42, nil
}
func (r contributorStatsRepo) GetContributorTotalRank(context.Context, types.ID) (uint, error) {
	return 3, nil
}
func (r contributorStatsRepo) GetContributorProjectScores(context.Context, types.ID) (map[types.ID]float64, error) {
	return map[types.ID]float64{types.ID(8): 12}, nil
}
func (r contributorStatsRepo) GetContributorScoreHistory(context.Context, types.ID) ([]ScoreRecord, error) {
	return r.history, nil
}

type contributorStatsContributor struct {
	requestedID types.ID
}

func (c *contributorStatsContributor) GetContributorInfo(_ context.Context, id types.ID) (contributorclient.InfoResponse, error) {
	c.requestedID = id
	return contributorclient.InfoResponse{GitHubID: 987654}, nil
}

type contributorStatsScoring struct {
	LeaderboardScoringClient
	userID string
}

func (s *contributorStatsScoring) GetVCSUserScoreBreakdown(_ context.Context, userID string) (lbscoringclient.UserScoreBreakdownResponse, error) {
	s.userID = userID
	return lbscoringclient.UserScoreBreakdownResponse{PR: 11, Review: 7, Issue: 4}, nil
}

func TestGetContributorStats_ScoreBreakdownAndHistory(t *testing.T) {
	initStatsTestLogger(t)
	contributorID := types.ID(23)
	earnedAt := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	repo := contributorStatsRepo{history: []ScoreRecord{{ProjectID: types.ID(8), Activity: "pull request", Score: 9, EarnedAt: earnedAt}}}
	contributor := &contributorStatsContributor{}
	scoring := &contributorStatsScoring{}
	svc := Service{repository: repo, contributorClient: contributor, lbScoringClient: scoring}

	got, err := svc.GetContributorStats(context.Background(), contributorID)
	require.NoError(t, err)
	assert.Equal(t, contributorID, contributor.requestedID)
	assert.Equal(t, "987654", scoring.userID)
	assert.Equal(t, ScoreBreakdown{PR: 11, Review: 7, Issue: 4}, got.ScoreBreakdown)
	assert.Equal(t, []ScoreEntry{{Activity: "pull request", Score: 9, EarnedAt: earnedAt}}, got.ScoreHistory[types.ID(8)])
}

func initStatsTestLogger(t *testing.T) {
	t.Helper()
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))
}

// fakeRedisLB returns fixed data instead of talking to Redis.
type fakeRedisLB struct {
	entries []UserScoreEntry
	total   int64
	err     error
}

func (f *fakeRedisLB) GetPublicLeaderboardPaginated(ctx context.Context, projectID types.ID, page, pageSize int32) ([]UserScoreEntry, int64, error) {
	return f.entries, f.total, f.err
}

func (f *fakeRedisLB) SetPublicLeaderboard(ctx context.Context, projectID types.ID, userScores map[int]float64, ttl time.Duration) error {
	return nil
}

func TestGetPublicLeaderboard(t *testing.T) {
	tests := []struct {
		name           string
		repo           *fakeRedisLB
		page           int32
		pageSize       int32
		wantRanks      []uint64
		wantTotalPages int32
		wantErr        bool
	}{
		{
			name: "first page",
			repo: &fakeRedisLB{
				entries: []UserScoreEntry{{UserID: 1, Score: 50}, {UserID: 2, Score: 40}},
				total:   2,
			},
			page:           1,
			pageSize:       10,
			wantRanks:      []uint64{1, 2},
			wantTotalPages: 1,
		},
		{
			name: "second page",
			repo: &fakeRedisLB{
				entries: []UserScoreEntry{{UserID: 11, Score: 30}, {UserID: 12, Score: 20}, {UserID: 13, Score: 10}},
				total:   13,
			},
			page:           2,
			pageSize:       10,
			wantRanks:      []uint64{11, 12, 13},
			wantTotalPages: 2,
		},
		{
			name:           "empty project",
			repo:           &fakeRedisLB{entries: nil, total: 0},
			page:           1,
			pageSize:       10,
			wantRanks:      nil,
			wantTotalPages: 0,
		},
		{
			name:     "unknown project returns error",
			repo:     &fakeRedisLB{err: errors.New("project not found")},
			page:     1,
			pageSize: 10,
			wantErr:  true,
		},
		{
			name:     "zero page size returns error",
			repo:     &fakeRedisLB{entries: []UserScoreEntry{{UserID: 1, Score: 50}}, total: 1},
			page:     1,
			pageSize: 0,
			wantErr:  true,
		},
		{
			name:     "negative page size returns error",
			repo:     &fakeRedisLB{entries: []UserScoreEntry{{UserID: 1, Score: 50}}, total: 1},
			page:     1,
			pageSize: -1,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := Service{redisLeaderboardRepo: tt.repo}

			got, err := s.GetPublicLeaderboard(context.Background(), types.ID(1), tt.pageSize, tt.page)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)

			var ranks []uint64
			for _, u := range got.UsersScore {
				ranks = append(ranks, u.Rank)
			}
			assert.Equal(t, tt.wantRanks, ranks)
			assert.Equal(t, tt.wantTotalPages, got.TotalPages)
		})
	}
}
