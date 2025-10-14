package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorStatRPC struct{}

func NewContributorStatRPC() ContributorStatRPC {
	return ContributorStatRPC{}
}

// GetContributorStat fetches contributor stat information
// TODO: Implement actual RPC call to contributor stat app
func (a ContributorStatRPC) GetContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	// Placeholder implementation

	return userprofile.ContributorStat{}, nil
}
