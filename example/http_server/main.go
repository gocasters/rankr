package main

import (
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"time"
)

func main() {
	newHttpServerWithoutOtel()
	//newHttpServerWithOtel()
}

// newHttpServerWithoutOtel For test this example go to http://127.0.0.1:8080/ping
func newHttpServerWithoutOtel() {
	serverConfig := httpserver.Config{
		Port:            8080,
		Cors:            httpserver.Cors{AllowOrigins: []string{"*"}},
		ShutDownTimeout: 10 * time.Second,
		OtelMiddleware:  nil, // No middleware is injected.
	}

	server := httpserver.New(serverConfig)

	server.Router.GET(
		"/ping",
		func(c echo.Context) error {
			return c.String(http.StatusOK, "pong")
		},
	)

	fmt.Println("Server without Otel is ready.")
	if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}

// Instructions for the Future Developer:
// 1. Delete this entire mocks.
// 2. Import the real `otel` package.
// 3. Replace the mock functions with the ones from the package.

type mockOtelAdapter struct{}

func (m *mockOtelAdapter) IsConfigured() bool { return true }

func NewOtelAdapter() (*mockOtelAdapter, error) {
	return &mockOtelAdapter{}, nil
}

func EchoRequestLoggerBeforeNextFunc(pkg, fn string) func(c echo.Context) {
	return func(c echo.Context) {
		fmt.Printf("   [Otel Mock] -> BeforeNextFunc called for %s.\n", pkg)
	}
}

func EchoRequestLoggerLogValuesFunc(pkg, fn string) func(c echo.Context, v middleware.RequestLoggerValues) error {
	return func(c echo.Context, v middleware.RequestLoggerValues) error {
		fmt.Printf("   [Otel Mock] -> LogValuesFunc called for %s with status %d.\n", pkg, v.Status)
		return nil
	}
}

// newHttpServerWithOtel For test this example go to http://127.0.0.1:8080/ping-otel
func newHttpServerWithOtel() {
	otelAdapter, err := NewOtelAdapter()
	if err != nil {
		log.Panic(err)
	}

	var otelMiddleware echo.MiddlewareFunc

	if otelAdapter.IsConfigured() {
		otelMiddleware = middleware.RequestLoggerWithConfig(
			middleware.RequestLoggerConfig{
				LogURI:    true,
				LogStatus: true,
				// any config...
				BeforeNextFunc: EchoRequestLoggerBeforeNextFunc("httpserver", "Serve"),
				LogValuesFunc:  EchoRequestLoggerLogValuesFunc("httpserver", "Serve"),
			},
		)
	}

	serverConfig := httpserver.Config{
		Port:            8080,
		Cors:            httpserver.Cors{AllowOrigins: []string{"*"}},
		ShutDownTimeout: 10 * time.Second,
		OtelMiddleware:  otelMiddleware,
	}

	server := httpserver.New(serverConfig)

	server.Router.GET(
		"/ping-otel",
		func(c echo.Context) error {
			return c.String(http.StatusOK, "pong with otel")
		},
	)

	fmt.Println("Server with Otel (mock) is ready.")
	if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}
