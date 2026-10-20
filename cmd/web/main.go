package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/spf13/cobra"
)

func main() {

	ctx, stop := signal.NotifyContext(
		context.Background(),
		os.Interrupt,
		syscall.SIGTERM,
	)

	defer stop()

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

			return serve(cmd.Context(), server)
		},
	})

	if err := root.ExecuteContext(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func serve(ctx context.Context, server *httpserver.Server) error {

	done := make(chan error, 1)

	go func() {
		done <- server.Start()
	}()

	var err error

	select {
	case err = <-done:
		// The server stopped or failed to start

	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(
			context.Background(),
			server.GetConfig().ShutdownTimeout,
		)
		defer cancel()

		if err := server.Stop(shutdownCtx); err != nil {
			return fmt.Errorf("web: shutdown: %w", err)
		}

		err = <-done
	}

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("web: serve: %w", err)
	}
	return nil
}
