package http

import (
	"github.com/gocasters/rankr/notifapp/service/notification"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) getNotification(c echo.Context) error {

	var req notification.GetRequest

	userId, errUserId := getID(c, userID)
	notifyID, errNotifyID := getID(c, notificationID)

	if errUserId != nil || errNotifyID != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.NotificationID = notifyID
	req.UserID = userId

	if fieldErr, err := h.validator.GetNotificationValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	res, err := h.service.Get(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": res})
}

func (h Handler) listNotification(c echo.Context) error {

	var req notification.ListRequest

	userId, err := getID(c, userID)

	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = userId

	if fieldErr, err := h.validator.ListNotificationsValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	res, err := h.service.List(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": res})
}

func (h Handler) getUnreadCount(c echo.Context) error {

	var req notification.CountUnreadRequest

	userId, err := getID(c, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = userId

	if fieldErr, err := h.validator.GetUnreadCountNotificationValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	res, err := h.service.GetUnreadCount(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": res})
}
