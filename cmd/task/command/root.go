package command

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "task_service",
	Short: "A CLI for task Service",
	Long: `task Service CLI is a tool to manage and run 
the task service, including migrations and server startup.`,
}
