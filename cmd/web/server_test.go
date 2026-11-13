package main

import (
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

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
	var body map[string]string
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		t.Fatalf("decode health response: %v", err)
	}
	if body["status"] != "ok" {
		t.Fatalf("expected status field ok, got %q", body["status"])
	}

}

func TestServeStopsWhenContextCanceled(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("create listener: %v", err)
	}
	defer listener.Close()

	server, err := newServer(httpserver.Config{
		Host: "127.0.0.1",
		Port: listener.Addr().(*net.TCPAddr).Port,
	})
	if err != nil {
		t.Fatalf("create server: %v", err)
	}

	server.GetRouter().Listener = listener
	defer server.GetRouter().Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- serve(ctx, server)
	}()

	client := &http.Client{Timeout: 2 * time.Second}
	response, err := client.Get(
		"http://" + listener.Addr().String() + "/v1/health-check",
	)
	if err != nil {
		t.Fatalf("server did not respond before cancellation: %v", err)
	}
	response.Body.Close()

	if response.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200, got %d", response.StatusCode)
	}

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("expected clean shutdown, got: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("server did not stop after cancellation")
	}
}
