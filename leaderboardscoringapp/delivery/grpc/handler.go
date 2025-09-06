package grpc

import (
	"context"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	leaderboardscoringpb "github.com/gocasters/rankr/protobuf/golang/leaderboardscoring/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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
		h.Logger.Error(
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
