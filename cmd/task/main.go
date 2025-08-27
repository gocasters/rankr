package main

import (
	"os"

	"github.com/gocasters/rankr/cmd/task/command"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
