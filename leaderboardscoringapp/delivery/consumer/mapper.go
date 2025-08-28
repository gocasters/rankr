package consumer

import (
	"fmt"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
)

func ProtobufToEventRequest(eventPB *eventpb.Event) (leaderboardscoring.EventRequest, error) {
	if eventPB == nil {
		return leaderboardscoring.EventRequest{}, fmt.Errorf("nil event")
	}

	//payload := eventPB.GetContributionRegisteredPayload()
	//if payload == nil {
	//	return leaderboardscoring.EventRequest{},
	//		fmt.Errorf("event with ID %s has a missing or invalid payload", eventPB.Id)
	//}

	ts := eventPB.GetTime()
	if ts == nil {
		return leaderboardscoring.EventRequest{}, fmt.Errorf("event with ID %s has a missing timestamp", eventPB.Id)
	}
	contributionEvent := leaderboardscoring.EventRequest{
		ID:              eventPB.Id,
		EventName:       string(eventPB.EventName),
		RepositoryID:    eventPB.RepositoryId,
		RepositoryName:  eventPB.RepositoryName,
		SourceReference: "",
		Timestamp:       ts.AsTime().UTC(),
	}

	return contributionEvent, nil
}
