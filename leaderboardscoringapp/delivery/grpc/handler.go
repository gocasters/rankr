package grpc

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/protobuf/leaderboardscoring/golang/leaderboardscoringpb"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

type Handler struct {
	leaderboardscoringpb.UnimplementedLeaderboardScoringServiceServer
	leaderboardScoringSvc leaderboardscoring.Service
}

func NewHandler(leaderboardScoringSvc leaderboardscoring.Service) Handler {
	return Handler{
		UnimplementedLeaderboardScoringServiceServer: leaderboardscoringpb.UnimplementedLeaderboardScoringServiceServer{},
		leaderboardScoringSvc:                        leaderboardScoringSvc,
	}
}

func (h Handler) GetLeaderboard(ctx context.Context, req *leaderboardscoringpb.GetLeaderboardRequest) (*leaderboardscoringpb.GetLeaderboardResponse, error) {
	logger := logger.L()

	var projectIDPtr *string
	if pid := req.GetProjectId(); pid != "" {
		projectIDPtr = &pid
	}

	leaderboardReq := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: leaderboardscoring.Timeframe(req.GetTimeframe()),
		ProjectID: projectIDPtr,
		PageSize:  req.GetPageSize(),
		Offset:    req.GetOffset(),
	}

	leaderboardRes, err := h.leaderboardScoringSvc.GetLeaderboard(ctx, leaderboardReq)
	if err != nil {
		logger.Error(
			"failed to get leaderboard scoring from service",
			slog.String("error", err.Error()),
			slog.Any("request", req),
		)

		// TODO: replace with concrete error mapping from service layer
		return nil, status.Errorf(codes.Internal, "get leaderboard failed: %v", err)
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

	var projectID string
	if leaderboardRes.ProjectID != nil {
		projectID = *leaderboardRes.ProjectID
	}
	leaderboardPBRes := &leaderboardscoringpb.GetLeaderboardResponse{
		Timeframe: leaderboardscoringpb.Timeframe(leaderboardRes.Timeframe),
		ProjectId: &projectID,
		Rows:      rows,
	}
	return leaderboardPBRes
}
