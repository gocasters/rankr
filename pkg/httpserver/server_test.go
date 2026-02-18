package httpserver_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/authhttp"
	"github.com/gocasters/rankr/pkg/httpserver"

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
	userInfo, encErr := authhttp.EncodeUserInfo("1")
	assert.NoError(t, encErr)
	req.Header.Set("X-User-Info", userInfo)
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
	userInfo, encErr := authhttp.EncodeUserInfo("1")
	assert.NoError(t, encErr)
	req.Header.Set("X-User-Info", userInfo)
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
	stopErr := server.Stop(t.Context())
	assert.NoError(t, stopErr, "StopWithTimeout should not return an error when stopping a running server")

	// Now verify Start() exited due to shutdown
	startErr := <-errCh
	assert.ErrorIs(t, startErr, http.ErrServerClosed, "server.Start() should return ErrServerClosed after shutdown")
}

func TestPublicRoutesAndProtectedRoutes(t *testing.T) {
	server, err := httpserver.New(httpserver.Config{Port: 8080})
	assert.NoError(t, err)

	server.GetRouter().GET("/v1/module/health-check", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	server.GetRouter().POST("/v1/login", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})
	server.GetRouter().GET("/secure", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	healthReq := httptest.NewRequest(http.MethodGet, "/v1/module/health-check", nil)
	healthRec := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(healthRec, healthReq)
	assert.Equal(t, http.StatusOK, healthRec.Code)

	loginReq := httptest.NewRequest(http.MethodPost, "/v1/login", nil)
	loginRec := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(loginRec, loginReq)
	assert.Equal(t, http.StatusOK, loginRec.Code)

	secureReq := httptest.NewRequest(http.MethodGet, "/secure", nil)
	secureRec := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(secureRec, secureReq)
	assert.Equal(t, http.StatusUnauthorized, secureRec.Code)
}

func TestConfigPublicPathsAreSkipped(t *testing.T) {
	server, err := httpserver.New(httpserver.Config{
		Port:        8080,
		PublicPaths: []string{" v1/public/info/ "},
	})
	assert.NoError(t, err)

	server.GetRouter().GET("/v1/public/info", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/public/info", nil)
	rec := httptest.NewRecorder()
	server.GetRouter().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
}
