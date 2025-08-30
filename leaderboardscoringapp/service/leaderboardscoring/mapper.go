package leaderboardscoring

import (
	"fmt"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
)

func ProtobufToEventRequest(eventPB *eventpb.Event) (*EventRequest, error) {
	if eventPB == nil {
		return &EventRequest{}, fmt.Errorf("nil event")
	}

	ts := eventPB.GetTime()
	if ts == nil {
		return &EventRequest{}, fmt.Errorf("event with ID %s has a missing timestamp", eventPB.Id)
	}

	var contribID uint64 = 0

	switch eventPB.EventName.Type() {
	case eventpb.EventName_PULL_REQUEST_OPENED.Type():
		contribID = eventPB.GetPrOpenedPayload().GetUserId()
	case eventpb.EventName_PULL_REQUEST_CLOSED.Type():
		contribID = eventPB.GetPrClosedPayload().GetUserId()
	case eventpb.EventName_PULL_REQUEST_REVIEW_SUBMITTED.Type():
		contribID = eventPB.GetPrReviewPayload().GetReviewerUserId()
	case eventpb.EventName_ISSUE_OPENED.Type():
		contribID = eventPB.GetIssueOpenedPayload().GetUserId()
	case eventpb.EventName_ISSUE_CLOSED.Type():
		contribID = eventPB.GetIssueClosedPayload().GetUserId()
	case eventpb.EventName_ISSUE_COMMENTED.Type():
		contribID = eventPB.GetIssueCommentedPayload().GetUserId()
	case eventpb.EventName_COMMIT_PUSHED.Type():
		contribID = eventPB.GetPushPayload().GetUserId()
	}

	contributionEvent := &EventRequest{
		ID:             eventPB.Id,
		EventName:      EventType(eventPB.GetEventName()),
		RepositoryID:   eventPB.RepositoryId,
		RepositoryName: eventPB.RepositoryName,
		ContributorID:  contribID,
		Timestamp:      ts.AsTime().UTC(),
	}

	return contributionEvent, nil
}
