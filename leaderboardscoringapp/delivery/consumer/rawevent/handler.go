package rawevent

import (
	"errors"
	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	leaderboardSvc     *leaderboardscoring.Service
	idempotencyChecker *IdempotencyChecker
}

func NewHandler(svc *leaderboardscoring.Service, checker *IdempotencyChecker) Handler {
	return Handler{
		leaderboardSvc:     svc,
		idempotencyChecker: checker,
	}
}

// isPermanent reports errors that fail the same way on every redelivery.
// Such events are ACKed instead of retried forever: the raw event is kept in
// webhook_db, so it can be replayed after the cause is fixed.
func isPermanent(err error) bool {
	return errors.Is(err, leaderboardscoring.ErrInvalidEventRequest)
}

func (h Handler) HandleEvent(msg *message.Message) error {
	logger := logger.L()
	var event eventpb.Event
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		logger.Warn(
			"Failed to unmarshal event payload",
			slog.String("msg_id", msg.UUID),
			slog.String("error", err.Error()),
		)
		return nil // ACK: a retry cannot decode it either
	}

	eventReq, rErr := leaderboardscoring.NewEventRequest().MapProtoEventToEventRequest(&event)
	if rErr != nil {
		logger.Warn(
			"Failed to map event payload",
			slog.String("event_id", event.GetId()),
			slog.String("error", rErr.Error()),
		)
		return nil // ACK: a retry cannot map it either
	}

	processEventFunc := func() error {
		return h.leaderboardSvc.ProcessScoreEvent(msg.Context(), eventReq)
	}

	// Wrap the business logic with the idempotency check.
	err := h.idempotencyChecker.CheckEvent(msg.Context(), eventReq.ID, processEventFunc)
	return handleProcessingResult(eventReq.ID, err)
}

func handleProcessingResult(eventID string, err error) error {
	log := logger.L()
	switch {
	case err == nil:
		// Success
		log.Debug("Event processed successfully", slog.String("event_id", eventID))
		return nil // ACK

	case errors.Is(err, ErrEventAlreadyProcessed):
		// This is not an error, it's an expected outcome for duplicates.
		log.Info("Event skipped: already processed", slog.String("event_id", eventID))
		return nil // ACK

	case errors.Is(err, ErrEventLocked):
		log.Info("Event skipped: currently locked by another processor", slog.String("event_id", eventID))
		return err // NACK; the other processor may fail.

	case isPermanent(err):
		log.Warn(
			"Event dropped: invalid event",
			slog.String("event_id", eventID),
			slog.String("error", err.Error()),
		)
		return nil // ACK: see isPermanent

	default:
		log.Error(
			"Error processing event",
			slog.String("event_id", eventID),
			slog.String("error", err.Error()),
		)
		return err // NACK and requeue
	}
}
