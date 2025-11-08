package http

import (
	"errors"
	"github.com/gocasters/rankr/notifapp/service/notification"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

var notificationID = "notification_id"

type Handler struct {
	service notification.Service
}

func NewHandler(svc notification.Service) Handler {
	return Handler{service: svc}
}

func getNotificationID(c echo.Context) (types.ID, error) {

	idStr := c.Param(notificationID)
	if idStr == "" {
		return 0, errors.New("invalid notification id in param")
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, errors.New("invalid notification id in param")
	}

	return types.ID(id), nil
}

func getUserID(c echo.Context) types.ID {
	return c.Get("userInfo").(types.UserClaim).ID
}

func (h Handler) createNotification(c echo.Context) error {

	var req notification.CreateRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.ErrBadRequest)
	}

	req.UserID = getUserID(c)

	res, err := h.service.Create(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, res)
}

func (h Handler) deleteNotification(c echo.Context) error {

	var req notification.DeleteRequest

	req.UserID = getUserID(c)

	notifyID, err := getNotificationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	req.NotificationID = notifyID

	err = h.service.Delete(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.NoContent(http.StatusNoContent)
}

func (h Handler) markAsRead(c echo.Context) error {

	var req notification.MarkAsReadRequest

	req.UserID = getUserID(c)

	notifyID, err := getNotificationID(c)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	req.NotificationID = notifyID

	res, err := h.service.MarkAsRead(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) markAllAsRead(c echo.Context) error {

	var req notification.MarkAllAsReadRequest

	req.UserID = getUserID(c)

	err := h.service.MarkAllAsRead(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "Successfully done all notifications as read"})
}

func (h Handler) getNotification(c echo.Context) error {

	var req notification.GetRequest

	notifyID, err := getNotificationID(c)

	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}

	req.NotificationID = notifyID
	req.UserID = getUserID(c)

	res, err := h.service.Get(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) listNotification(c echo.Context) error {

	var req notification.ListRequest

	req.UserID = getUserID(c)

	res, err := h.service.List(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) getUnreadCount(c echo.Context) error {

	var req notification.CountUnreadRequest

	req.UserID = getUserID(c)

	res, err := h.service.GetUnreadCount(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}
