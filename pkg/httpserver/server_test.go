package httpserver_test

import (
	"github.com/gocasters/rankr/pkg/httpserver"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

// TestNew tests the constructor function for various scenarios.
func TestNew(t *testing.T) {
	t.Run("successful creation with valid config", func(t *testing.T) {
		// Arrange
		cfg := httpserver.Config{
			Port:            8080,
			ShutdownTimeout: 5 * time.Second,
		}

		// Act
		server, err := httpserver.New(cfg)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.NotNil(t, server.GetRouter(), "GetRouter should return a non-nil router")
	})

	t.Run("error on invalid port", func(t *testing.T) {
		// Arrange
		cfg := httpserver.Config{Port: 0} // Invalid port

		// Act
		server, err := httpserver.New(cfg)

		// Assert
		assert.Error(t, err)
		assert.Nil(t, server)
		assert.Contains(t, err.Error(), "invalid port")
	})

	t.Run("sets default shutdown timeout", func(t *testing.T) {
		// Arrange
		cfg := httpserver.Config{
			Port:            8080,
			ShutdownTimeout: 0, // No timeout provided
		}

		// Act
		server, err := httpserver.New(cfg)

		// Assert
		assert.NoError(t, err)
		assert.NotNil(t, server)
		assert.Equal(t, httpserver.DefaultShutdownTimeout, server.GetConfig().ShutdownTimeout, "Default timeout should be set")
	})
}

// TestOtelMiddlewareInjection verifies that optional middleware is correctly added.
func TestOtelMiddlewareInjection(t *testing.T) {
	// Arrange
	middlewareWasCalled := false
	mockOtelMiddleware := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			middlewareWasCalled = true
			return next(c)
		}
	}

	cfg := httpserver.Config{
		Port:           8080,
		OtelMiddleware: mockOtelMiddleware,
	}

	server, err := httpserver.New(cfg)
	assert.NoError(t, err)

	server.GetRouter().GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	rec := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rec, req)

	// Assert
	assert.True(t, middlewareWasCalled, "The injected Otel middleware should have been called")
}

// TestRouteRegistrationAndResponse confirms that routes can be added and respond correctly.
func TestRouteRegistrationAndResponse(t *testing.T) {
	// Arrange
	server, err := httpserver.New(httpserver.Config{Port: 8080})
	assert.NoError(t, err)

	expectedResponse := "Hello, Tester!"

	// Act: Register a new GET route using the GetRouter() method.
	server.GetRouter().GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, expectedResponse)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expectedResponse, rec.Body.String())
}

func TestStopWithTimeout(t *testing.T) {
	// Arrange
	server, err := httpserver.New(httpserver.Config{Port: 9090}) // Use a different port to avoid conflicts
	assert.NoError(t, err)

	// Act: Start the server in a separate goroutine
	errCh := make(chan error, 1)
	go func() {
		errCh <- server.Start()
	}()

	// Stop the server and verify shutdown succeeds
	stopErr := server.StopWithTimeout()
	assert.NoError(t, stopErr, "StopWithTimeout should not return an error when stopping a running server")

	// Now verify Start() exited due to shutdown
	startErr := <-errCh
	assert.ErrorIs(t, startErr, http.ErrServerClosed, "server.Start() should return ErrServerClosed after shutdown")
}
