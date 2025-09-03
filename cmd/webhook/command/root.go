package command

import (
	"github.com/spf13/cobra"
)

var RootCmd = &cobra.Command{
	Use:   "webhook_service",
	Short: "A CLI for webhook service",
	Long: `webhook Service CLI is a tool to manage and run 
the webhook service, including migrations and server startup.`,
}
