package adminhttp

import (
	"fmt"
	"github.com/gocasters/rankr/contributorapp/dashboard"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"net/http"
)

type Handler struct {
	service dashboard.Service
}

func NewHandler(svc dashboard.Service) Handler {
	return Handler{service: svc}
}

func (h Handler) importContributors(c echo.Context) error {

	claim := c.Get("Authorization").(*types.UserClaim)
	if claim.Role.String() != types.Admin.String() {
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "unauthorized",
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"message": "fail to get file",
			"error":   err.Error(),
		})
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": fmt.Sprintf("failed to open file: %v", err),
		})
	}

	fileType, _ := c.Get("FileType").(string)

	res, err := h.service.ImportJob(c.Request().Context(), dashboard.ImportJobRequest{
		File:     srcFile,
		FileName: fileHeader.Filename,
		FileType: fileType,
	})
	if err != nil {
		if vEer, ok := err.(validator.Error); ok {
			return c.JSON(vEer.StatusCode(), map[string]interface{}{
				"message": vEer.Err,
				"errors":  vEer.Fields,
			})
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, res)
}
