package http

import (
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
)

type Handler struct {
	ContributorService contributor.Service
	Logger             *slog.Logger
}

func NewHandler(contributorSrv contributor.Service, logger *slog.Logger) Handler {
	return Handler{
		ContributorService: contributorSrv,
		Logger:             logger,
	}
}

func (h Handler) getProfile(c echo.Context) error {

	// TODO complete user auth with token
	userId := uint64(1)

	profileRequest := contributor.GetProfileRequest{
		ID: types.ID(userId),
	}

	res, err := h.ContributorService.GetProfile(c.Request().Context(), profileRequest)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}
		return c.JSON(statuscode.MapToHTTPStatusCode(err.(errmsg.ErrorResponse)), err)
	}

	return c.JSON(http.StatusOK, res)
}
