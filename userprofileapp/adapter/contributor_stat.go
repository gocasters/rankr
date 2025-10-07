package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorStatRPC struct{}

func NewContributorStatRPC() ContributorStatRPC {
	return ContributorStatRPC{}
}

// getContributorStat fetches contributor stat information
// TODO: Implement actual RPC call to contributorstatapp
func (a ContributorStatRPC) getContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	// Placeholder implementation

	return userprofile.ContributorStat{}, nil
}
