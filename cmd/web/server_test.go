package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gocasters/rankr/pkg/httpserver"
)

func TestHealthCheck(t *testing.T) {
	server, err := newServer(httpserver.Config{
		Port: 8085,
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	request := httptest.NewRequest(
		http.MethodGet, "/v1/health-check", nil,
	)
	response := httptest.NewRecorder()

	server.GetRouter().ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("want status 200, got %d", response.Code)
	}
}
