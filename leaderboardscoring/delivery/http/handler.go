package http

import (
	"github.com/labstack/echo/v4"
	"log/slog"
)

type Handler struct {
	Logger *slog.Logger
}

func NewHandler(logger *slog.Logger) Handler {
	return Handler{Logger: logger}
}

func (h Handler) GetLeaderboard(c echo.Context) error {
	return nil
}
