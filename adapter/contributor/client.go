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

func (c *Client) VerifyPassword(ctx context.Context, username string, password string) (types.ID, bool, error) {
	req := &contributorpb.VerifyPasswordRequest{
		Password:      password,
		GithubUsername: username,
	}

	res, err := c.contributorService.VerifyPassword(ctx, req)
	if err != nil {
		return 0, false, err
	}
	if res == nil {


		
		return 0, false, fmt.Errorf("empty verify password response")
	}

	return types.ID(res.GetContributorId()), res.GetValid(), nil
}
