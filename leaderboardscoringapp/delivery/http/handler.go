package http

import (
	"errors"
	"net/http"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	leaderboardScoringSvc *leaderboardscoring.Service
}

func NewHandler(leaderboardScoringSvc *leaderboardscoring.Service) Handler {
	return Handler{
		leaderboardScoringSvc: leaderboardScoringSvc,
	}
}

type GetLeaderboardQueryParams struct {
	ProjectID string `query:"project_id"`
	Timeframe string `query:"timeframe"`
	PageSize  int32  `query:"page_size"`
	Offset    int32  `query:"offset"`
}

type LeaderboardRowResponse struct {
	Rank   int64  `json:"rank"`
	UserID string `json:"user_id"`
	Score  int64  `json:"score"`
}

type GetLeaderboardHTTPResponse struct {
	Timeframe string                   `json:"timeframe"`
	ProjectID string                   `json:"project_id,omitempty"`
	Rows      []LeaderboardRowResponse `json:"rows"`
}

func (h Handler) GetLeaderboard(c echo.Context) error {
	var params GetLeaderboardQueryParams
	if err := c.Bind(&params); err != nil {
		logger.L().Error("failed to bind query params", "error", err.Error())
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid query parameters"})
	}

	if params.Timeframe == "" {
		params.Timeframe = leaderboardscoring.AllTime.String()
	}

	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	if params.PageSize > 100 {
		params.PageSize = 100
	}

	var projectIDPtr *string
	if params.ProjectID != "" {
		projectIDPtr = &params.ProjectID
	}

	req := &leaderboardscoring.GetLeaderboardRequest{
		Timeframe: params.Timeframe,
		ProjectID: projectIDPtr,
		PageSize:  params.PageSize,
		Offset:    params.Offset,
	}

	res, err := h.leaderboardScoringSvc.GetLeaderboard(c.Request().Context(), req)
	if err != nil {
		logger.L().Error("failed to get leaderboard", "error", err.Error())
		if errors.Is(err, leaderboardscoring.ErrInvalidArguments) {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request parameters"})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "internal server error"})
	}

	rows := make([]LeaderboardRowResponse, 0, len(res.LeaderboardRows))
	for _, r := range res.LeaderboardRows {
		rows = append(rows, LeaderboardRowResponse{
			Rank:   r.Rank,
			UserID: r.UserID,
			Score:  r.Score,
		})
	}

	projectID := ""
	if res.ProjectID != nil {
		projectID = *res.ProjectID
	}

	response := GetLeaderboardHTTPResponse{
		Timeframe: res.Timeframe,
		ProjectID: projectID,
		Rows:      rows,
	}

	return c.JSON(http.StatusOK, response)
}
