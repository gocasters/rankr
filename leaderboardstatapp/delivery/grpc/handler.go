package grpc

import (
	"context"
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/logger"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	types "github.com/gocasters/rankr/type"
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

func (h Handler) GetContributorTotalStats(ctx context.Context, req *leaderboardstatpb.ContributorStatRequest) (*leaderboardstatpb.ContributorStatResponse, error) {
	log := logger.L()
	log.Info("gRPC GetLeaderboard request received", slog.Any("request", req))

	contributorId := req.GetContributorId()

	statsRes, err := h.leaderboardStatSvc.GetContributorTotalStats(ctx, types.ID(contributorId))
	if err != nil {
		return nil, err
	}

	contributorStatResponse := &leaderboardstatpb.ContributorStatResponse{
		ContributorId: contributorId,
		GlobalRank:    int64(statsRes.GlobalRank),
		TotalScore:    statsRes.TotalScore,
		ProjectsScore: statsRes.ProjectsScore,
	}

	return contributorStatResponse, err
}
