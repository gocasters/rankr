package command

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "contributor_service",
	Short: "A CLI for contributor Service",
	Long: `contributor Service CLI is a tool to manage and run 
the contributor service, including migrations and server startup.`,
}
