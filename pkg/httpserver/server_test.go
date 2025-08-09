package httpserver_test

import (
	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestNewServer checks if the default middlewares are applied by testing their effects.
func TestNewServer(t *testing.T) {
	// Arrange
	cfg := httpserver.Config{
		Port: 8080,
		Cors: httpserver.Cors{
			AllowOrigins: []string{"https://example.com"},
		},
		OtelMiddleware: nil,
	}

	server := httpserver.New(cfg)
	server.Router.GET("/test", func(c echo.Context) error {
		return c.String(http.StatusOK, "ok")
	})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Origin", "https://example.com") // Set Origin to trigger CORS
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	// Assert
	assert.NotNil(t, server)
	assert.NotNil(t, server.Router)

	// Check if the CORS middleware added the correct header.
	assert.Equal(t, "https://example.com", rec.Header().Get(echo.HeaderAccessControlAllowOrigin))
	// Check for the RequestID header.
	assert.NotEmpty(t, rec.Header().Get(echo.HeaderXRequestID))
}

// TestRouteRegistrationAndResponse confirms that routes can be added and respond correctly.
// This test remains unchanged.
func TestRouteRegistrationAndResponse(t *testing.T) {
	// Arrange
	server := httpserver.New(httpserver.Config{Port: 8080})
	expectedResponse := "Hello, Tester!"

	// Act
	server.Router.GET("/hello", func(c echo.Context) error {
		return c.String(http.StatusOK, expectedResponse)
	})

	req := httptest.NewRequest(http.MethodGet, "/hello", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, expectedResponse, rec.Body.String())
}

// TestNotFoundRoute checks the server's behavior for a route that does not exist.
// This test remains unchanged.
func TestNotFoundRoute(t *testing.T) {
	// Arrange
	server := httpserver.New(httpserver.Config{Port: 8080})

	// Act
	req := httptest.NewRequest(http.MethodGet, "/this-route-does-not-exist", nil)
	rec := httptest.NewRecorder()
	server.Router.ServeHTTP(rec, req)

	// Assert
	assert.Equal(t, http.StatusNotFound, rec.Code)
}
