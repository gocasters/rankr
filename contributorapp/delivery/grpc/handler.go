package grpc

import (
	"context"

	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/logger"
	contributorpb "github.com/gocasters/rankr/protobuf/golang/contributor/v1"
	types "github.com/gocasters/rankr/type"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)
type Handler struct {
	contributorpb.UnimplementedContributorServiceServer
	svc contributor.Service
}

func NewHandler(svc contributor.Service) Handler {
	return Handler{svc: svc}
}

func (h Handler) VerifyPassword(ctx context.Context, req *contributorpb.VerifyPasswordRequest) (*contributorpb.VerifyPasswordResponse, error) {
	res, err := h.svc.VerifyPassword(ctx, contributor.VerifyPasswordRequest{
		ID:             types.ID(req.GetContributorId()),
		GitHubUsername: req.GetGithubUsername(),
		Password:       req.GetPassword(),
	})
	if err != nil {
		logger.L().Warn("verify_password_failed", "error", err)
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	return &contributorpb.VerifyPasswordResponse{
		Valid:         res.Valid,
		ContributorId: int64(res.ID),
	}, nil
}
