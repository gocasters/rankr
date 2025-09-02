package grpc

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/protobuf/leaderboardscoring/golang/leaderboardscoringpb"
	"log/slog"
)

type Handler struct {
	leaderboardscoringpb.UnimplementedLeaderboardScoringServiceServer
	leaderboardScoringSvc leaderboardscoring.Service
	Logger                *slog.Logger
}

func NewHandler(leaderboardScoringSvc leaderboardscoring.Service, logger *slog.Logger) Handler {
	return Handler{
		UnimplementedLeaderboardScoringServiceServer: leaderboardscoringpb.UnimplementedLeaderboardScoringServiceServer{},
		leaderboardScoringSvc:                        leaderboardScoringSvc,
		Logger:                                       logger,
	}
}

func (h Handler) GetLeaderboard(ctx context.Context, req *leaderboardscoringpb.GetLeaderboardRequest) (*leaderboardscoringpb.GetLeaderboardResponse, error) {

	leaderboardReq := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Timeframe(req.GetTimeframe()),
		ProjectID: req.ProjectId,
		PageSize:  req.GetPageSize(),
		Offset:    req.GetOffset(),
	}

	leaderboardRes, err := h.leaderboardScoringSvc.GetLeaderboard(ctx, leaderboardReq)
	if err != nil {
		h.Logger.Error(
			"failed to get leaderboard scoring from service",
			slog.String("error", err.Error()),
			slog.Any("request", req),
		)
		// TODO: Map service errors to gRPC status codes

		return nil, err
	}

	leaderboardPBRes := leaderboardResToProtobuf(leaderboardRes)

	return leaderboardPBRes, nil
}

func leaderboardResToProtobuf(leaderboardRes leaderboardscoring.GetLeaderboardResponse) *leaderboardscoringpb.GetLeaderboardResponse {
	rows := make([]*leaderboardscoringpb.LeaderboardRow, 0, len(leaderboardRes.LeaderboardRows))
	for _, r := range leaderboardRes.LeaderboardRows {
		leaderboardRow := &leaderboardscoringpb.LeaderboardRow{
			Rank:   r.Rank,
			UserId: r.UserID,
			Score:  r.Score,
		}

		rows = append(rows, leaderboardRow)
	}

	leaderboardPBRes := &leaderboardscoringpb.GetLeaderboardResponse{
		Timeframe: leaderboardscoringpb.Timeframe(leaderboardRes.Timeframe),
		ProjectId: leaderboardRes.ProjectID,
		Rows:      rows,
	}
	return leaderboardPBRes
}
