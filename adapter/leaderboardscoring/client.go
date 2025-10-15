package leaderboardscoring

import (
	"context"
	"fmt"
	lbscoring "github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/grpc"
	leaderboardscoringpb "github.com/gocasters/rankr/protobuf/golang/leaderboardscoring/v1"
)

type Client struct {
	rpcClient                *grpc.RPCClient
	leaderboardScoringClient leaderboardscoringpb.LeaderboardScoringServiceClient
}

func New(rpcClient *grpc.RPCClient) (*Client, error) {
	if rpcClient == nil || rpcClient.Conn == nil {
		return nil, fmt.Errorf("grpc RPC client no initialized (nil connection)")
	}

	return &Client{
		rpcClient:                rpcClient,
		leaderboardScoringClient: leaderboardscoringpb.NewLeaderboardScoringServiceClient(rpcClient.Conn),
	}, nil
}

func (c *Client) GetLeaderboard(ctx context.Context, getLeaderboardReq *lbscoring.GetLeaderboardRequest) (*lbscoring.GetLeaderboardResponse, error) {

	leaderboardPBReq := &leaderboardscoringpb.GetLeaderboardRequest{
		Timeframe: lbscoring.ToProtoTimeframe(getLeaderboardReq.Timeframe),
		ProjectId: getLeaderboardReq.ProjectID,
		PageSize:  getLeaderboardReq.PageSize,
		Offset:    getLeaderboardReq.Offset,
	}

	leaderboardPBRes, err := c.leaderboardScoringClient.GetLeaderboard(ctx, leaderboardPBReq)
	if err != nil {
		return nil, err
	}

	getLeaderboardRes := protobufToLeaderboardRes(leaderboardPBRes)

	return getLeaderboardRes, nil
}

func protobufToLeaderboardRes(leaderboardPBRes *leaderboardscoringpb.GetLeaderboardResponse) *lbscoring.GetLeaderboardResponse {
	var rows = make([]lbscoring.LeaderboardRow, 0, len(leaderboardPBRes.Rows))
	for _, r := range leaderboardPBRes.Rows {
		row := lbscoring.LeaderboardRow{
			Rank:   r.Rank,
			UserID: r.UserId,
			Score:  int64(r.Score),
		}
		rows = append(rows, row)
	}

	var getLeaderboardRes = &lbscoring.GetLeaderboardResponse{
		Timeframe:       lbscoring.FromProtoTimeframe(leaderboardPBRes.Timeframe),
		ProjectID:       leaderboardPBRes.ProjectId,
		LeaderboardRows: rows,
	}
	return getLeaderboardRes
}

func (c *Client) Close() {
	if c.rpcClient != nil {
		c.rpcClient.Close()
	}
}
