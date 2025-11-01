package http

import (
	"github.com/gocasters/rankr/notifapp/service/notification"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"strconv"
)

var (
	userID         = "user_id"
	notificationID = "notification_id"
)

type Handler struct {
	validator notification.Validate
	service   notification.Service
}

func NewHandler(svc notification.Service, validate notification.Validate) Handler {
	return Handler{validator: validate, service: svc}
}

func getID(c echo.Context, param string) (types.ID, error) {
	idStr := c.Param(param)
	if idStr == "" {
		return 0, echo.ErrBadRequest
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, echo.ErrBadRequest
	}

	return types.ID(id), nil
}
