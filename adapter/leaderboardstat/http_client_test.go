package leaderboardstat

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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
			if r.URL.Query().Get("offset") != "0" {
				t.Error("expected offset=0")
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

	result, err := client.GetPublicLeaderboard(t.Context(), 5001, 10, 0)
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
