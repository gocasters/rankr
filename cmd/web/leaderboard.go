package main

import (
	"io"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
)

type leaderboardView struct {
	ProjectID string
	Rows      []leaderboardstat.PublicLeaderboardRow
	Error     bool
}

func renderLeaderboard(w io.Writer, view leaderboardView) error {
	return nil
}
