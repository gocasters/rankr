package command

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "leaderboardscoring_service",
	Short: "A CLI for leaderboardscoring service",
	Long: `leaderboardscoring Service CLI is a tool to manage and run 
the leaderboardscoring service, including migrations and server startup.`,
}
