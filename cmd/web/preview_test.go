package main

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gocasters/rankr/pkg/httpserver"
)

func TestLeaderboardPreview(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		enabled bool
		status  int
		want    string
	}{
		{"disabled", "/preview", false, 404, ""},
		{"rows", "/preview", true, 200, "پیش‌نمایش با دادهٔ آزمایشی"},
		{"empty", "/preview?state=empty", true, 200, "هنوز امتیازی ثبت نشده است."},
		{"error", "/preview?state=error", true, 200, "دریافت اطلاعات انجام نشد."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, err := newServer(httpserver.Config{
				Port:        8085,
				PublicPaths: []string{"/preview"},
			})
			if err != nil {
				t.Fatalf("create server: %v", err)
			}

			registerLeaderboardPreview(server.GetRouter(), tt.enabled)

			request := httptest.NewRequest("GET", tt.path, nil)
			response := httptest.NewRecorder()
			server.GetRouter().ServeHTTP(response, request)

			if response.Code != tt.status {
				t.Fatalf("expected status %d, got %d",
					tt.status, response.Code)
			}
			if tt.want != "" && !strings.Contains(response.Body.String(), tt.want) {
				t.Errorf("expected response containing %q", tt.want)
			}
		})
	}
}
