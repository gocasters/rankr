package middleware

import (
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

const AdminRoleName = "admin"

func (m Middleware) CheckRole(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		claimVal := c.Get("UserInfo")
		claim, ok := claimVal.(*types.UserClaim)
		if !ok || claim == nil {
			return c.JSON(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
		}

		idStr := strconv.Itoa(int(claim.ID))

		resAuth, err := m.Client.GetRole(c.Request().Context(), idStr)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"message": "failed to fetch role",
			})
		}

		if resAuth.Role.Name != AdminRoleName {
			return c.JSON(http.StatusForbidden, map[string]string{
				"message": "forbidden",
			})
		}

		return next(c)
	}
}
