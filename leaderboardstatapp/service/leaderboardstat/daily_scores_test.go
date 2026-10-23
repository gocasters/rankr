package leaderboardstat

import (
	"context"
	"errors"
	"testing"

	"github.com/gocasters/rankr/adapter/project"
	lbscoring "github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/cachemanager"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeScoringClient struct {
	rowsByProject map[string][]lbscoring.LeaderboardRow
	err           error
	requests      []lbscoring.GetLeaderboardRequest
}

func (f *fakeScoringClient) GetLeaderboard(_ context.Context, req *lbscoring.GetLeaderboardRequest) (*lbscoring.GetLeaderboardResponse, error) {
	f.requests = append(f.requests, *req)
	if f.err != nil {
		return nil, f.err
	}
	if req.ProjectID == nil {
		return &lbscoring.GetLeaderboardResponse{Timeframe: req.Timeframe}, nil
	}
	return &lbscoring.GetLeaderboardResponse{
		Timeframe:       req.Timeframe,
		ProjectID:       req.ProjectID,
		LeaderboardRows: f.rowsByProject[*req.ProjectID],
	}, nil
}

type fakeProjectClient struct {
	projects []project.ProjectItem
	err      error
}

func (f *fakeProjectClient) ListProjects(_ context.Context, req *project.ListProjectsRequest) (*project.ListProjectsResponse, error) {
	if f.err != nil {
		return nil, f.err
	}
	if int(req.Offset) >= len(f.projects) {
		return &project.ListProjectsResponse{}, nil
	}
	end := min(int(req.Offset+req.PageSize), len(f.projects))
	return &project.ListProjectsResponse{Projects: f.projects[req.Offset:end], TotalCount: int32(len(f.projects))}, nil
}

type fakeDailyScoreRepo struct {
	Repository
	stored []DailyContributorScore
}

func (f *fakeDailyScoreRepo) StoreDailyContributorScores(_ context.Context, scores []DailyContributorScore) error {
	f.stored = append(f.stored, scores...)
	return nil
}

func (f *fakeDailyScoreRepo) GetPendingDailyScores(context.Context) ([]DailyContributorScore, error) {
	return nil, nil
}

func initDailyScoresTestLogger(t *testing.T) {
	t.Helper()
	require.NoError(t, logger.Init(logger.Config{Level: "error"}))
}

func ptr(s string) *string { return &s }

func TestGetDailyContributorScores_ProjectAttribution(t *testing.T) {
	initDailyScoresTestLogger(t)

	const (
		repoA = "1001"
		repoB = "2002"
	)

	tests := []struct {
		name      string
		projects  []project.ProjectItem
		rows      map[string][]lbscoring.LeaderboardRow
		scoreErr  error
		projErr   error
		wantErr   bool
		wantCalls int
		want      []DailyContributorScore
	}{
		{
			name: "contributor scores land on the project that produced them",
			projects: []project.ProjectItem{
				{ProjectID: "p-a", GitRepoID: repoA},
				{ProjectID: "p-b", GitRepoID: repoB},
			},
			rows: map[string][]lbscoring.LeaderboardRow{
				repoA: {{Rank: 1, UserID: "7", Score: 30, ProjectID: ptr(repoA)}},
				repoB: {
					{Rank: 1, UserID: "8", Score: 50, ProjectID: ptr(repoB)},
					{Rank: 2, UserID: "7", Score: 10, ProjectID: ptr(repoB)},
				},
			},
			wantCalls: 2,
			want: []DailyContributorScore{
				{ContributorID: 7, UserID: "7", Score: 30, Rank: 1, ProjectID: 1001, Timeframe: "daily"},
				{ContributorID: 8, UserID: "8", Score: 50, Rank: 1, ProjectID: 2002, Timeframe: "daily"},
				{ContributorID: 7, UserID: "7", Score: 10, Rank: 2, ProjectID: 2002, Timeframe: "daily"},
			},
		},
		{
			name:     "row without project_id falls back to the requested project",
			projects: []project.ProjectItem{{ProjectID: "p-a", GitRepoID: repoA}},
			rows: map[string][]lbscoring.LeaderboardRow{
				repoA: {{Rank: 1, UserID: "7", Score: 30}},
			},
			wantCalls: 1,
			want: []DailyContributorScore{
				{ContributorID: 7, UserID: "7", Score: 30, Rank: 1, ProjectID: 1001, Timeframe: "daily"},
			},
		},
		{
			name: "project without git_repo_id is skipped",
			projects: []project.ProjectItem{
				{ProjectID: "p-empty"},
				{ProjectID: "p-a", GitRepoID: repoA},
			},
			rows: map[string][]lbscoring.LeaderboardRow{
				repoA: {{Rank: 1, UserID: "7", Score: 30, ProjectID: ptr(repoA)}},
			},
			wantCalls: 1,
			want: []DailyContributorScore{
				{ContributorID: 7, UserID: "7", Score: 30, Rank: 1, ProjectID: 1001, Timeframe: "daily"},
			},
		},
		{
			name:     "non-numeric user_id is skipped",
			projects: []project.ProjectItem{{ProjectID: "p-a", GitRepoID: repoA}},
			rows: map[string][]lbscoring.LeaderboardRow{
				repoA: {
					{Rank: 1, UserID: "ghost", Score: 99, ProjectID: ptr(repoA)},
					{Rank: 2, UserID: "7", Score: 30, ProjectID: ptr(repoA)},
				},
			},
			wantCalls: 1,
			want: []DailyContributorScore{
				{ContributorID: 7, UserID: "7", Score: 30, Rank: 2, ProjectID: 1001, Timeframe: "daily"},
			},
		},
		{
			name:      "no projects stores nothing",
			wantCalls: 0,
		},
		{
			name:     "project listing failure stores nothing",
			projects: []project.ProjectItem{{ProjectID: "p-a", GitRepoID: repoA}},
			projErr:  errors.New("project service down"),
			wantErr:  true,
		},
		{
			name: "scoring failure on any project stores nothing",
			projects: []project.ProjectItem{
				{ProjectID: "p-a", GitRepoID: repoA},
				{ProjectID: "p-b", GitRepoID: repoB},
			},
			scoreErr:  errors.New("scoring down"),
			wantErr:   true,
			wantCalls: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scoring := &fakeScoringClient{rowsByProject: tt.rows, err: tt.scoreErr}
			repo := &fakeDailyScoreRepo{}
			svc := NewService(repo, Validator{}, cachemanager.CacheManager{}, nil, scoring, &fakeProjectClient{projects: tt.projects, err: tt.projErr})

			err := svc.GetDailyContributorScores(context.Background())

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			assert.Len(t, scoring.requests, tt.wantCalls)
			for _, req := range scoring.requests {
				require.NotNil(t, req.ProjectID, "daily scores must be requested per project, never from the global board")
			}
			assert.Equal(t, tt.want, repo.stored)
		})
	}
}

func TestGetDailyContributorScores_RowProjectIDIsAuthoritative(t *testing.T) {
	initDailyScoresTestLogger(t)

	scoring := &fakeScoringClient{rowsByProject: map[string][]lbscoring.LeaderboardRow{
		"1001": {{Rank: 1, UserID: "7", Score: 30, ProjectID: ptr("2002")}},
	}}
	repo := &fakeDailyScoreRepo{}
	svc := NewService(repo, Validator{}, cachemanager.CacheManager{}, nil, scoring, &fakeProjectClient{projects: []project.ProjectItem{{GitRepoID: "1001"}}})

	require.NoError(t, svc.GetDailyContributorScores(context.Background()))

	require.Len(t, repo.stored, 1)
	assert.Equal(t, types.ID(2002), repo.stored[0].ProjectID)
}
