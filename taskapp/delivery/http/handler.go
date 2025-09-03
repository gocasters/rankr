package http

import (
	"github.com/gocasters/rankr/taskapp/service/task"
	"log/slog"
)

type Handler struct {
	TaskService task.Service
	Logger      *slog.Logger
}

func NewHandler(taskSrv task.Service, logger *slog.Logger) Handler {
	return Handler{
		TaskService: taskSrv,
		Logger:      logger,
	}
}
