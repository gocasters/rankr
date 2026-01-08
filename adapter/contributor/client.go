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

func (c *Client) VerifyPassword(ctx context.Context, username string, password string) (types.ID, string, bool, error) {
	req := &contributorpb.VerifyPasswordRequest{
		Password:      password,
		GithubUsername: username,
	}

	res, err := c.contributorService.VerifyPassword(ctx, req)
	if err != nil {
		return 0, "", false, err
	}
	if res == nil {
		return 0, "", false, fmt.Errorf("empty verify password response")
	}

	return types.ID(res.GetContributorId()), res.GetRole(), res.GetValid(), nil
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
