package leaderboardstat

import (
	"context"
	"fmt"
	lbstat "github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/gocasters/rankr/pkg/grpc"
	"github.com/gocasters/rankr/pkg/slice"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	types "github.com/gocasters/rankr/type"
	//"google.golang.org/protobuf/types/known/timestamppb"
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

func (c *Client) GetContributorStats(ctx context.Context, contributorStatsReq *lbstat.ContributorStatsRequest) (*lbstat.ContributorStatsResponse, error) {
	statPbReq := &leaderboardstatpb.ContributorStatRequest{
		ContributorId: uint64(contributorStatsReq.ContributorID),
	}

	statPbRes, err := c.leaderboardStatClient.GetContributorStats(ctx, statPbReq)
	if err != nil {
		return nil, err
	}

	return &lbstat.ContributorStatsResponse{
		ContributorID: types.ID(statPbRes.ContributorId),
		GlobalRank:    int(statPbRes.GlobalRank),
		TotalScore:    statPbRes.TotalScore,
		ProjectsScore: slice.MapFromUint64Float64ToIDFloat64(statPbRes.ProjectsScore),
		ScoreHistory:  convertScoreHistory(statPbRes.ScoreHistory),
	}, nil
}

func convertScoreHistory(pbMap map[uint64]*leaderboardstatpb.ProjectScoreHistory) map[types.ID][]lbstat.ScoreEntry {
	result := make(map[types.ID][]lbstat.ScoreEntry, len(pbMap))

	for projectID, history := range pbMap {
		entries := make([]lbstat.ScoreEntry, 0, len(history.Entries))
		for _, e := range history.Entries {
			entries = append(entries, lbstat.ScoreEntry{
				Activity: e.Activity,
				Score:    e.Score,
				EarnedAt: e.EarnedAt.AsTime(),
			})
		}
		result[types.ID(projectID)] = entries
	}

	return result
}
