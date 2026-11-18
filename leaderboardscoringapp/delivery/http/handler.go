package http

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gocasters/rankr/leaderboardscoringapp/service/leaderboardscoring"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	service             *leaderboardscoring.Service
	activityFeedEnabled bool
}

func NewHandler(svc *leaderboardscoring.Service, activityFeedEnabled bool) Handler {
	return Handler{
		service:             svc,
		activityFeedEnabled: activityFeedEnabled,
	}
}

func (h Handler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

type contributorActivityResponse struct {
	UserID   string                                   `json:"user_id"`
	Limit    int                                      `json:"limit"`
	Activity []leaderboardscoring.ContributorActivity `json:"activity"`
}

// GetContributorActivity serves the latest scored events of one contributor.
// The route is off unless the profile.activity_feed flag is enabled.
func (h Handler) GetContributorActivity(c echo.Context) error {
	if !h.activityFeedEnabled {
		return c.JSON(http.StatusNotFound, echo.Map{
			"error": "activity feed is disabled",
		})
	}

	userID := strings.TrimSpace(c.Param("user_id"))
	if userID == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": "user_id is required",
		})
	}

	limit := leaderboardscoring.NormalizeActivityLimit(0)
	if rawLimit := strings.TrimSpace(c.QueryParam("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 {
			return c.JSON(http.StatusBadRequest, echo.Map{
				"error": "limit must be a positive integer",
			})
		}

		limit = leaderboardscoring.NormalizeActivityLimit(parsed)
	}

	activity, err := h.service.GetContributorActivity(c.Request().Context(), userID, limit)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": "failed to get contributor activity",
		})
	}

	if activity == nil {
		activity = make([]leaderboardscoring.ContributorActivity, 0)
	}

	return c.JSON(http.StatusOK, contributorActivityResponse{
		UserID:   userID,
		Limit:    limit,
		Activity: activity,
	})
}
