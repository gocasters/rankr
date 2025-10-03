package adapter

import (
	"context"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type Adapter struct{}

func NewAdapter() Adapter {
	return Adapter{}
}

func (a Adapter) GetProfileInfo(ctx context.Context, userID int64) (userprofile.ContributorInfo, error) {
	panic("implement me")
}

func (a Adapter) GetTasks(ctx context.Context, userID int64) ([]userprofile.Task, error) {
	panic("implement me")
}

func (a Adapter) GetUserStat(ctx context.Context, userID int64) (userprofile.ContributorStat, error) {
	panic("impelement me")
}
