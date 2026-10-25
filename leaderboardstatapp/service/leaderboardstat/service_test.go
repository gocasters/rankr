package leaderboardstat

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

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
			name:     "znegative page size returns error",
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
