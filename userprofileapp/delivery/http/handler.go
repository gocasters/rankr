package http

import (
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
	"github.com/gocasters/rankr/userprofileapp/service/userprofile"
	"github.com/labstack/echo/v4"
	"net/http"
	"strconv"
)

type Handler struct {
	service userprofile.Service
}

func NewHandler(srv userprofile.Service) Handler {
	return Handler{service: srv}
}

func (s Server) profile(c echo.Context) error {
	idStr := c.Param("id")
	if idStr == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": ErrMsgIDMustNotBeEmpty,
		})
	}

	userID, err := strconv.Atoi(idStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{
			"error": ErrMsgInvalidID,
		})
	}

	profileResponse, err := s.Handler.service.GetUserProfile(c.Request().Context(), int64(userID))
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if resErr, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(resErr), resErr)
		}

		return c.JSON(http.StatusInternalServerError, echo.Map{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": profileResponse,
	})
}
