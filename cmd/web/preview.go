package main

import (
	"bytes"
	"net/http"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
	"github.com/labstack/echo/v4"
)

func registerLeaderboardPreview(router *echo.Echo, enabled bool) {
	if !enabled {
		return
	}

	router.GET("/preview", func(c echo.Context) error {
		view := leaderboardView{
			ProjectID: "5001",
			Preview:   true,
			Rows: []leaderboardstat.PublicLeaderboardRow{
				{Rank: 1, UserID: 101, Score: 12},
				{Rank: 2, UserID: 202, Score: 8},
				{Rank: 3, UserID: 303, Score: 3},
			},
		}

		switch c.QueryParam("state") {
		case "empty":
			view.Rows = nil
		case "error":
			view.Error = true
		}

		var output bytes.Buffer
		if err := renderLeaderboard(&output, view); err != nil {
			return err
		}

		return c.HTML(http.StatusOK, output.String())
	})
}
