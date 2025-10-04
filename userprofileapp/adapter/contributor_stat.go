package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorStatRPC struct{}

func NewContributorStatRPC() ContributorStatRPC {
	return ContributorStatRPC{}
}

func (a ContributorStatRPC) GetUserStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	// TODO: Implement me

	return userprofile.ContributorStat{}, fmt.Errorf("implement me. userprofileapp/adapter/contributor_rpc.go GetUserStat")
}
