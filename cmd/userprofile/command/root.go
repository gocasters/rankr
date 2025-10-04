package command

import "github.com/spf13/cobra"

var RootCmd = &cobra.Command{
	Use:   "userprofile_service",
	Short: "A CLI for userprofile service",
	Long:  `userprofile service cli is a tool to manage and run the userprofile service`,
}
