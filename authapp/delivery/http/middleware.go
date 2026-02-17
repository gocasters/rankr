package http

import (
	"net/http"

	"github.com/gocasters/rankr/authapp/service/tokenservice"
	"github.com/gocasters/rankr/pkg/authhttp"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

const accessClaimsContextKey = "auth_access_claims"

type requireAccessClaimsConfig struct {
	Skipper middleware.Skipper
}

func (h Handler) requireAccessClaimsWithConfig(cfg requireAccessClaimsConfig) echo.MiddlewareFunc {
	if cfg.Skipper == nil {
		cfg.Skipper = middleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if cfg.Skipper(c) {
				return next(c)
			}

			token := authhttp.ExtractBearerToken(c.Request())
			if token == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "token is required"})
			}

			claims, err := h.tokenService.VerifyToken(token)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
			}
			if _, ok := role.Parse(claims.Role); !ok {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid role"})
			}

			c.Set(accessClaimsContextKey, claims)
			return next(c)
		}
	}
}

func (h Handler) requireAccessClaims(next echo.HandlerFunc) echo.HandlerFunc {
	return h.requireAccessClaimsWithConfig(requireAccessClaimsConfig{})(next)
}

func accessClaimsFromContext(c echo.Context) (*tokenservice.UserClaims, bool) {
	claims, ok := c.Get(accessClaimsContextKey).(*tokenservice.UserClaims)
	return claims, ok && claims != nil
}
