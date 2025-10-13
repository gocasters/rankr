package leaderboardscoring

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
)

type Validator struct{}

func NewValidator() Validator {
	return Validator{}
}

// ValidateEvent checks if the domain event has valid data.
func (v Validator) ValidateEvent(event *EventRequest) error {

	// TODO: This validation needs to be completed for all event attributes.
	return validation.ValidateStruct(event,
		validation.Field(&event.ID, validation.Required, is.UUID),
		validation.Field(&event.EventName, validation.Required, validation.In(
			PullRequestOpened.String(),
			PullRequestClosed.String(),
			PullRequestReview.String(),
			IssueOpened.String(),
			IssueClosed.String(),
			IssueComment.String(),
			CommitPush.String(),
		)),
		validation.Field(&event.RepositoryID, validation.Required),
		validation.Field(&event.RepositoryName, validation.Required),
		validation.Field(&event.Timestamp, validation.Required),
	)
}

func (v Validator) ValidateGetLeaderboard(request *GetLeaderboardRequest) error {
	// TODO - Implement validation for GetLeaderboardRequest
	return validation.ValidateStruct(request)
}
