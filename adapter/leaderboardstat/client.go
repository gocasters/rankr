package leaderboardstat

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/pkg/grpc"
	leaderboardstatpb "github.com/gocasters/rankr/protobuf/golang/leaderboardstat"
	types "github.com/gocasters/rankr/type"
	"time"
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

type GetContributorStatRequest struct {
	ContributorID types.ID
}

type GetContributorStatResponse struct {
	ContributorID types.ID
	GlobalRank    int
	TotalScore    float64
	ProjectsScore map[types.ID]float64
	ScoreHistory  map[types.ID][]ScoreEntry
}

type ScoreEntry struct {
	Activity string
	Score    float64
	EarnedAt time.Time
}

func (c *Client) GetContributorStats(ctx context.Context, req GetContributorStatRequest) (GetContributorStatResponse, error) {
	statPbReq := &leaderboardstatpb.ContributorStatRequest{
		ContributorId: uint64(req.ContributorID),
	}

	statPbRes, err := c.leaderboardStatClient.GetContributorStats(ctx, statPbReq)
	if err != nil {
		return GetContributorStatResponse{}, err
	}

	projectsScore := make(map[types.ID]float64, len(statPbRes.ProjectsScore))

	for k, v := range statPbRes.ProjectsScore {
		projectsScore[types.ID(k)] = v
	}

	scoreHistory := make(map[types.ID][]ScoreEntry, len(statPbRes.ScoreHistory))
	for i, sh := range statPbRes.ScoreHistory {
		scoreEntries := make([]ScoreEntry, 0, len(sh.Entries))
		for _, se := range sh.Entries {
			earnedAt := se.GetEarnedAt()
			scoreEntries = append(scoreEntries, ScoreEntry{
				Score:    se.Score,
				Activity: se.Activity,
				EarnedAt: earnedAt.AsTime(),
			})
		}

		scoreHistory[types.ID(i)] = scoreEntries
	}

	return GetContributorStatResponse{
		ContributorID: types.ID(statPbRes.ContributorId),
		GlobalRank:    int(statPbRes.GlobalRank),
		TotalScore:    statPbRes.TotalScore,
		ProjectsScore: projectsScore,
		ScoreHistory:  scoreHistory,
	}, nil
}
