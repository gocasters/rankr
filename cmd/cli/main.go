package main

import (
	contributorCommand "github.com/gocasters/rankr/cmd/contributor/command"
	leaderboardscoringCommand "github.com/gocasters/rankr/cmd/leaderboardscoring/command"
	leaderboardstatCommand "github.com/gocasters/rankr/cmd/leaderboardstat/command"
	taskCommand "github.com/gocasters/rankr/cmd/task/command"
	webhookCommand "github.com/gocasters/rankr/cmd/webhook/command"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "cli",
	Short: "Command line interface for all services",
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(leaderboardscoringCommand.RootCmd)
	rootCmd.AddCommand(leaderboardstatCommand.RootCmd)
	rootCmd.AddCommand(taskCommand.RootCmd)
	rootCmd.AddCommand(webhookCommand.RootCmd)
	rootCmd.AddCommand(contributorCommand.RootCmd)
}
