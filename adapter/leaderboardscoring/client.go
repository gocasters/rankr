package leaderboardscoring

import (
	"context"
	lbscoring "github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/protobuf/leaderboardscoring/golang/leaderboardscoringpb"
	"log/slog"
)

type Client struct {
	rpcClient                *grpc.RPCClient
	leaderboardScoringClient leaderboardscoringpb.LeaderboardScoringServiceClient
	logger                   *slog.Logger
}

func New(rpcClient *grpc.RPCClient, logger *slog.Logger) Client {
	return Client{
		rpcClient:                rpcClient,
		leaderboardScoringClient: leaderboardscoringpb.NewLeaderboardScoringServiceClient(rpcClient.Conn),
		logger:                   logger,
	}
}

func (c *Client) GetLeaderboard(ctx context.Context, getLeaderboardReq *lbscoring.GetLeaderboardRequest) (*lbscoring.GetLeaderboardResponse, error) {

	leaderboardPBReq := &leaderboardscoringpb.GetLeaderboardRequest{
		Timeframe: leaderboardscoringpb.Timeframe(getLeaderboardReq.Timeframe),
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
			Score:  r.Score,
		}
		rows = append(rows, row)
	}

	var getLeaderboardRes = &lbscoring.GetLeaderboardResponse{
		Timeframe:       lbscoring.Timeframe(leaderboardPBRes.GetTimeframe()),
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
