package http

import (
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"strconv"
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

func (h Handler) GetLeaderboards(c echo.Context) error {
	var req leaderboardstat.LeaderboardFilterRequest

	if err := c.Bind(&req); err != nil {
		return err
	}

	response, err := h.LeaderboardStatService.GetLeaderboardByFilters(c.Request().Context(), req)

	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, response)
}

func (h Handler) GetContributorStats(c echo.Context) error {
	var contributorID int
	var err error

	if contributorID, err = strconv.Atoi(c.Param("id")); err != nil {
		// TODO - use error pattern
		return err
	}

	response, err := h.LeaderboardStatService.GetContributorStats(c.Request().Context(), contributorID)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, response)
}
