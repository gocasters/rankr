package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gocasters/rankr/pkg/path"
	"github.com/spf13/cobra"
)

func newRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use: "web",
	}

	preview := root.PersistentFlags().Bool(
		"preview", false, "Enable a local preview with sample data",
	)

	root.AddCommand(&cobra.Command{
		Use:   "serve",
		Short: "Start the web service",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			configPath := os.Getenv("CONFIG_PATH")

			if configPath == "" {
				projectRoot, err := path.PathProjectRoot()
				if err != nil {
					return fmt.Errorf("web: find project root: %w", err)
				}

				configPath = filepath.Join(
					projectRoot, "deploy", "web", "development", "config.yml",
				)
			}

			cfg, err := loadConfig(configPath)
			if err != nil {
				return err
			}

			if *preview {
				cfg.HTTPServer.PublicPaths = append(
					cfg.HTTPServer.PublicPaths, "/preview",
				)
			}

			server, err := newServer(cfg.HTTPServer)
			if err != nil {
				return err
			}

			registerLeaderboardPreview(server.GetRouter(), *preview)

			return serve(cmd.Context(), server)
		},
	})

	return root
}
