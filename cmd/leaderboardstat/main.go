package main

import (
	"github.com/gocasters/rankr/cmd/leaderboardstat/command"
	"os"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
