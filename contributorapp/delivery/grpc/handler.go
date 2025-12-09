package grpc

import (
	"context"

	"github.com/gocasters/rankr/pkg/logger"
	contributorpb "github.com/gocasters/rankr/protobuf/golang/contributor/v1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Handler implements the contributor gRPC service.
// For now we return a static credential so the auth service can complete login flows locally.
type Handler struct {
	contributorpb.UnimplementedContributorServiceServer
}

func NewHandler() Handler {
	return Handler{}
}

func (h Handler) GetContributor(_ context.Context, req *contributorpb.GetContributorRequest) (*contributorpb.GetContributorResponse, error) {
	username := req.GetGithubUsername()
	if username == "" {
		return nil, status.Error(codes.InvalidArgument, "github_username is required")
	}

	logger.L().Info("GetContributor request received", "github_username", username)

	// Temporary hardcoded credentials for local testing.
	// Only allow a single known user to avoid authenticating arbitrary usernames.
	if username != "testuser" {
		return nil, status.Error(codes.NotFound, "contributor not found")
	}

	return &contributorpb.GetContributorResponse{
		ContributorId: 1,
		Password:      "pass",
	}, nil
}
