package grpc

import (
	"context"
	"log/slog"

	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/logger"
	contributorpb "github.com/gocasters/rankr/protobuf/golang/contributor/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	contributorpb.UnimplementedContributorServiceServer
	contributorSvc contributor.Service
}

func NewHandler(contributorSvc contributor.Service) Handler {
	return Handler{
		contributorSvc: contributorSvc,
	}
}

func (h Handler) GetContributorsByVCS(ctx context.Context, req *contributorpb.GetContributorsByVCSRequest) (*contributorpb.GetContributorsByVCSResponse, error) {
	log := logger.L()
	log.Info("gRPC GetContributorsByVCS request received", slog.Any("request", req))

	if !contributor.IsValidVcsProvider(req.VcsProvider) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid vcs_provider: %s", req.VcsProvider)
	}

	if len(req.Usernames) == 0 {
		return nil, status.Error(codes.InvalidArgument, "usernames cannot be empty")
	}

	serviceReq := contributor.GetContributorsByVCSRequest{
		VcsProvider: contributor.VcsProvider(req.VcsProvider),
		Usernames:   req.Usernames,
	}

	serviceResp, err := h.contributorSvc.GetContributorsByVCS(ctx, serviceReq)
	if err != nil {
		log.Error("failed to get contributors by VCS", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to retrieve contributors")
	}

	mappings := make([]*contributorpb.ContributorMapping, 0, len(serviceResp.Contributors))
	for _, c := range serviceResp.Contributors {
		mappings = append(mappings, &contributorpb.ContributorMapping{
			ContributorId: c.ContributorID,
			VcsUsername:   c.VcsUsername,
			VcsUserId:     c.VcsUserID,
		})
	}

	return &contributorpb.GetContributorsByVCSResponse{
		VcsProvider:  string(serviceResp.VcsProvider),
		Contributors: mappings,
	}, nil
}
