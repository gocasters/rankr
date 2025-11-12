package historical

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/gocasters/rankr/adapter/webhook/github"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func sanitizeUTF8(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	return strings.ToValidUTF8(s, "")
}

func TransformPRToEvents(pr *github.PullRequest) ([]*eventpb.Event, error) {
	events := []*eventpb.Event{}

	openedEvent := &eventpb.Event{
		Id:             fmt.Sprintf("historical-pr-%d-opened", pr.Number),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_OPENED,
		Time:           timestamppb.New(pr.CreatedAt),
		RepositoryId:   pr.Base.Repo.ID,
		RepositoryName: sanitizeUTF8(pr.Base.Repo.FullName),
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrOpenedPayload{
			PrOpenedPayload: &eventpb.PullRequestOpenedPayload{
				UserId:       pr.User.ID,
				PrId:         pr.ID,
				PrNumber:     pr.Number,
				Title:        sanitizeUTF8(pr.Title),
				BranchName:   sanitizeUTF8(pr.Head.Ref),
				TargetBranch: sanitizeUTF8(pr.Base.Ref),
				Labels:       extractLabels(pr.Labels),
				Assignees:    extractAssignees(pr.Assignees),
			},
		},
	}
	events = append(events, openedEvent)

	if pr.State == "closed" && pr.ClosedAt != nil {
		var mergerID uint64
		if pr.MergedBy != nil {
			mergerID = pr.MergedBy.ID
		} else {
			mergerID = pr.User.ID
		}

		closeReason := eventpb.PrCloseReason_PR_CLOSE_REASON_CLOSED_WITHOUT_MERGE
		if pr.Merged {
			closeReason = eventpb.PrCloseReason_PR_CLOSE_REASON_MERGED
		}

		closedEvent := &eventpb.Event{
			Id:             fmt.Sprintf("historical-pr-%d-closed", pr.Number),
			EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_CLOSED,
			Time:           timestamppb.New(*pr.ClosedAt),
			RepositoryId:   pr.Base.Repo.ID,
			RepositoryName: sanitizeUTF8(pr.Base.Repo.FullName),
			Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
			Payload: &eventpb.Event_PrClosedPayload{
				PrClosedPayload: &eventpb.PullRequestClosedPayload{
					UserId:       pr.User.ID,
					MergerUserId: mergerID,
					PrId:         pr.ID,
					PrNumber:     pr.Number,
					CloseReason:  closeReason,
					Merged:       &pr.Merged,
					Additions:    pr.Additions,
					Deletions:    pr.Deletions,
					FilesChanged: pr.ChangedFiles,
					CommitsCount: pr.Commits,
					Labels:       extractLabels(pr.Labels),
					TargetBranch: sanitizeUTF8(pr.Base.Ref),
					Assignees:    extractAssignees(pr.Assignees),
				},
			},
		}
		events = append(events, closedEvent)
	}

	return events, nil
}

func TransformReviewToEvent(review *github.Review, pr *github.PullRequest) (*eventpb.Event, error) {
	reviewState := mapReviewState(review.State)

	event := &eventpb.Event{
		Id:             fmt.Sprintf("historical-pr-%d-review-%d", pr.Number, review.ID),
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
		Time:           timestamppb.New(review.SubmittedAt),
		RepositoryId:   pr.Base.Repo.ID,
		RepositoryName: sanitizeUTF8(pr.Base.Repo.FullName),
		Provider:       eventpb.EventProvider_EVENT_PROVIDER_GITHUB,
		Payload: &eventpb.Event_PrReviewPayload{
			PrReviewPayload: &eventpb.PullRequestReviewSubmittedPayload{
				ReviewerUserId: review.User.ID,
				PrAuthorUserId: pr.User.ID,
				PrId:           pr.ID,
				PrNumber:       pr.Number,
				State:          reviewState,
			},
		},
	}

	return event, nil
}

func extractLabels(labels []github.Label) []string {
	result := make([]string, len(labels))
	for i, label := range labels {
		result[i] = sanitizeUTF8(label.Name)
	}
	return result
}

func extractAssignees(assignees []github.User) []uint64 {
	result := make([]uint64, len(assignees))
	for i, assignee := range assignees {
		result[i] = assignee.ID
	}
	return result
}

func mapReviewState(state string) eventpb.ReviewState {
	switch state {
	case "APPROVED":
		return eventpb.ReviewState_REVIEW_STATE_APPROVED
	case "CHANGES_REQUESTED":
		return eventpb.ReviewState_REVIEW_STATE_CHANGES_REQUESTED
	case "COMMENTED":
		return eventpb.ReviewState_REVIEW_STATE_COMMENTED
	default:
		return eventpb.ReviewState_REVIEW_STATE_UNSPECIFIED
	}
}
