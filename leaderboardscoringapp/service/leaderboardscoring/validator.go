package leaderboardscoring

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Validator struct{}

func NewValidator() Validator {
	return Validator{}
}

// ValidateContributionEvent checks if the domain event has valid data.
func (v Validator) ValidateContributionEvent(event EventRequest) error {
	// TODO: This validation needs to be completed for all event attributes.
	return validation.ValidateStruct(event,
		// Example validation rules can be added here, for instance:
		validation.Field(&event.ID, validation.Required, is.UUID),
		validation.Field(&event.ContributorID, validation.Required),
		validation.Field(&event.RepositoryID, validation.Required),
		//validation.Field(&event.Type, validation.Required,
		//	validation.In(string(ContributionCommit),
		//		string(ContributionReview),
		//		string(ContributionIssueClosed),
		//	),
		//),
	)
}
