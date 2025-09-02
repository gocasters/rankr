package consumer

import (
	"errors"
	"fmt"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/protobuf/golang/eventpb"
	"google.golang.org/protobuf/proto"
	"log/slog"
)

type Handler struct {
	leaderboardSvc     leaderboardscoring.Service
	idempotencyChecker *IdempotencyChecker
	logger             *slog.Logger
}

func NewHandler(svc leaderboardscoring.Service, checker *IdempotencyChecker, logger *slog.Logger) Handler {
	return Handler{
		leaderboardSvc:     svc,
		idempotencyChecker: checker,
		logger:             logger,
	}
}

func (h Handler) HandleEvent(msg *message.Message) error {
	var event eventpb.Event
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.logger.Error(
			"Failed to unmarshal event payload",
			slog.String("msg_id", msg.UUID),
			slog.String("error", err.Error()),
		)
		return err // NACK
	}

	eventReq, rErr := protobufToEventRequest(&event)
	if rErr != nil {
		return rErr // NACK
	}

	processEventFunc := func() error {
		return h.leaderboardSvc.ProcessScoreEvent(msg.Context(), eventReq)
	}

	bufferedEventFunc := func() error {
		return h.leaderboardSvc.QueueEventForPersistence(msg.Context(), msg.Payload)
	}

	// Wrap the business logic with the idempotency check.
	err := h.idempotencyChecker.CheckEvent(msg.Context(), eventReq.ID, processEventFunc, bufferedEventFunc)

	switch {
	case err == nil:
		// Success
		h.logger.Debug("Event processed successfully", slog.String("event_id", eventReq.ID))
		return nil // ACK
	case errors.Is(err, ErrEventAlreadyProcessed):
		// This is not an error, it's an expected outcome for duplicates.
		h.logger.Info("Event skipped: already processed", slog.String("event_id", eventReq.ID))
		return nil // ACK
	case errors.Is(err, ErrEventLocked):
		h.logger.Info("Event skipped: currently locked by another processor", slog.String("event_id", eventReq.ID))
		return nil // we ACK to remove it, as another worker is responsible.
	default:
		h.logger.Error(
			"Error processing event",
			slog.String("event_id", eventReq.ID),
			slog.String("error", err.Error()),
		)
		return err // NACK and requeue
	}
}

func protobufToEventRequest(eventPB *eventpb.Event) (*leaderboardscoring.EventRequest, error) {
	if eventPB == nil {
		return &leaderboardscoring.EventRequest{}, fmt.Errorf("nil event")
	}

	ts := eventPB.GetTime()
	if ts == nil {
		return &leaderboardscoring.EventRequest{}, fmt.Errorf("event with ID %s has a missing timestamp", eventPB.Id)
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

	contributionEvent := &leaderboardscoring.EventRequest{
		ID:             eventPB.Id,
		EventName:      leaderboardscoring.EventType(eventPB.GetEventName()),
		RepositoryID:   eventPB.RepositoryId,
		RepositoryName: eventPB.RepositoryName,
		ContributorID:  contribID,
		Timestamp:      ts.AsTime().UTC(),
	}

	return contributionEvent, nil
}
