package role

import "testing"

func TestIsPublicPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
		want bool
	}{
		{name: "empty path", path: "", want: true},
		{name: "login exact", path: "/v1/login", want: true},
		{name: "login with trailing slash", path: "/v1/login/", want: true},
		{name: "login callback subpath should not pass", path: "/v1/login/callback", want: false},
		{name: "me exact", path: "/v1/me", want: true},
		{name: "me settings subpath should not pass", path: "/v1/me/settings", want: false},
		{name: "health check exact segment", path: "/v1/projects/health-check", want: true},
		{name: "health check underscore segment", path: "/v1/projects/health_check", want: true},
		{name: "health check as substring should not pass", path: "/v1/projects/my-health-check-endpoint", want: false},
		{name: "health check suffix should not pass", path: "/v1/projects/health-check-admin", want: false},
		{name: "health check with query", path: "/v1/projects/health-check?verbose=1", want: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := isPublicPath(tc.path)
			if got != tc.want {
				t.Fatalf("isPublicPath(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestRequiredPermissionFailClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		path   string
		host   string
		want   Permission
	}{
		{name: "public path", method: "GET", path: "/v1/me", host: "auth.local", want: ""},
		{name: "unresolvable empty path", method: "GET", path: "", host: "project.local", want: PermissionUnresolvable},
		{name: "unresolvable method", method: "TRACE", path: "/v1/projects", host: "project.local", want: PermissionUnresolvable},
		{name: "unresolvable module", method: "GET", path: "/v1//foo", host: "", want: PermissionUnresolvable},
		{name: "resolvable", method: "GET", path: "/v1/projects", host: "project.local", want: Permission("project:read")},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := RequiredPermission(tc.method, tc.path, tc.host)
			if got != tc.want {
				t.Fatalf("RequiredPermission(%q, %q, %q) = %q, want %q", tc.method, tc.path, tc.host, got, tc.want)
			}
		})
	}
}

func TestHasPermissionUnresolvableAlwaysDenied(t *testing.T) {
	t.Parallel()

	if HasPermission([]string{"*"}, PermissionUnresolvable) {
		t.Fatal("PermissionUnresolvable must be denied even for wildcard access")
	}
	if HasPermission([]string{"project:read"}, PermissionUnresolvable) {
		t.Fatal("PermissionUnresolvable must be denied")
	}
}
