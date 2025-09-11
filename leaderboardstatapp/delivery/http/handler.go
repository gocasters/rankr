package http

import (
	"github.com/gocasters/rankr/leaderboardstatapp/service/leaderboardstat"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type Handler struct {
	LeaderboardStatService leaderboardstat.Service
}

func NewHandler(leaderboardStatService leaderboardstat.Service) Handler {
	return Handler{
		LeaderboardStatService: leaderboardStatService,
	}
}

func (h Handler) GetContributorStats(c echo.Context) error {
	idStr := c.Param("id")

	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		// TODO - use error pattern
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid contributor ID",
		})
	}

	contributorID := types.ID(idInt)

	response, err := h.LeaderboardStatService.GetContributorTotalStats(c.Request().Context(), contributorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get contributor stats",
		})
	}

	return c.JSON(http.StatusOK, response)
}
