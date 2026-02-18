package echomiddleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
)

func TestRequireUserInfo(t *testing.T) {
	t.Parallel()

	e := echo.New()

	claimsRaw, err := json.Marshal(types.UserClaim{ID: 42})
	if err != nil {
		t.Fatalf("failed to marshal user claim: %v", err)
	}
	encodedClaims := base64.StdEncoding.EncodeToString(claimsRaw)

	tests := []struct {
		name       string
		path       string
		userInfo   string
		middleware echo.MiddlewareFunc
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "missing claims header",
			path:       "/v1/private",
			middleware: RequireUserInfo(RequireUserInfoOptions{}),
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "valid claims header",
			path:       "/v1/private",
			userInfo:   encodedClaims,
			middleware: RequireUserInfo(RequireUserInfoOptions{}),
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
		{
			name: "skip configured path",
			path: "/v1/health-check",
			middleware: RequireUserInfo(RequireUserInfoOptions{
				Skipper: SkipExactPaths("/v1/health-check"),
			}),
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			req := httptest.NewRequest(http.MethodGet, tc.path, nil)
			if tc.userInfo != "" {
				req.Header.Set("X-User-Info", tc.userInfo)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			next := func(c echo.Context) error {
				nextCalled = true
				return c.NoContent(http.StatusNoContent)
			}

			if err := tc.middleware(next)(c); err != nil {
				t.Fatalf("middleware returned error: %v", err)
			}

			if rec.Code != tc.wantStatus {
				t.Fatalf("status code = %d, want %d", rec.Code, tc.wantStatus)
			}
			if nextCalled != tc.wantCalled {
				t.Fatalf("next called = %v, want %v", nextCalled, tc.wantCalled)
			}
		})
	}
}

func TestSkipExactPaths_NormalizesPaths(t *testing.T) {
	t.Parallel()

	skipper := SkipExactPaths("v1/health-check")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/v1/health-check/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if !skipper(c) {
		t.Fatal("expected path to be skipped")
	}
}
