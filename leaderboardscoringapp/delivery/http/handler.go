package http

import (
	"github.com/labstack/echo/v4"
	"log/slog"
	nethttp "net/http"
)

type Handler struct {
	Logger *slog.Logger
}

func NewHandler(logger *slog.Logger) Handler {
	return Handler{Logger: logger}
}

func (h Handler) GetLeaderboard(c echo.Context) error {
	h.Logger.Info("GetLeaderboard called")
	return c.NoContent(nethttp.StatusNotImplemented) // TODO: implement me
}
