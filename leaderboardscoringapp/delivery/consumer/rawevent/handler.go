package rawevent

import (
	"errors"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"google.golang.org/protobuf/proto"
	"log/slog"
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

func (h Handler) HandleEvent(msg *message.Message) error {
	logger := logger.L()
	var event eventpb.Event
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		logger.Error(
			"Failed to unmarshal event payload",
			slog.String("msg_id", msg.UUID),
			slog.String("error", err.Error()),
		)
		return err // NACK
	}

	eventReq, rErr := leaderboardscoring.NewEventRequest().ProtobufToEventRequest(&event)
	if rErr != nil {
		return rErr // NACK
	}

	processEventFunc := func() error {
		return h.leaderboardSvc.ProcessScoreEvent(msg.Context(), eventReq)
	}

	// Wrap the business logic with the idempotency check.
	err := h.idempotencyChecker.CheckEvent(msg.Context(), eventReq.ID, processEventFunc)

	switch {
	case err == nil:
		// Success
		logger.Debug("Event processed successfully", slog.String("event_id", eventReq.ID))
		return nil // ACK

	case errors.Is(err, ErrEventAlreadyProcessed):
		// This is not an error, it's an expected outcome for duplicates.
		logger.Info("Event skipped: already processed", slog.String("event_id", eventReq.ID))
		return nil // ACK

	case errors.Is(err, ErrEventLocked):
		logger.Info("Event skipped: currently locked by another processor", slog.String("event_id", eventReq.ID))
		return nil // we ACK to remove it, as another worker is responsible.

	default:
		logger.Error(
			"Error processing event",
			slog.String("event_id", eventReq.ID),
			slog.String("error", err.Error()),
		)
		return err // NACK and requeue
	}
}
