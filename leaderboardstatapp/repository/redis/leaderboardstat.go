package redis

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/redis/go-redis/v9"
	"strconv"
)

type LeaderboardRepo struct {
	Redis *redis.Client
}

func NewLeaderboardRepo(redis *redis.Client) *LeaderboardRepo {
	return &LeaderboardRepo{}
}

func (d DB) GetLeaderboardByFilters(ctx context.Context, page int, pageSize int, category string, timeframe string) (leaderboardstat.ScoreboardResponse, error) {

	key := fmt.Sprintf("%s:%s", category, timeframe)
	client := d.adapter.Client()
	list, err := client.ZRevRangeWithScores(ctx,
		key,
		int64((page-1)*pageSize),
		int64(page*pageSize-1),
	).Result()

	if err != nil {
		return leaderboardstat.ScoreboardResponse{}, fmt.Errorf("failed to get leaderboard: %w", err)
	}

	fmt.Println(list)

	var response leaderboardstat.ScoreboardResponse
	for _, z := range list {

		contributorID, err := strconv.Atoi(z.Member.(string))
		if err != nil {
			return leaderboardstat.ScoreboardResponse{}, fmt.Errorf("invalid contributor ID format: %w", err)
		}

		rank, err := client.ZRevRank(ctx, key, z.Member.(string)).Result()
		if err != nil {
			return leaderboardstat.ScoreboardResponse{}, fmt.Errorf("failed to get rank: %w", err)
		}

		response.Entries = append(response.Entries, leaderboardstat.ScoreboardItem{
			Rank:          int(rank) + 1,
			ContributorID: contributorID,
			Score:         int(z.Score),
		})
	}

	return response, nil
}

func (d DB) CreateNewScoreList(ctx context.Context, redisKey string, entries []leaderboardstat.LeaderboardEntry) error {
	client := d.adapter.Client()
	pipe := client.Pipeline()

	// Delete existing key to replace with fresh data
	pipe.Del(ctx, redisKey)

	// Prepare ZADD commands for all entries
	zMembers := make([]redis.Z, len(entries))
	for i, entry := range entries {
		zMembers[i] = redis.Z{
			Score:  float64(entry.Score),
			Member: entry.ContributorID,
		}
	}

	pipe.ZAdd(ctx, redisKey, zMembers...)

	_, err := pipe.Exec(ctx)
	return err
}

func (d DB) BulkCreateScoreLists(ctx context.Context, keys []string, entries []leaderboardstat.LeaderboardEntry) error {
	client := d.adapter.Client()
	pipe := client.Pipeline()

	for _, key := range keys {
		// Prepare Z members for each key
		zMembers := make([]redis.Z, len(entries))
		for i, entry := range entries {
			zMembers[i] = redis.Z{
				Score:  float64(entry.Score),
				Member: entry.ContributorID,
			}
		}

		// Clear existing and add new entries
		pipe.Del(ctx, key)
		pipe.ZAdd(ctx, key, zMembers...)
	}

	_, err := pipe.Exec(ctx)
	return err
}
