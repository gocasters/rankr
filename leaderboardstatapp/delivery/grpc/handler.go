package grpc

import (
	"context"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/slice"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	types "github.com/gocasters/rankr/type"
	"google.golang.org/protobuf/types/known/timestamppb"
	"log/slog"
)

type Handler struct {
	leaderboardstatpb.UnimplementedLeaderboardStatServiceServer
	leaderboardStatSvc leaderboardstat.Service
}

func NewHandler(leaderboardStatSvc leaderboardstat.Service) Handler {
	return Handler{
		leaderboardStatSvc: leaderboardStatSvc,
	}
}

func (h Handler) GetContributorStats(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
	log := logger.L()
	log.Info("gRPC GetLeaderboard request received", slog.Any("request", req))

	contributorId := req.GetContributorId()

	statsRes, err := h.leaderboardStatSvc.GetContributorStats(ctx, types.ID(contributorId))
	if err != nil {
		return nil, err
	}

	scoreHistory := make(map[uint64]*leaderboardstatpb.ProjectScoreHistory)
	for projectID, entries := range statsRes.ScoreHistory {
		projectHistory := &leaderboardstatpb.ProjectScoreHistory{
			Entries: transformScoreEntries(entries),
		}
		scoreHistory[uint64(projectID)] = projectHistory
	}

	contributorStatResponse := &leaderboardstatpb.ContributorStatResponse{
		ContributorId: contributorId,
		GlobalRank:    uint64(statsRes.GlobalRank),
		TotalScore:    statsRes.TotalScore,
		ProjectsScore: slice.MapFromIDFloat64ToUint64Float64(statsRes.ProjectsScore),
		ScoreHistory:  scoreHistory,
	}

	return contributorStatResponse, nil
}

func transformScoreEntries(entries []leaderboardstat.ScoreEntry) []*leaderboardstatpb.ScoreEntry {
	pbEntries := make([]*leaderboardstatpb.ScoreEntry, 0, len(entries))

	for _, entry := range entries {
		pbEntry := &leaderboardstatpb.ScoreEntry{
			Activity: entry.Activity,
			Score:    entry.Score,
			EarnedAt: timestamppb.New(entry.EarnedAt),
		}
		pbEntries = append(pbEntries, pbEntry)
	}

	return pbEntries
}

/*
func (h Handler) GetContributorTotalStats(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
	stats, err := h.leaderboardStatSvc.GetContributorTotalStats(ctx, types.ID(req.GetContributorId()))
	if err != nil {
		return nil, err
	}

	return &leaderboardstatpb.ContributorStatResponse{
		ContributorId: req.GetContributorId(),
		GlobalRank:    uint64(stats.GlobalRank),
		TotalScore:    stats.TotalScore,
		ProjectsScore: slice.MapFromIDFloat64ToUint64Float64(stats.ProjectsScore),
		ScoreHistory:  nil, // TODO - add history
	}, nil
}
*/
