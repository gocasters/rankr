package main

import (
	"fmt"
	"github.com/gocasters/rankr/cmd/leaderboardscoring/command"
	"os"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
