package contributor

import (
	"context"
	"fmt"

	"github.com/gocasters/rankr/pkg/grpc"
	contributorpb "github.com/gocasters/rankr/protobuf/golang/contributor/v1"
	types "github.com/gocasters/rankr/type"
)

type Client struct {
	rpcClient          *grpc.RPCClient
	contributorService contributorpb.ContributorServiceClient
}

type Credentials struct {
	ID       types.ID
	Password string
}

type Mapping struct {
	ContributorID types.ID
	VcsUsername   string
	VcsUserID     int64
}

func New(rpcClient *grpc.RPCClient) (*Client, error) {
	if rpcClient == nil || rpcClient.Conn == nil {
		return nil, fmt.Errorf("grpc RPC client not initialized (nil connection)")
	}

	return &Client{
		rpcClient:          rpcClient,
		contributorService: contributorpb.NewContributorServiceClient(rpcClient.Conn),
	}, nil
}

func (c *Client) Close() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}

func (c *Client) GetCredentialsByGitHubUsername(ctx context.Context, githubUsername string) (Credentials, error) {
	req := &contributorpb.GetContributorRequest{
		GithubUsername: githubUsername,
	}

	res, err := c.contributorService.GetContributor(ctx, req)
	if err != nil {
		return Credentials{}, err
	}
	if res == nil {
		return Credentials{}, fmt.Errorf("empty contributor response")
	}

	return Credentials{
		ID:       types.ID(res.ContributorId),
		Password: res.Password,
	}, nil
}

func (c *Client) GetContributorsByVCS(ctx context.Context, vcsProvider string, usernames []string) ([]Mapping, error) {
	req := &contributorpb.GetContributorsByVCSRequest{
		VcsProvider: vcsProvider,
		Usernames:   usernames,
	}

	res, err := c.contributorService.GetContributorsByVCS(ctx, req)
	if err != nil {
		return nil, err
	}
	if res == nil {
		return nil, fmt.Errorf("empty contributor response")
	}

	mappings := make([]Mapping, 0, len(res.Contributors))
	for _, c := range res.Contributors {
		mappings = append(mappings, Mapping{
			ContributorID: types.ID(c.ContributorId),
			VcsUsername:   c.VcsUsername,
			VcsUserID:     c.VcsUserId,
		})
	}

	return mappings, nil
}
