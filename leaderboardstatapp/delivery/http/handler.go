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
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid contributor ID",
		})
	}

	contributorID := types.ID(idInt)
	response, err := h.LeaderboardStatService.GetContributorStats(c.Request().Context(), contributorID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get contributor stats",
		})
	}

	return c.JSON(http.StatusOK, response)
}

type PublicLeaderboardRowResponse struct {
	Rank   uint64  `json:"rank"`
	UserID uint64  `json:"user_id"`
	Score  float64 `json:"score"`
}

type GetPublicLeaderboardResponse struct {
	ProjectID   string                         `json:"project_id"`
	Rows        []PublicLeaderboardRowResponse `json:"rows"`
	LastUpdated *string                        `json:"last_updated,omitempty"`
}

func (h Handler) GetPublicLeaderboard(c echo.Context) error {
	projectID := c.Param("project_id")
	if projectID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "project_id is required",
		})
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid project_id",
		})
	}

	pageSizeStr := c.QueryParam("page_size")
	pageSize := int32(10)
	if pageSizeStr != "" {
		ps, err := strconv.ParseInt(pageSizeStr, 10, 32)
		if err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := int32(0)
	if offsetStr != "" {
		off, err := strconv.ParseInt(offsetStr, 10, 32)
		if err == nil && off >= 0 {
			offset = int32(off)
		}
	}

	result, err := h.LeaderboardStatService.GetPublicLeaderboard(c.Request().Context(), types.ID(projectIDInt), pageSize, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get public leaderboard",
		})
	}

	rows := make([]PublicLeaderboardRowResponse, 0, len(result.UsersScore))
	for _, us := range result.UsersScore {
		rows = append(rows, PublicLeaderboardRowResponse{
			Rank:   us.Rank,
			UserID: uint64(us.ContributorID),
			Score:  us.Score,
		})
	}

	var lastUpdatedStr *string
	if result.LastUpdated != nil {
		formatted := result.LastUpdated.Format("2006-01-02T15:04:05Z07:00")
		lastUpdatedStr = &formatted
	}

	response := GetPublicLeaderboardResponse{
		ProjectID:   projectID,
		Rows:        rows,
		LastUpdated: lastUpdatedStr,
	}

	return c.JSON(http.StatusOK, response)
}

type PublicLeaderboardRowResponse struct {
	Rank   uint64  `json:"rank"`
	UserID uint64  `json:"user_id"`
	Score  float64 `json:"score"`
}

type GetPublicLeaderboardResponse struct {
	ProjectID string                         `json:"project_id"`
	Rows      []PublicLeaderboardRowResponse `json:"rows"`
}

func (h Handler) GetPublicLeaderboard(c echo.Context) error {
	projectID := c.Param("project_id")
	if projectID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "project_id is required",
		})
	}

	projectIDInt, err := strconv.ParseUint(projectID, 10, 64)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid project_id",
		})
	}

	pageSizeStr := c.QueryParam("page_size")
	pageSize := int32(10)
	if pageSizeStr != "" {
		ps, err := strconv.ParseInt(pageSizeStr, 10, 32)
		if err == nil && ps > 0 && ps <= 100 {
			pageSize = int32(ps)
		}
	}

	offsetStr := c.QueryParam("offset")
	offset := int32(0)
	if offsetStr != "" {
		off, err := strconv.ParseInt(offsetStr, 10, 32)
		if err == nil && off >= 0 {
			offset = int32(off)
		}
	}

	result, err := h.LeaderboardStatService.GetPublicLeaderboard(c.Request().Context(), types.ID(projectIDInt), pageSize, offset)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to get public leaderboard",
		})
	}

	rows := make([]PublicLeaderboardRowResponse, 0, len(result.UsersScore))
	for _, us := range result.UsersScore {
		rows = append(rows, PublicLeaderboardRowResponse{
			Rank:   us.Rank,
			UserID: uint64(us.ContributorID),
			Score:  us.Score,
		})
	}

	response := GetPublicLeaderboardResponse{
		ProjectID: projectID,
		Rows:      rows,
	}

	return c.JSON(http.StatusOK, response)
}
