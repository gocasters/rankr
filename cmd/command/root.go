package command

import (
    "github.com/spf13/cobra"
    "github.com/gocasters/rankr/cmd/contributorcmd" // Import the contributor commands
)

var RootCmd = &cobra.Command{
    Use: "rankr",
    Short: "A CLI for rankr app",
}

func init() {
    // Add contributor commands
    RootCmd.AddCommand(contributorcmd.ContributorCmd)
}