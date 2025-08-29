package http

import (
	"net/http"

	echo "github.com/labstack/echo/v4"
)

func (s Server) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{
		"message": "everything is good!",
	})
}
