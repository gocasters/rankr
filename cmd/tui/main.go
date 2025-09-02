package main

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	leaderboardscoringCommand "github.com/gocasters/rankr/cmd/leaderboardscoring/command"
	taskCommand "github.com/gocasters/rankr/cmd/task/command"
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
	rootCmd.AddCommand(taskCommand.RootCmd)
}
