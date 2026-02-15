package role

import "strings"

type Permission string

const (
	PermissionAll          Permission = "*"
	PermissionUnresolvable Permission = "unresolvable"
)

func HasPermission(access []string, permission Permission) bool {
	if permission == PermissionUnresolvable {
		return false
	}
	if permission == "" {
		return true
	}
	for _, perm := range access {
		if perm == string(PermissionAll) || perm == string(permission) {
			return true
		}
	}
	return false
}

func RequiredPermission(method, path, host string) Permission {
	if path == "" {
		return PermissionUnresolvable
	}
	if isPublicPath(path) {
		return ""
	}

	module := moduleFromHost(host)
	if module == "" {
		module = moduleFromPath(path)
	}
	operation := operationFromMethod(method)
	if module == "" || operation == "" {
		return PermissionUnresolvable
	}

	return Permission(module + ":" + operation)
}

func isPublicPath(path string) bool {
	normalizedPath := strings.Trim(trimQuery(path), " /")
	if normalizedPath == "" {
		return true
	}

	publicEndpoints := []string{"v1/login", "v1/me"}
	for _, endpoint := range publicEndpoints {
		if normalizedPath == endpoint {
			return true
		}
	}

	segments := strings.Split(normalizedPath, "/")
	if len(segments) == 0 {
		return false
	}
	lastSegment := segments[len(segments)-1]
	return lastSegment == "health-check" || lastSegment == "health_check"
}

func operationFromMethod(method string) string {
	switch strings.ToUpper(method) {
	case "GET", "HEAD":
		return "read"
	case "POST":
		return "create"
	case "PUT", "PATCH":
		return "update"
	case "DELETE":
		return "delete"
	default:
		return ""
	}
}

func moduleFromHost(host string) string {
	host = strings.TrimSpace(host)
	if host == "" {
		return ""
	}
	if idx := strings.Index(host, ":"); idx >= 0 {
		host = host[:idx]
	}
	host = strings.Trim(host, ".")
	if host == "" {
		return ""
	}
	parts := strings.Split(host, ".")
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}

func moduleFromPath(path string) string {
	path = strings.Trim(trimQuery(path), "/")
	if path == "" {
		return ""
	}
	if strings.HasPrefix(path, "v1/") {
		path = strings.TrimPrefix(path, "v1/")
	}
	if path == "" {
		return ""
	}
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return ""
	}
	return strings.ToLower(parts[0])
}

func trimQuery(path string) string {
	if idx := strings.Index(path, "?"); idx >= 0 {
		return path[:idx]
	}
	return path
}
