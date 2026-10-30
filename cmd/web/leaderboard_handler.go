package main

import (
	"github.com/gocasters/rankr/adapter/leaderboardstat"
	"github.com/labstack/echo/v4"
)

func registerLeaderboard(
	router *echo.Echo,
	client *leaderboardstat.HTTPClient,
	enabled bool,
) {
	// Implementation follows the failing tests.
}
