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

func New(rpcClient *grpc.RPCClient) (*Client, error) {
	if rpcClient == nil || rpcClient.Conn == nil {
		return nil, fmt.Errorf("grpc RPC client no initialized (nil connection)")
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
