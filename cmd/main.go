package main

import (
	"github.com/gocasters/rankr/cmd/command"
	"os"
)

// main executes the application's root CLI command. It calls command.RootCmd.Execute()
// and terminates the process with exit status 1 if that call returns a non-nil error; otherwise
// it returns normally.
func main() {
	if err := command.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
