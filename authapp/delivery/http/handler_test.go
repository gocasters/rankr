package http

import (
	"net/http"
	"net/url"
	"testing"
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
			req, err := http.NewRequest(http.MethodGet, "http://localhost/auth/v1/me", nil)
			if err != nil {
				t.Fatalf("failed to build request: %v", err)
			}
			if tc.setup != nil {
				tc.setup(req)
			}
			got := extractRefreshToken(req)
			if got != tc.wantTok {
				t.Fatalf("extractRefreshToken() = %q, want %q", got, tc.wantTok)
			}
		})
	}
}

func TestSameAccessList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		a    []string
		b    []string
		want bool
	}{
		{name: "same ordered", a: []string{"a:read", "b:update"}, b: []string{"a:read", "b:update"}, want: true},
		{name: "same unordered", a: []string{"a:read", "b:update"}, b: []string{"b:update", "a:read"}, want: true},
		{name: "different values", a: []string{"a:read"}, b: []string{"a:update"}, want: false},
		{name: "different length", a: []string{"a:read"}, b: []string{"a:read", "b:update"}, want: false},
		{name: "both empty", a: nil, b: nil, want: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := sameAccessList(tc.a, tc.b)
			if got != tc.want {
				t.Fatalf("sameAccessList(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}
