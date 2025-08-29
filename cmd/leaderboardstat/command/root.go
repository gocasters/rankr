package command

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "leaderboardstat_service",
	Short: "A CLI for leaderboardstat service",
	Long: `leaderboardstat Service CLI is a tool to manage and run 
the leaderboardstat service, including migrations and server startup.`,
}
