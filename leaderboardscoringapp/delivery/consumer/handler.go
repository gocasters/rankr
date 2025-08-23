package consumer

import (
	"context"
	"errors"
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

func (h Handler) HandleContributionRegistered(msg *message.Message) error {
	var event eventpb.Event
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		h.logger.Error(
			"Failed to unmarshal event payload",
			slog.String("msg_id", msg.UUID),
			slog.String("error", err.Error()),
		)
		return err // NACK
	}

	processFunc := func() error {
		eventReq, err := ProtobufToEventRequest(&event)
		if err != nil {
			return err
		}

		return h.leaderboardSvc.ProcessScoreEvent(context.Background(), eventReq)
	}

	// Wrap the business logic with the idempotency check.
	err := h.idempotencyChecker.Process(context.Background(), event.Id, processFunc)

	switch {
	case err == nil:
		// Success
		h.logger.Debug("Event processed successfully", slog.String("event_id", event.Id))
		return nil // ACK
	case errors.Is(err, ErrEventAlreadyProcessed):
		// This is not an error, it's an expected outcome for duplicates.
		h.logger.Info("Event skipped: already processed", slog.String("event_id", event.Id))
		return nil // ACK
	case errors.Is(err, ErrEventLocked):
		h.logger.Info("Event skipped: currently locked by another processor", slog.String("event_id", event.Id))
		return nil // we ACK to remove it, as another worker is responsible.
	default:
		h.logger.Error(
			"Error processing event",
			slog.String("event_id", event.Id),
			slog.String("error", err.Error()),
		)
		return err // NACK and requeue
	}
}
