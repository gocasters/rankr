package consumer

import (
	"context"

	"log/slog"

	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/gocasters/rankr/pkg/logger"
	eventpb "github.com/gocasters/rankr/protobuf/golang/event/v1"
	"github.com/gocasters/rankr/taskapp/service/task"
	"google.golang.org/protobuf/proto"
)

type Handler struct {
	taskSvc *task.Service
}

func NewHandler(svc *task.Service) Handler {
	return Handler{
		taskSvc: svc,
	}
}

func (h Handler) HandleEvent(msg *message.Message) error {
	log := logger.L()
	var event eventpb.Event
	if err := proto.Unmarshal(msg.Payload, &event); err != nil {
		log.Error(
			"Failed to unmarshal event payload",
			slog.String("msg_id", msg.UUID),
			slog.String("error", err.Error()),
		)
		return err
	}

	log.Info("Received task event",
		slog.String("event_id", event.Id),
		slog.String("event_name", event.EventName.String()),
		slog.String("repository", event.RepositoryName),
	)

	switch event.EventName {
	case eventpb.EventName_EVENT_NAME_ISSUE_OPENED:
		return h.handleIssueOpened(msg.Context(), &event)
	case eventpb.EventName_EVENT_NAME_ISSUE_CLOSED:
		return h.handleIssueClosed(msg.Context(), &event)
	default:

		log.Debug("Skipping non-task event", slog.String("event_name", event.EventName.String()))
		return nil
	}
}

func (h Handler) handleIssueOpened(ctx context.Context, event *eventpb.Event) error {
	log := logger.L()

	payload := event.GetIssueOpenedPayload()
	if payload == nil {
		log.Error("IssueOpenedPayload is nil", slog.String("event_id", event.Id))
		return nil
	}

	taskParam := task.CreateTaskParam{
		GithubID:       int64(payload.IssueId),
		IssueNumber:    int(payload.IssueNumber),
		Title:          payload.Title,
		State:          "open",
		RepositoryName: event.RepositoryName,
		Labels:         payload.Labels,
		CreatedAt:      event.Time.AsTime(),
	}

	err := h.taskSvc.CreateTask(ctx, taskParam)
	if err != nil {
		log.Error(
			"Failed to create task",
			slog.String("event_id", event.Id),
			slog.Int("issue_number", int(payload.IssueNumber)),
			slog.String("error", err.Error()),
		)
		return err
	}

	log.Info(
		"Task created successfully",
		slog.String("event_id", event.Id),
		slog.Int("issue_number", int(payload.IssueNumber)),
		slog.String("title", payload.Title),
	)

	return nil
}

func (h Handler) handleIssueClosed(ctx context.Context, event *eventpb.Event) error {
	log := logger.L()

	payload := event.GetIssueClosedPayload()
	if payload == nil {
		log.Error("IssueClosedPayload is nil", slog.String("event_id", event.Id))
		return nil
	}

	updateParam := task.UpdateTaskParam{
		IssueNumber:    int(payload.IssueNumber),
		RepositoryName: event.RepositoryName,
		State:          "closed",
		ClosedAt:       event.Time.AsTime(),
	}

	err := h.taskSvc.UpdateTask(ctx, updateParam)
	if err != nil {
		log.Error(
			"Failed to update task",
			slog.String("event_id", event.Id),
			slog.Int("issue_number", int(payload.IssueNumber)),
			slog.String("error", err.Error()),
		)
		return err
	}

	log.Info(
		"Task updated successfully",
		slog.String("event_id", event.Id),
		slog.Int("issue_number", int(payload.IssueNumber)),
		slog.String("state", "closed"),
	)

	return nil
}
