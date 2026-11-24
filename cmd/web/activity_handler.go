package main

import (
	"bytes"
	"embed"
	"fmt"
	"html/template"
	"io"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

const (
	activityFeedLimit  = 20
	activityTimeLayout = "2006-01-02 15:04"
)

//go:embed templates/activity.html
var activityTemplates embed.FS

var activityTypeLabels = map[string]string{
	"pull_request_opened": "باز کردن پول‌ریکوئست",
	"pull_request_closed": "بستن پول‌ریکوئست",
	"pull_request_review": "بازبینی پول‌ریکوئست",
	"issue_opened":        "ایجاد ایشو",
	"issue_closed":        "بستن ایشو",
	"issue_comment":       "کامنت ایشو",
	"commit_push":         "پوش",
}

type activityRow struct {
	Type       string
	Subject    string
	Repository string
	Points     int64
	When       string
}

type activityView struct {
	UserID string
	Rows   []activityRow
	Error  bool
}

// registerActivity serves a contributor's recent scored events from the
// leaderboardscoring activity API. It is off unless the
// profile.activity_feed flag is enabled.
func registerActivity(router *echo.Echo, client *activityClient, enabled bool) {
	if !enabled {
		return
	}

	router.GET("/profile", func(c echo.Context) error {
		userID, err := strconv.ParseUint(c.QueryParam("user_id"), 10, 64)
		if err != nil || userID == 0 {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"شناسهٔ کاربر نامعتبر است.",
			)
		}

		view := activityView{
			UserID: strconv.FormatUint(userID, 10),
		}

		activity, err := client.GetContributorActivity(
			c.Request().Context(), userID, activityFeedLimit,
		)

		status := http.StatusOK
		if err != nil {
			c.Logger().Error(err)
			view.Error = true
			status = http.StatusBadGateway
		} else {
			view.Rows = activityRows(activity)
		}

		var output bytes.Buffer
		if err := renderActivity(&output, view); err != nil {
			return err
		}

		return c.HTML(status, output.String())
	})
}

func activityRows(activity []contributorActivity) []activityRow {
	rows := make([]activityRow, 0, len(activity))
	for _, item := range activity {
		label, ok := activityTypeLabels[item.Type]
		if !ok {
			label = item.Type
		}

		rows = append(rows, activityRow{
			Type:       label,
			Subject:    activitySubject(item),
			Repository: orDash(item.Repository),
			Points:     item.Points,
			When:       item.OccurredAt.Local().Format(activityTimeLayout),
		})
	}

	return rows
}

func activitySubject(item contributorActivity) string {
	if item.SubjectNumber <= 0 {
		return "—"
	}

	switch item.SubjectType {
	case "pull_request":
		return "PR #" + strconv.FormatInt(int64(item.SubjectNumber), 10)
	case "issue":
		return "Issue #" + strconv.FormatInt(int64(item.SubjectNumber), 10)
	default:
		return strconv.FormatInt(int64(item.SubjectNumber), 10)
	}
}

func orDash(value string) string {
	if value == "" {
		return "—"
	}

	return value
}

func renderActivity(w io.Writer, view activityView) error {
	page, err := template.ParseFS(
		activityTemplates,
		"templates/activity.html",
	)
	if err != nil {
		return fmt.Errorf("web: parse activity template: %w", err)
	}

	if err := page.Execute(w, view); err != nil {
		return fmt.Errorf("web: render activity: %w", err)
	}

	return nil
}
