package command

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "project_service",
	Short: "A CLI for project Service",
	Long: `project Service CLI is a tool to manage and run 
the project service, including migrations and server startup.`,
}
