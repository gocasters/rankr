package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type contributorInfoRPC struct{}

func newContributorInfoRPC() contributorInfoRPC {
	return contributorInfoRPC{}
}

func (a contributorInfoRPC) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	// TODO: Implement me

	return userprofile.ContributorInfo{}, fmt.Errorf("implement me. userprofileapp/adapter/contributor_rpc.go GetProfileInfo")
}
