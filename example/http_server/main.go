package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/contrib/instrumentation/github.com/labstack/echo/otelecho"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.25.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"net/http"
	"time"
)

func main() {
	//newHttpServerWithoutOtel()
	newHttpServerWithOtel()
}

// newHttpServerWithoutOtel For test this example go to http://127.0.0.1:8080/ping
func newHttpServerWithoutOtel() {
	serverConfig := httpserver.Config{
		Port:            8080,
		CORS:            httpserver.CORS{AllowOrigins: []string{"*"}},
		ShutdownTimeout: 10 * time.Second,
		OtelMiddleware:  nil, // No middleware is injected.
	}

	server, nErr := httpserver.New(serverConfig)
	if nErr != nil {
		log.Fatalf("failed to initial server: %v", nErr)
	}

	server.GetRouter().GET(
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

// =============================================================================
// TODO: Refactor after 'otel' package merge
//
// ATTENTION: The 'initTracerProvider' function below is a temporary,
// localized implementation for setting up OpenTelemetry tracing.
// This was created because the centralized 'otel' package, which handles
// all telemetry configurations, is not yet merged into the main branch.
//
// Once the 'otel' package is available, please perform the following steps:
//  1. Delete this entire 'initTracerProvider' function.
//  2. Import the 'otel' package.
//  3. In the 'newHttpServerWithOtel' function, replace the call to
//     'initTracerProvider' with a call to 'otel.NewOtelAdapter(config)'.
//     The adapter will handle setting the global tracer provider internally.
//
// This change is crucial to centralize telemetry logic, avoid code
// duplication, and adhere to our project's architectural design.
// =============================================================================
func initTracerProvider() (*sdktrace.TracerProvider, error) {
	ctx := context.Background()

	conn, err := grpc.NewClient("127.0.0.1:4317", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to create gRPC connection to collector: %w", err)
	}

	traceExporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, fmt.Errorf("failed to create trace exporter: %w", err)
	}

	resource := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String("httpserver"),
	)

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(resource),
	)

	return tp, nil
}

// newHttpServerWithOtel For test this example go to http://127.0.0.1:8080/ping-otel
func newHttpServerWithOtel() {

	traceProvider, err := initTracerProvider()
	if err != nil {
		log.Fatal(err)
	}

	otel.SetTracerProvider(traceProvider)

	otelMiddleware := otelecho.Middleware("httpserver")
	serverConfig := httpserver.Config{
		Port:            8080,
		CORS:            httpserver.CORS{AllowOrigins: []string{"*"}},
		ShutdownTimeout: 10 * time.Second,
		OtelMiddleware:  otelMiddleware,
	}

	server, nErr := httpserver.New(serverConfig)
	if nErr != nil {
		log.Fatalf("failed to initial server: %v", nErr)
	}

	server.GetRouter().GET(
		"/ping-otel",
		func(c echo.Context) error {
			return c.String(http.StatusOK, "pong with otel")
		},
	)

	if err := server.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("failed to start server: %v", err)
	}
}
