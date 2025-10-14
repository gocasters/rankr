package leaderboardscoring

import (
	"fmt"
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
		).Error(fmt.Sprintf("EventName must be one of: %s, %s, %s, %s, %s, %s, %s",
			PullRequestOpened.String(),
			PullRequestClosed.String(),
			PullRequestReview.String(),
			IssueOpened.String(),
			IssueClosed.String(),
			IssueComment.String(),
			CommitPush.String()))),

		validation.Field(&event.RepositoryID, validation.Required),

		validation.Field(&event.RepositoryName, validation.Required),

		validation.Field(&event.Timestamp, validation.Required),
	)
}

func (v Validator) ValidateGetLeaderboard(request *GetLeaderboardRequest) error {
	return validation.ValidateStruct(request,
		validation.Field(&request.Timeframe,
			validation.Required.Error("timeframe is required"),
			validation.In(
				AllTime.String(),
				Yearly.String(),
				Monthly.String(),
				Weekly.String(),
			).Error(fmt.Sprintf("timeframe must be one of: %s, %s, %s, %s",
				AllTime.String(),
				Yearly.String(),
				Monthly.String(),
				Weekly.String()))),

		validation.Field(&request.Offset,
			validation.Required.Error("offset is required"),
			validation.Min(int32(minOffset)).Error("offset cannot be negative"),
			validation.Max(int32(maxOffset)).Error(fmt.Sprintf("offset cannot exceed %d", maxOffset)),
		),

		validation.Field(&request.PageSize,
			validation.Required.Error("page_size is required"),
			validation.Min(int32(minPageSize)).Error(fmt.Sprintf("page_size must be at least %d", minPageSize)),
			validation.Max(int32(maxPageSize)).Error(fmt.Sprintf("page_size cannot exceed %d", maxPageSize)),
		),
	)
}
