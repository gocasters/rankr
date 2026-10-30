package leaderboardstat

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestHTTPClientGetPublicLeaderboard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/v1/leaderboard/public/5001" {
				t.Errorf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("page_size") != "10" {
				t.Error("expected page_size=10")
			}
			if r.URL.Query().Get("offset") != "20" {
				t.Error("expected offset=20")
			}

			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{
				"project_id": "5001",
				"rows": [{"rank": 1, "user_id": 101, "score": 12}]
			}`)
		},
	))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, time.Second)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	result, err := client.GetPublicLeaderboard(t.Context(), 5001, 10, 20)
	if err != nil {
		t.Fatalf("get leaderboard: %v", err)
	}

	if result.ProjectID != "5001" {
		t.Fatalf("expected project 5001, got %q", result.ProjectID)
	}
	if len(result.Rows) != 1 {
		t.Fatalf("expected one row, got %d", len(result.Rows))
	}

	want := PublicLeaderboardRow{Rank: 1, UserID: 101, Score: 12}
	if result.Rows[0] != want {
		t.Errorf("expected row %+v, got %+v", want, result.Rows[0])
	}
}

func TestHTTPClientGetPublicLeaderboardErrors(t *testing.T) {
	tests := []struct {
		name      string
		status    int
		body      string
		wantError string
	}{
		{
			name:      "server error",
			status:    http.StatusInternalServerError,
			body:      `{}`,
			wantError: "unexpected HTTP status: 500",
		},
		{
			name:      "unauthorized",
			status:    http.StatusUnauthorized,
			body:      `{}`,
			wantError: "unexpected HTTP status: 401",
		},
		{
			name:      "invalid JSON",
			status:    http.StatusOK,
			body:      `not-json`,
			wantError: "decode leaderboard",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.status)
					fmt.Fprint(w, tt.body)
				},
			))
			defer server.Close()

			client, err := NewHTTPClient(server.URL, time.Second)
			if err != nil {
				t.Fatalf("create client: %v", err)
			}

			_, err = client.GetPublicLeaderboard(t.Context(), 5001, 10, 0)
			if err == nil || !strings.Contains(err.Error(), tt.wantError) {
				t.Fatalf("expected error containing %q, got: %v",
					tt.wantError, err)
			}
		})
	}
}

func TestHTTPClientGetPublicLeaderboardTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			<-r.Context().Done()
		},
	))
	defer server.Close()

	client, err := NewHTTPClient(server.URL, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("create client: %v", err)
	}

	ctx, cancel := context.WithTimeout(t.Context(), 3*time.Second)
	defer cancel()

	_, err = client.GetPublicLeaderboard(ctx, 5001, 10, 0)
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("expected timeout error, got: %v", err)
	}

	if ctx.Err() != nil {
		t.Fatal("parent context expired before the client timeout worked")
	}
}
