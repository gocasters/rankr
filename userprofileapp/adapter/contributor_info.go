package adapter

import (
	"context"
	"github.com/gocasters/rankr/adapter/contributor"
	types "github.com/gocasters/rankr/type"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
)

type ContributorClient interface {
	GetContributorInfo(ctx context.Context, id types.ID) (contributor.InfoResponse, error)
}

type ContributorAdapter struct {
	contributorClient ContributorClient
}

func NewContributorAdapter(contributorClient ContributorClient) ContributorAdapter {
	return ContributorAdapter{contributorClient: contributorClient}
}

func (a ContributorAdapter) GetProfileInfo(ctx context.Context, contributorID int64) (userprofile.ContributorInfo, error) {

	res, err := a.contributorClient.GetContributorInfo(ctx, types.ID(contributorID))
	if err != nil {
		return userprofile.ContributorInfo{}, err
	}

	return userprofile.ContributorInfo{
		ID:             res.ID,
		GitHubID:       res.GitHubID,
		GitHubUsername: res.GitHubUsername,
		ProfileImage:   &res.ProfileImage,
		PrivacyMode:    res.PrivacyMode,
		DisplayName:    &res.DisplayName,
		Bio:            &res.Bio,
		CreatedAt:      res.CreatedAt,
	}, nil
}
