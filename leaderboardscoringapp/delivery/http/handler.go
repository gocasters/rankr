package http

import (
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/labstack/echo/v4"
	nethttp "net/http"
)

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func (h Handler) GetLeaderboard(c echo.Context) error {
	logger.L().Info("GetLeaderboard called")
	return c.NoContent(nethttp.StatusNotImplemented) // TODO: implement me
}
