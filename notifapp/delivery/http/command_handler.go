package http

import (
	"github.com/gocasters/rankr/notifapp/service/notification"
	"github.com/labstack/echo/v4"
	"net/http"
)

func (h Handler) createNotification(c echo.Context) error {

	var req notification.CreateRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	userId, err := getID(c, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = userId

	if fieldErr, err := h.validator.CreateNotificationValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	res, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, echo.Map{"data": res})
}

func (h Handler) deleteNotification(c echo.Context) error {

	var req notification.DeleteRequest

	userId, errUserId := getID(c, userID)
	notifyID, errNotifyID := getID(c, notificationID)

	if errNotifyID != nil || errUserId != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = userId
	req.NotificationID = notifyID

	if fieldErr, err := h.validator.DeleteNotificationValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	err := h.service.Delete(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h Handler) markAsRead(c echo.Context) error {

	var req notification.MarkAsReadRequest

	userId, errUserId := getID(c, userID)
	notifyID, errNotifyID := getID(c, notificationID)

	if errNotifyID != nil || errUserId != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = userId
	req.NotificationID = notifyID

	if fieldErr, err := h.validator.MarkAsReadNotificationValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	res, err := h.service.MarkAsRead(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"data": res})
}

func (h Handler) markAllAsRead(c echo.Context) error {

	var req notification.MarkAllAsReadRequest

	userId, err := getID(c, userID)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = userId

	if fieldErr, err := h.validator.MarkAllAsReadNotificationValidate(req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"message": err.Error(), "error": fieldErr})
	}

	err = h.service.MarkAllAsRead(c.Request().Context(), req)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully done all notifications as read"})
}
