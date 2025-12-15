package http

import (
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct{}

func NewHandler() Handler {
	return Handler{}
}

func (h Handler) HealthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}
