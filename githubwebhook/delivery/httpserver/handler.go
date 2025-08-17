package httpserver

import (
	"log/slog"
)

type Handler struct {
	Logger *slog.Logger
}

func NewHandler(logger *slog.Logger) Handler {
	return Handler{Logger: logger}
}
