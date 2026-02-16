package http

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/gocasters/rankr/pkg/authhttp"
)

func TestExtractRefreshToken(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(*http.Request)
		wantTok string
	}{
		{
			name: "from x refresh token header",
			setup: func(r *http.Request) {
				r.Header.Set("X-Refresh-Token", "header-token")
			},
			wantTok: "header-token",
		},
		{
			name: "from refresh token header",
			setup: func(r *http.Request) {
				r.Header.Set("Refresh-Token", "header-token-2")
			},
			wantTok: "header-token-2",
		},
		{
			name: "from cookie",
			setup: func(r *http.Request) {
				r.AddCookie(&http.Cookie{Name: "refresh_token", Value: "cookie-token"})
			},
			wantTok: "cookie-token",
		},
		{
			name: "query parameter is ignored",
			setup: func(r *http.Request) {
				r.URL = &url.URL{RawQuery: "refresh_token=query-token"}
			},
			wantTok: "",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			req, err := http.NewRequest(http.MethodPost, "http://localhost/auth/v1/refresh-token", nil)
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}
			if tc.setup != nil {
				tc.setup(req)
			}
			got := authhttp.ExtractRefreshToken(req)
			if got != tc.wantTok {
				t.Fatalf("authhttp.ExtractRefreshToken() = %q, want %q", got, tc.wantTok)
			}
		})
	}
}
