package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorRPC struct{}

func NewContributorRPC() ContributorRPC {
	return ContributorRPC{}
}

func (a ContributorRPC) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	// TODO: Implement me
	return userprofile.ContributorInfo{}, fmt.Errorf("implement me. userprofileapp/adapter/contributor_rpc.go GetProfileInfo")
}
