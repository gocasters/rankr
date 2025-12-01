package dashboard

import (
	"context"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
)

type ContributorAdapter struct {
	contributorSvc contributor.Service
}

func NewContributorAdapter(contrSvc contributor.Service) ContributorAdapter {
	return ContributorAdapter{contributorSvc: contrSvc}
}

func (c ContributorAdapter) Upsert(ctx context.Context, contri contributor.Contributor) error {
	exists, err := c.contributorSvc.GetContributorByGithubUsername(ctx, contri.GitHubUsername)
	if err != nil {
		return err
	}

	if !exists {
		var createContributor contributor.CreateContributorRequest

		createContributor.GitHubUsername = contri.GitHubUsername
		createContributor.GitHubID = contri.GitHubID
		createContributor.PrivacyMode = contri.PrivacyMode
		createContributor.Bio = contri.Bio
		createContributor.ProfileImage = contri.ProfileImage
		createContributor.DisplayName = contri.DisplayName

		_, err := c.contributorSvc.CreateContributor(ctx, createContributor)
		if err != nil {
			return err
		}

		return nil
	}

	var updateContributor contributor.UpdateProfileRequest

	updateContributor.GitHubID = contri.GitHubID
	updateContributor.GitHubUsername = contri.GitHubUsername
	updateContributor.DisplayName = contri.DisplayName
	updateContributor.Bio = contri.Bio
	updateContributor.PrivacyMode = contri.PrivacyMode
	updateContributor.ProfileImage = contri.ProfileImage

	_, err = c.contributorSvc.UpdateProfile(ctx, updateContributor)
	if err != nil {
		return err
	}

	return nil
}
