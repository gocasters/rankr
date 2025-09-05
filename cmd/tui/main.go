package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	contributorCommand "github.com/gocasters/rankr/cmd/contributor/command"
	leaderboardscoringCommand "github.com/gocasters/rankr/cmd/leaderboardscoring/command"
	leaderboardstatCommand "github.com/gocasters/rankr/cmd/leaderboardstat/command"
	taskCommand "github.com/gocasters/rankr/cmd/task/command"
	webhookCommand "github.com/gocasters/rankr/cmd/webhook/command"
	"github.com/gocasters/rankr/pkg/ui"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tui",
	Short: "Interactive CI for all services",
}

func main() {
	m := ui.NewModel(rootCmd)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Println(err)
		return
	}
}

func init() {
	rootCmd.AddCommand(leaderboardscoringCommand.RootCmd)
	rootCmd.AddCommand(leaderboardstatCommand.RootCmd)
	rootCmd.AddCommand(taskCommand.RootCmd)
	rootCmd.AddCommand(webhookCommand.RootCmd)
	rootCmd.AddCommand(contributorCommand.RootCmd)
}
