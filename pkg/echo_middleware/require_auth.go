package echomiddleware

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"

	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
)

func RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		raw := c.Request().Header.Get("X-User-Info")
		if raw == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
		}

		decoded, err := base64.StdEncoding.DecodeString(raw)
		if err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
		}

		var info types.UserClaim
		if err := json.Unmarshal(decoded, &info); err != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
		}
		if !info.ID.IsValid() {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "unauthorized"})
		}

		// Make user info available to handlers and refresh X-User-ID header for downstream use.
		c.Set("userInfo", &info)
		c.Request().Header.Set("X-User-ID", strconv.FormatUint(uint64(info.ID), 10))

		return next(c)
	}
}
