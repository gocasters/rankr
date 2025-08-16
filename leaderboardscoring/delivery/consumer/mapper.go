package consumer

import (
	"fmt"
	"github.com/gocasters/rankr/leaderboardscoring/service/leaderboardscoring"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
)

func ProtobufToEventRequest(eventPB *eventpb.Event) (leaderboardscoring.EventRequest, error) {
	payload := eventPB.GetContributionRegisteredPayload()
	if payload == nil {
		return leaderboardscoring.EventRequest{},
			fmt.Errorf("event with ID %s has a missing or invalid payload", eventPB.Id)
	}

	contributionEvent := leaderboardscoring.EventRequest{
		ID:              eventPB.Id,
		UserID:          payload.UserId,
		ProjectID:       payload.ProjectId,
		Type:            payload.EventType,
		ScoreValue:      int(payload.ScoreValue),
		SourceReference: payload.SourceReference,
		Timestamp:       eventPB.GetTime().AsTime(),
	}

	return contributionEvent, nil
}
