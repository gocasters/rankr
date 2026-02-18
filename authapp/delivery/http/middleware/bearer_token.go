package middleware

import (
	"net/http"

	"github.com/gocasters/rankr/authapp/service/tokenservice"
	"github.com/gocasters/rankr/pkg/authhttp"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
)

const accessClaimsContextKey = "auth_access_claims"

type RequireBearerTokenOptions struct {
	Skipper echomiddleware.Skipper
}

type Middleware struct {
	tokenService *tokenservice.AuthService
}

func New(tokenService *tokenservice.AuthService) Middleware {
	return Middleware{tokenService: tokenService}
}

func (m Middleware) RequireBearerToken(opts RequireBearerTokenOptions) echo.MiddlewareFunc {
	if opts.Skipper == nil {
		opts.Skipper = echomiddleware.DefaultSkipper
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if opts.Skipper(c) {
				return next(c)
			}

			token := authhttp.ExtractBearerToken(c.Request())
			if token == "" {
				return c.JSON(http.StatusUnauthorized, echo.Map{"error": "token is required"})
			}

			claims, err := m.tokenService.VerifyToken(token)
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

func AccessClaimsFromContext(c echo.Context) (*tokenservice.UserClaims, bool) {
	claims, ok := c.Get(accessClaimsContextKey).(*tokenservice.UserClaims)
	return claims, ok
}
