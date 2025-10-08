package command

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "realtime_service",
	Short: "A CLI for realtime Service",
	Long: `realtime Service CLI is a tool to manage and run
the realtime service for publishing real-time events to clients.`,
}
