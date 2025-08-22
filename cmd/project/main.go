package main

import (
	"os"

	"github.com/gocasters/rankr/cmd/command"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
