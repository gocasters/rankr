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

	if !constant.IsValidVcsProvider(req.RepoProvider) {
		return nil, status.Errorf(codes.InvalidArgument, "invalid repo_provider: %s", req.RepoProvider)
	}

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

func (h Handler) ListProjects(ctx context.Context, req *projectpb.ListProjectsRequest) (*projectpb.ListProjectsResponse, error) {
	log := logger.L()
	log.Info("gRPC ListProjects request received", slog.Any("request", req))

	serviceInput := project.ListProjectsInput{
		PageSize: req.PageSize,
		Offset:   req.Offset,
	}

	serviceResp, err := h.projectSvc.ListProjects(ctx, serviceInput)
	if err != nil {
		log.Error("failed to list projects", slog.String("error", err.Error()))
		return nil, status.Error(codes.Internal, "failed to retrieve projects")
	}

	projects := make([]*projectpb.ProjectItem, 0, len(serviceResp.Projects))
	for _, p := range serviceResp.Projects {
		projects = append(projects, &projectpb.ProjectItem{
			ProjectId: p.ID,
			Slug:      p.Slug,
			Name:      p.Name,
		})
	}

	return &projectpb.ListProjectsResponse{
		Projects:   projects,
		TotalCount: serviceResp.TotalCount,
	}, nil
}
