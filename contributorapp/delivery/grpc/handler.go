package grpc

import (
	"context"

	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/projectapp/constant"
	contributorpb "github.com/gocasters/rankr/protobuf/golang/contributor/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

type Handler struct {
	contributorpb.UnimplementedContributorServiceServer
	contributorSvc *contributor.Service
}

func NewHandler(contributorSvc *contributor.Service) Handler {
	return Handler{
		contributorSvc: contributorSvc,
	}
}

func (h Handler) GetContributor(ctx context.Context, req *contributorpb.GetContributorRequest) (*contributorpb.GetContributorResponse, error) {
	log := logger.L()
	log.Info("gRPC GetContributor request received", slog.Any("request", req))

	return nil, status.Error(codes.Unimplemented, "method GetContributor not yet implemented")
}

func (h Handler) GetContributorsByRepo(ctx context.Context, req *contributorpb.GetContributorsByRepoRequest) (*contributorpb.GetContributorsByRepoResponse, error) {
	log := logger.L()
	log.Info("gRPC GetContributorsByRepo request received", slog.Any("request", req))

	serviceReq := contributor.GetContributorsByVCSRequest{
		Provider:  constant.VcsProvider(req.RepoProvider),
		Usernames: req.Usernames,
	}

	serviceResp, err := h.contributorSvc.GetContributorsByVCS(ctx, serviceReq)
	if err != nil {
		log.Error("failed to get contributors by VCS", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to retrieve contributors")
	}

	mappings := make([]*contributorpb.ContributorMapping, 0, len(serviceResp.Contributors))
	for _, c := range serviceResp.Contributors {
		mappings = append(mappings, &contributorpb.ContributorMapping{
			ContributorId: uint64(c.ContributorID),
			VcsUsername:   c.VCSUsername,
			VcsUserId:     uint64(c.VCSUserID),
		})
	}

	return &contributorpb.GetContributorsByRepoResponse{
		RepoProvider: string(serviceResp.Provider),
		Contributors: mappings,
	}, nil
}
