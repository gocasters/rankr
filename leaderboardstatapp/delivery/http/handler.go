package http

import (
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

type Handler struct {
	LeaderboardStatService leaderboardstat.Service
	Logger                 *slog.Logger
}

func NewHandler(leaderboardStatService leaderboardstat.Service, logger *slog.Logger) *Handler {
	return &Handler{
		LeaderboardStatService: leaderboardStatService,
		Logger:                 logger,
	}
}

func (h Handler) GetScoreboard(c echo.Context) error {
	var req leaderboardstat.ScoreboardFilterRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	response, err := h.LeaderboardStatService.GetScoreboardByFilters(c.Request().Context(), req)

	if err != nil {
		return err
		// TODO return handleServiceError
	}
	return c.JSON(http.StatusOK, response)
}

func (h Handler) GetContributorScores(c echo.Context) error {
	return nil
}
