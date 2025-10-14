package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorInfoRPC struct{}

func NewContributorInfoRPC() ContributorInfoRPC {
	return ContributorInfoRPC{}
}

// GetProfileInfo fetches contributor profile information
// TODO: Implement actual RPC call to contributor app
func (a ContributorInfoRPC) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	// Placeholder implementation

	return userprofile.ContributorInfo{}, nil
}
