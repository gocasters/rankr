package job

import (
	"context"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/pkg/validator"
)

type ContributorRepo interface {
	Upsert(ctx context.Context, contributor contributor.UpsertContributorRequest) (contributor.UpsertContributorResponse, error)
}

type ContributorAdapter struct {
	contributorSvc ContributorRepo
}

func NewContributorAdapter(contrSvc ContributorRepo) ContributorAdapter {
	return ContributorAdapter{contributorSvc: contrSvc}
}

func (c ContributorAdapter) UpsertContributor(ctx context.Context, req ContributorRecord) error {
	_, err := c.contributorSvc.Upsert(ctx, req.mapContributorRecordToUpsertRequest())
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return RecordProcessError{Err: vErr, Type: ErrTypeValidation}
		}

		return RecordProcessError{Err: err, Type: ErrTypeUnexpected}
	}

	return nil
}
