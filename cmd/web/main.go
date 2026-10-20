package main

import (
	"fmt"
	"os"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/spf13/cobra"
)

func main() {
	root := &cobra.Command{
		Use: "web",
	}

	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the web service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			server, err := newServer(httpserver.Config{
				Host: "127.0.0.1",
				Port: 8085,
			})
			if err != nil {
				return err
			}

			return server.Start()
		},
	})

	if err := root.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
