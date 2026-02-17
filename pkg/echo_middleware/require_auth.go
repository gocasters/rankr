package echomiddleware

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const userInfoContextKey = "userInfo"

type RequireClaimsConfig struct {
	Skipper middleware.Skipper
}

// Backward-compatible alias.
type RequireAccessClaimsConfig = RequireClaimsConfig

func RequireClaimsWithConfig(cfg RequireClaimsConfig) echo.MiddlewareFunc {
	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			info, err := userInfoFromHeader(c.Request().Header.Get("X-User-Info"))
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
			}

		
			c.Set(userInfoContextKey, info)
			c.Request().Header.Set("X-User-ID", strconv.FormatUint(uint64(info.ID), 10))

			return next(c)
		}
	}
}

// Backward-compatible alias.
func RequireAccessClaimsWithConfig(cfg RequireAccessClaimsConfig) echo.MiddlewareFunc {
	return RequireClaimsWithConfig(cfg)
}

func RequireClaims(next echo.HandlerFunc) echo.HandlerFunc {
	return RequireClaimsWithConfig(RequireClaimsConfig{})(next)
}

func RequireAccessClaims(next echo.HandlerFunc) echo.HandlerFunc {
	return RequireClaims(next)
}

// RequireAuth remains as backward-compatible alias.
func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return RequireClaims(next)
}

func SkipExactPaths(paths ...string) middleware.Skipper {
	skipped := make(map[string]struct{}, len(paths))
	for _, path := range paths {
		normalized := normalizePath(path)
		if normalized == "" {
			continue
		}
		skipped[normalized] = struct{}{}
	}

	return func(c echo.Context) bool {
		path := normalizePath(c.Request().URL.Path)
		_, ok := skipped[path]
		return ok
	}
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func userInfoFromHeader(raw string) (*types.UserClaim, error) {
	if raw == "" {
		return nil, errors.New("missing x-user-info")
	}

	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return nil, err
	}

	var info types.UserClaim
	if err := json.Unmarshal(decoded, &info); err != nil {
		return nil, err
	}
	if !info.ID.IsValid() {
		return nil, errors.New("invalid user id in claim")
	}

	return &info, nil
}
