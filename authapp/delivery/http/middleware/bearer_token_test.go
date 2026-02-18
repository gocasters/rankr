package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gocasters/rankr/authapp/service/tokenservice"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/labstack/echo/v4"
)

func TestRequireBearerToken(t *testing.T) {
	t.Parallel()

	tokenSvc := tokenservice.NewAuthService("test-secret", time.Hour, 2*time.Hour)
	validToken, err := tokenSvc.IssueToken("42", string(role.User), []string{"project:read"})
	if err != nil {
		t.Fatalf("failed to issue valid token: %v", err)
	}
	invalidRoleToken, err := tokenSvc.IssueToken("42", "invalid-role", nil)
	if err != nil {
		t.Fatalf("failed to issue invalid-role token: %v", err)
	}

	tests := []struct {
		name       string
		authz      string
		wantStatus int
		wantCalled bool
	}{
		{
			name:       "missing authorization header",
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "invalid token",
			authz:      "Bearer not-a-valid-token",
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "invalid role in claims",
			authz:      "Bearer " + invalidRoleToken,
			wantStatus: http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:       "valid token",
			authz:      "Bearer " + validToken,
			wantStatus: http.StatusNoContent,
			wantCalled: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/v1/me", nil)
			if tc.authz != "" {
				req.Header.Set("Authorization", tc.authz)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			nextCalled := false
			mid := New(tokenSvc)
			mw := mid.RequireBearerToken(RequireBearerTokenOptions{})(func(c echo.Context) error {
				nextCalled = true
				claims, ok := AccessClaimsFromContext(c)
				if !ok || claims == nil {
					t.Fatal("claims were not set in context")
				}
				if claims.UserID != "42" {
					t.Fatalf("unexpected user id: got %q", claims.UserID)
				}
				if claims.Role != string(role.User) {
					t.Fatalf("unexpected role: got %q", claims.Role)
				}
				return c.NoContent(http.StatusNoContent)
			})

			if err := mw(c); err != nil {
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
