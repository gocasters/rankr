package delivery

import (
	"context"
	"encoding/json"
	"fmt"

	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (s *Service) HandlePullRequestReviewEvent(provider eventpb.EventProvider, action string, body []byte, deliveryUID string) error {
	switch action {
	case "submitted":
		var reviewData PullRequestReviewSubmittedRequest
		if err := json.Unmarshal(body, &reviewData); err != nil {
			return err
		}
		return s.PublishPullRequestReviewSubmitted(reviewData, provider, deliveryUID)

	default:
		return fmt.Errorf("pull request review action '%s' not handled", action)
	}
}

func (s *Service) PublishPullRequestReviewSubmitted(req PullRequestReviewSubmittedRequest, provider eventpb.EventProvider, deliveryUID string) error {
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

	ctx := context.Background()
	return s.saveEvent(ctx, ev)
	//return s.publishEvent(ev, eventpb.EventName_EVENT_NAME_PULL_REQUEST_REVIEW_SUBMITTED, TopicGithubReview, metadata)
}
