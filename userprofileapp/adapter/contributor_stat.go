package adapter

import (
	"context"
	"fmt"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type contributorStatRPC struct{}

func newContributorStatRPC() contributorStatRPC {
	return contributorStatRPC{}
}

func (a contributorStatRPC) GetContributorStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	// TODO: Implement me

	return userprofile.ContributorStat{}, fmt.Errorf("implement me. userprofileapp/adapter/contributor_rpc.go GetContributorStat")
}
