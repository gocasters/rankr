package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandlePullRequestReviewEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) error {
	ev, err := buildPullRequestReviewEvent(provider, action, body, deliveryUID)
	if err != nil {
		return err
	}
	return s.saveEvent(context.Background(), ev)
}

func buildPullRequestReviewEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) (*eventpb.Event, error) {
	switch action {
	case "submitted":
		var reviewData PullRequestReviewSubmittedRequest
		if err := json.Unmarshal(body, &reviewData); err != nil {
			return nil, err
		}
		return newPullRequestReviewSubmittedEvent(reviewData, provider, deliveryUID), nil

	default:
		return nil, fmt.Errorf("pull request review action '%s' not handled", action)
	}
}

func newPullRequestReviewSubmittedEvent(req PullRequestReviewSubmittedRequest, provider eventpb.EventProvider, deliveryUID string) *eventpb.Event {
	ev := &eventpb.Event{
		Id:             deliveryUID,
		EventName:      eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED,
		Provider:       provider,
		Time:           timestamppb.New(req.Review.SubmittedAt),
		RepositoryId:   req.Repository.ID,
		RepositoryName: req.Repository.FullName,
		Payload: &eventpb.Event_PrReviewPayload{
			PrReviewPayload: &eventpb.PullRequestReviewSubmittedPayload{
				ReviewerUserId: req.Review.User.ID,
				PrAuthorUserId: req.PullRequest.User.ID,
				PrId:           req.PullRequest.ID,
				PrNumber:       req.PullRequest.Number,
				State:          getReviewState(req.Review.State),
			},
		},
	}

	return ev
}
