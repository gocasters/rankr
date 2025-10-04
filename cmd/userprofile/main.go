package main

import (
	"fmt"
	"github.com/gocasters/rankr/cmd/userprofile/command"
	"os"
)

func main() {
	if err := command.RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
