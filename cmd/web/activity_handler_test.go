package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gocasters/rankr/pkg/httpserver"
)

func TestProfileActivityPage(t *testing.T) {
	tests := []struct {
		name           string
		userID         string
		enabled        bool
		upstreamStatus int
		upstreamBody   string
		wantStatus     int
		wantText       string
	}{
		{
			name: "disabled", userID: "90401",
			enabled: false, upstreamStatus: 200,
			wantStatus: 404,
		},
		{
			name: "feed", userID: "90401",
			enabled: true, upstreamStatus: 200,
			upstreamBody: `{"user_id":"90401","limit":20,"activity":[
				{"type":"pull_request_closed","repository":"gocasters/rankr",
				 "subject_type":"pull_request","subject_number":242,
				 "points":10,"time":"2026-11-18T10:30:00Z"}
			]}`,
			wantStatus: 200, wantText: "بستن پول‌ریکوئست",
		},
		{
			name: "feed shows subject", userID: "90401",
			enabled: true, upstreamStatus: 200,
			upstreamBody: `{"user_id":"90401","limit":20,"activity":[
				{"type":"issue_closed","repository":"gocasters/rankr",
				 "subject_type":"issue","subject_number":242,
				 "points":2,"time":"2026-11-18T10:30:00Z"}
			]}`,
			wantStatus: 200, wantText: "Issue #242",
		},
		{
			name: "empty", userID: "90401",
			enabled: true, upstreamStatus: 200,
			upstreamBody: `{"user_id":"90401","limit":20,"activity":[]}`,
			wantStatus:   200, wantText: "هنوز فعالیتی ثبت نشده است.",
		},
		{
			name: "upstream error", userID: "90401",
			enabled: true, upstreamStatus: 500,
			wantStatus: 502, wantText: "دریافت فعالیت‌ها انجام نشد.",
		},
		{
			name: "invalid user", userID: "abc",
			enabled: true, upstreamStatus: 200,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet ||
						r.URL.Path != "/v1/contributors/90401/activity" {
						t.Errorf("unexpected API request: %s %s",
							r.Method, r.URL.Path)
					}
					if r.URL.Query().Get("limit") != "20" {
						t.Errorf("unexpected API query: %s", r.URL.RawQuery)
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.upstreamStatus)
					_, _ = w.Write([]byte(tt.upstreamBody))
				},
			))
			defer upstream.Close()

			server, err := newServer(httpserver.Config{
				Port:        8085,
				PublicPaths: []string{"/profile"},
			})
			if err != nil {
				t.Fatal(err)
			}

			if tt.enabled {
				client, err := newActivityClient(upstream.URL, time.Second)
				if err != nil {
					t.Fatal(err)
				}

				registerActivity(server.GetRouter(), client, true)
			} else {
				registerActivity(server.GetRouter(), nil, false)
			}

			request := httptest.NewRequest(
				http.MethodGet,
				"/profile?user_id="+tt.userID,
				nil,
			)
			response := httptest.NewRecorder()
			server.GetRouter().ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d; body: %s",
					tt.wantStatus, response.Code, response.Body.String())
			}

			if tt.wantText != "" &&
				!strings.Contains(response.Body.String(), tt.wantText) {
				t.Errorf("expected HTML containing %q", tt.wantText)
			}
		})
	}
}
