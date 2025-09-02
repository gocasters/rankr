package main

import (
	leaderboardscoringCommand "github.com/gocasters/rankr/cmd/leaderboardscoring/command"
	taskCommand "github.com/gocasters/rankr/cmd/task/command"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "ci",
	Short: "Interactive CI for all services",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(leaderboardscoringCommand.RootCmd)
	rootCmd.AddCommand(taskCommand.RootCmd)
}
