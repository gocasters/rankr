package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorInfoRPC struct{}

func NewContributorInfoRPC() ContributorInfoRPC {
	return ContributorInfoRPC{}
}

// getProfileInfo fetches contributor profile information
// TODO: Implement actual RPC call to contributorapp
func (a ContributorInfoRPC) getProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	// Placeholder implementation

	return userprofile.ContributorInfo{}, nil
}
