package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type contributorStatRPC struct{}

func newContributorStatRPC() contributorStatRPC {
	return contributorStatRPC{}
}

// GetContributorStat fetches contributor stat information
// TODO: Implement actual RPC call to contributorstatapp
func (a contributorStatRPC) GetContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	// Placeholder implementation

	return userprofile.ContributorStat{}, nil
}
