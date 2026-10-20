package main

import (
	"fmt"
	"net/http"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/labstack/echo/v4"
)

func newServer(cfg httpserver.Config) (*httpserver.Server, error) {
	server, err := httpserver.New(cfg)
	if err != nil {
		return nil, fmt.Errorf("web: create server: %w", err)
	}

	router := server.GetRouter()

	router.GET("/v1/health-check", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{
			"status": "ok",
		})
	})

	return server, nil
}
