package main

import (
	"embed"
	"fmt"
	"html/template"
	"io"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
)

//go:embed templates/leaderboard.html
var leaderboardTemplates embed.FS

type leaderboardView struct {
	ProjectID string
	Rows      []leaderboardstat.PublicLeaderboardRow
	Error     bool
	Preview   bool
}

func renderLeaderboard(w io.Writer, view leaderboardView) error {
	page, err := template.ParseFS(
		leaderboardTemplates,
		"templates/leaderboard.html",
	)

	if err != nil {
		return fmt.Errorf("web: parse leaderboard template: %w", err)
	}

	if err := page.Execute(w, view); err != nil {
		return fmt.Errorf("web: render leaderboard: %w", err)
	}

	return nil
}
