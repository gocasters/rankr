package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gocasters/rankr/adapter/leaderboardstat"
	"github.com/gocasters/rankr/pkg/httpserver"
)

func TestLeaderboardPage(t *testing.T) {
	tests := []struct {
		name           string
		projectID      string
		enabled        bool
		upstreamStatus int
		upstreamBody   string
		wantStatus     int
		wantText       string
	}{
		{
			name: "disabled", projectID: "1028435569",
			enabled: false, upstreamStatus: 200,
			wantStatus: 404,
		},
		{
			name: "rows", projectID: "1028435569",
			enabled: true, upstreamStatus: 200,
			upstreamBody: `{"project_id":"1028435569","rows":[
				{"rank":1,"user_id":90401,"score":37}
			]}`,
			wantStatus: 200, wantText: "<td>90401</td>",
		},
		{
			name: "empty", projectID: "1028435569",
			enabled: true, upstreamStatus: 200,
			upstreamBody: `{"project_id":"1028435569","rows":[]}`,
			wantStatus:   200, wantText: "هنوز امتیازی ثبت نشده است.",
		},
		{
			name: "upstream error", projectID: "1028435569",
			enabled: true, upstreamStatus: 500,
			wantStatus: 502, wantText: "دریافت اطلاعات انجام نشد.",
		},
		{
			name: "invalid project", projectID: "abc",
			enabled: true, upstreamStatus: 200,
			wantStatus: 400,
		},
		{
			name:    "missing project",
			enabled: true, upstreamStatus: 200,
			wantStatus: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			upstream := httptest.NewServer(http.HandlerFunc(
				func(w http.ResponseWriter, r *http.Request) {
					if r.Method != http.MethodGet ||
						r.URL.Path != "/v1/leaderboard/public/1028435569" {
						t.Errorf("unexpected API request: %s %s",
							r.Method, r.URL.Path)
					}
					if r.URL.Query().Get("page_size") != "10" ||
						r.URL.Query().Get("offset") != "0" {
						t.Errorf("unexpected API query: %s", r.URL.RawQuery)
					}

					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(tt.upstreamStatus)
					_, _ = w.Write([]byte(tt.upstreamBody))
				},
			))
			defer upstream.Close()

			client, err := leaderboardstat.NewHTTPClient(
				upstream.URL, time.Second,
			)
			if err != nil {
				t.Fatal(err)
			}

			server, err := newServer(httpserver.Config{
				Port:        8085,
				PublicPaths: []string{"/leaderboard"},
			})
			if err != nil {
				t.Fatal(err)
			}

			registerLeaderboard(server.GetRouter(), client, tt.enabled)

			request := httptest.NewRequest(
				http.MethodGet,
				"/leaderboard?project_id="+tt.projectID,
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
