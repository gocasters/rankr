package main

import (
    "github.com/gocasters/rankr/cmd/command"
    "github.com/gocasters/rankr/cmd/servercmd"
    "os"
)

func main() {
    // Add server command to root command
    command.RootCmd.AddCommand(servercmd.ServerCmd)
    
    if err := command.RootCmd.Execute(); err != nil {
        os.Exit(1)
    }
}