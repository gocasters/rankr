package main

import (
	"bytes"
	"net/http"
	"strconv"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
	"github.com/labstack/echo/v4"
)

func registerLeaderboard(
	router *echo.Echo,
	client *leaderboardstat.HTTPClient,
	enabled bool,
) {
	if !enabled {
		return
	}

	router.GET("/leaderboard", func(c echo.Context) error {
		repositoryID, err := strconv.ParseUint(
			c.QueryParam("project_id"), 10, 64,
		)
		if err != nil || repositoryID == 0 {
			return echo.NewHTTPError(
				http.StatusBadRequest,
				"شناسهٔ مخزن نامعتبر است.",
			)
		}

		view := leaderboardView{
			ProjectID: strconv.FormatUint(repositoryID, 10),
		}

		result, err := client.GetPublicLeaderboard(
			c.Request().Context(), repositoryID, 10, 0,
		)

		status := http.StatusOK
		if err != nil {
			c.Logger().Error(err)
			view.Error = true
			status = http.StatusBadGateway
		} else {
			view.Rows = result.Rows
		}

		var output bytes.Buffer
		if err := renderLeaderboard(&output, view); err != nil {
			return err
		}

		return c.HTML(status, output.String())
	})
}
