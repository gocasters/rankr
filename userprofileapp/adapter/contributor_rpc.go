package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type contributorInfoRPC struct{}

func newContributorInfoRPC() contributorInfoRPC {
	return contributorInfoRPC{}
}

// getProfileInfo fetches contributor profile information
// TODO: Implement actual RPC call to contributorapp
func (a contributorInfoRPC) getProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	// Placeholder implementation

	return userprofile.ContributorInfo{}, nil
}
