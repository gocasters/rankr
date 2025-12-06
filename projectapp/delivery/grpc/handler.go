package grpc

import (
	"context"
	"errors"

	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/gocasters/rankr/projectapp/service/project"
	projectpb "github.com/gocasters/rankr/protobuf/golang/project/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log/slog"
)

type Handler struct {
	projectpb.UnimplementedProjectServiceServer
	projectSvc *project.Service
}

func NewHandler(projectSvc *project.Service) Handler {
	return Handler{
		projectSvc: projectSvc,
	}
}

func (h Handler) GetProjectByRepo(ctx context.Context, req *projectpb.GetProjectByRepoRequest) (*projectpb.GetProjectByRepoResponse, error) {
	log := logger.L()
	log.Info("gRPC GetProjectByRepo request received", slog.Any("request", req))

	serviceReq := project.GetProjectByVCSRepoRequest{
		Provider: constant.VcsProvider(req.RepoProvider),
		RepoID:   req.RepoId,
	}

	serviceResp, err := h.projectSvc.GetProjectByVCSRepo(ctx, serviceReq)
	if err != nil {
		if errors.Is(err, constant.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "project not found")
		}
		log.Error("failed to get project by VCS repo", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to retrieve project")
	}

	gitRepoID := ""
	if serviceResp.GitRepoID != nil {
		gitRepoID = *serviceResp.GitRepoID
	}

	repoProvider := ""
	if serviceResp.RepoProvider != nil {
		repoProvider = string(*serviceResp.RepoProvider)
	}

	return &projectpb.GetProjectByRepoResponse{
		ProjectId:    serviceResp.ID,
		Slug:         serviceResp.Slug,
		Name:         serviceResp.Name,
		RepoProvider: repoProvider,
		GitRepoId:    gitRepoID,
	}, nil
}
