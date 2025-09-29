package grpc

import (
	"context"
	"errors"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	leaderboardscoringpb "github.com/gocasters/rankr/protobuf/golang/leaderboardscoring/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

type Handler struct {
	leaderboardscoringpb.UnimplementedLeaderboardScoringServiceServer
	leaderboardScoringSvc *leaderboardscoring.Service
}

func NewHandler(leaderboardScoringSvc *leaderboardscoring.Service) Handler {
	return Handler{
		//UnimplementedLeaderboardScoringServiceServer: leaderboardscoringpb.UnimplementedLeaderboardScoringServiceServer{},
		leaderboardScoringSvc: leaderboardScoringSvc,
	}
}

func (h Handler) GetLeaderboard(ctx context.Context, req *leaderboardscoringpb.GetLeaderboardRequest) (*leaderboardscoringpb.GetLeaderboardResponse, error) {
	log := logger.L()
	log.Info("gRPC GetLeaderboard request received", slog.Any("request", req))

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
		log.Error(
			"failed to get leaderboard scoring from service",
			slog.String("error", err.Error()),
			slog.Any("request", req),
		)

		switch {
		case errors.Is(err, leaderboardscoring.ErrLeaderboardNotFound):
			return nil, status.Error(codes.NotFound, "The requested leaderboard could not be found.")

		case errors.Is(err, leaderboardscoring.ErrInvalidArguments):
			return nil, status.Error(codes.InvalidArgument, "Invalid request parameters provided.")

		default:
			return nil, status.Error(codes.Internal, "An unexpected internal error occurred.")
		}
	}

	leaderboardPBRes := leaderboardResToProtobuf(leaderboardRes)
	log.Debug("Successfully prepared gRPC response", slog.Int("row_count", len(leaderboardPBRes.Rows)))
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
