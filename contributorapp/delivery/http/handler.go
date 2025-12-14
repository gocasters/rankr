package http

import (
	"fmt"
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"github.com/gocasters/rankr/contributorapp/service/job"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
	"log/slog"
	"net/http"
	"strconv"
)

type Handler struct {
	ContributorService contributor.Service
	JobService         job.Service
	Logger             *slog.Logger
}

func NewHandler(contributorSrv contributor.Service, jobSvc job.Service, logger *slog.Logger) Handler {
	return Handler{
		ContributorService: contributorSrv,
		JobService:         jobSvc,
		Logger:             logger,
	}
}

func (h Handler) getProfile(c echo.Context) error {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)

	if idStr == "" || err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "id is required"})
	}

	res, err := h.ContributorService.GetProfile(c.Request().Context(), types.ID(id))
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}
		return c.JSON(statuscode.MapToHTTPStatusCode(err.(errmsg.ErrorResponse)), err)
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) createContributor(c echo.Context) error {

	var req contributor.CreateContributorRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	res, err := h.ContributorService.CreateContributor(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eResp, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eResp), eResp)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) updateProfile(c echo.Context) error {
	var req contributor.UpdateProfileRequest

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	res, err := h.ContributorService.UpdateProfile(c.Request().Context(), req)
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}

		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data": res,
	})
}

func (h Handler) uploadFile(c echo.Context) error {
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
	defer srcFile.Close()

	fileType, _ := c.Get("FileType").(string)

	res, err := h.JobService.CreateImportJob(c.Request().Context(), contributor.ImportContributorRequest{
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
