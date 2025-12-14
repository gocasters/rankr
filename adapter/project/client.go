package project

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/grpc"
	projectpb "github.com/gocasters/rankr/protobuf/golang/project/v1"
)

type Client struct {
	rpcClient     *grpc.RPCClient
	projectClient projectpb.ProjectServiceClient
}

func New(rpcClient *grpc.RPCClient) (*Client, error) {
	if rpcClient == nil || rpcClient.Conn == nil {
		return nil, fmt.Errorf("grpc RPC client not initialized (nil connection)")
	}

	return &Client{
		rpcClient:     rpcClient,
		projectClient: projectpb.NewProjectServiceClient(rpcClient.Conn),
	}, nil
}

type ProjectItem struct {
	ProjectID    string
	Slug         string
	Name         string
	RepoProvider string
	GitRepoID    string
}

type ListProjectsRequest struct {
	PageSize int32
	Offset   int32
}

type ListProjectsResponse struct {
	Projects   []ProjectItem
	TotalCount int32
}

func (c *Client) ListProjects(ctx context.Context, req *ListProjectsRequest) (*ListProjectsResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("ListProjects: request cannot be nil")
	}

	pbReq := &projectpb.ListProjectsRequest{
		PageSize: req.PageSize,
		Offset:   req.Offset,
	}

	pbRes, err := c.projectClient.ListProjects(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}

	projects := make([]ProjectItem, 0, len(pbRes.Projects))
	for _, p := range pbRes.Projects {
		projects = append(projects, ProjectItem{
			ProjectID:    p.ProjectId,
			Slug:         p.Slug,
			Name:         p.Name,
			RepoProvider: p.RepoProvider,
			GitRepoID:    p.GitRepoId,
		})
	}

	return &ListProjectsResponse{
		Projects:   projects,
		TotalCount: pbRes.TotalCount,
	}, nil
}

type GetProjectByRepoRequest struct {
	RepoProvider string
	RepoID       string
}

type GetProjectByRepoResponse struct {
	ProjectID    string
	Slug         string
	Name         string
	RepoProvider string
	GitRepoID    string
}

func (c *Client) GetProjectByRepo(ctx context.Context, req *GetProjectByRepoRequest) (*GetProjectByRepoResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("GetProjectByRepo: request cannot be nil")
	}

	pbReq := &projectpb.GetProjectByRepoRequest{
		RepoProvider: req.RepoProvider,
		RepoId:       req.RepoID,
	}

	pbRes, err := c.projectClient.GetProjectByRepo(ctx, pbReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get project by repo: %w", err)
	}

	return &GetProjectByRepoResponse{
		ProjectID:    pbRes.ProjectId,
		Slug:         pbRes.Slug,
		Name:         pbRes.Name,
		RepoProvider: pbRes.RepoProvider,
		GitRepoID:    pbRes.GitRepoId,
	}, nil
}

func (c *Client) Close() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}
