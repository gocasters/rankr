package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/adapter/leaderboardstat"
	types "github.com/gocasters/rankr/type"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorStatClient interface {
	GetContributorStats(ctx context.Context,
		req leaderboardstat.GetContributorStatRequest) (leaderboardstat.GetContributorStatResponse, error)
}

type ContributorStatAdapter struct {
	contributorStatClient ContributorStatClient
}

func NewContributorStatAdapter(contributorStatClient ContributorStatClient) ContributorStatAdapter {
	return ContributorStatAdapter{contributorStatClient: contributorStatClient}
}

func (a ContributorStatAdapter) GetContributorStat(ctx context.Context, contributorID int64) (userprofile.ContributorStat, error) {
	res, err := a.contributorStatClient.GetContributorStats(ctx,
		leaderboardstat.GetContributorStatRequest{ContributorID: types.ID(contributorID)})
	if err != nil {
		return userprofile.ContributorStat{}, fmt.Errorf("failed to get contributor stat: %w", err)
	}

	scoreHistory := make(map[types.ID][]userprofile.ScoreEntry)
	for id, se := range res.ScoreHistory {
		scoreEntries := make([]userprofile.ScoreEntry, 0, len(se))
		for _, e := range se {
			scoreEntries = append(scoreEntries, userprofile.ScoreEntry{
				Score:    e.Score,
				EarnedAt: e.EarnedAt,
				Activity: e.Activity,
			})
		}

		scoreHistory[id] = scoreEntries
	}

	return userprofile.ContributorStat{
		ContributorID: res.ContributorID,
		GlobalRank:    res.GlobalRank,
		TotalScore:    res.TotalScore,
		ProjectsScore: res.ProjectsScore,
		ScoreHistory:  scoreHistory,
	}, nil
}
