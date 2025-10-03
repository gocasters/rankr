package leaderboardstat

import (
	"context"
	"fmt"
	lbstat "github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/grpc"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	types "github.com/gocasters/rankr/type"
)

type Client struct {
	rpcClient             *grpc.RPCClient
	leaderboardStatClient leaderboardstatpb.LeaderboardStatServiceClient
}

func New(rpcClient *grpc.RPCClient) (*Client, error) {
	if rpcClient == nil || rpcClient.Conn == nil {
		return nil, fmt.Errorf("grpc RPC client no initialized (nil connection)")
	}

	return &Client{
		rpcClient:             rpcClient,
		leaderboardStatClient: leaderboardstatpb.NewLeaderboardStatServiceClient(rpcClient.Conn),
	}, nil
}

func (c *Client) Close() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}

func (c *Client) GetContributorTotalStats(ctx context.Context, getTotalStatsReq *lbstat.ContributorStatsRequest) (*lbstat.ContributorStatsResponse, error) {

	statPBReq := &leaderboardstatpb.ContributorStatRequest{
		ContributorId: int64(getTotalStatsReq.ContributorID),
	}

	statPRRes, err := c.leaderboardStatClient.GetContributorTotalStats(ctx, statPBReq)
	if err != nil {
		return nil, err
	}

	return &lbstat.ContributorStatsResponse{
		ContributorID: types.ID(statPRRes.ContributorId),
		GlobalRank:    int(statPRRes.GlobalRank),
		TotalScore:    statPRRes.TotalScore,
		ProjectsScore: statPRRes.ProjectsScore,
	}, nil
}
