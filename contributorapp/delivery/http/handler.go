package http

import (
	"fmt"
	"github.com/gocasters/rankr/contributorapp/client"
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
	Client             client.AuthClient
	Logger             *slog.Logger
}

func NewHandler(contributorSrv contributor.Service, jobSvc job.Service,
	client client.AuthClient, logger *slog.Logger) Handler {
	return Handler{
		ContributorService: contributorSrv,
		JobService:         jobSvc,
		Client:             client,
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
	claimVal := c.Get("Authorization")
	claim, ok := claimVal.(*types.UserClaim)
	if !ok || claim == nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "unauthorized",
		})
	}

	idStr := strconv.Itoa(int(claim.ID))

	resAuth, err := h.Client.GetRole(c.Request().Context(), idStr)
	if err != nil {
		h.Logger.Error("failed to get role", "error", err)
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"message": "unauthorized",
		})
	}

	if resAuth.Role.Name != "admin" {
		return c.JSON(http.StatusForbidden, map[string]string{
			"message": "forbidden",
		})
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"message": "fail to get file",
			"error":   err.Error(),
		})
	}

	srcFile, err := fileHeader.Open()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": fmt.Sprintf("failed to open file: %v", err),
		})
	}
	defer func() {
		if closeErr := srcFile.Close(); closeErr != nil {
			h.Logger.Error("failed to close uploaded file",
				"error", closeErr, "filename",
				fileHeader.Filename)
		}
	}()

	fileType, ok := c.Get("FileType").(string)
	if !ok || fileType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "file type not detected by middleware",
		})
	}

	idempotencyKey := fmt.Sprintf("%s-%d-%d", fileHeader.Filename, fileHeader.Size, claim.ID)

	res, err := h.JobService.CreateImportJob(c.Request().Context(), job.ImportContributorRequest{
		File:           srcFile,
		FileName:       fileHeader.Filename,
		FileType:       fileType,
		IdempotencyKey: idempotencyKey,
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

		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, res)
}

func (h Handler) getJobStatus(c echo.Context) error {
	jobIDStr := c.Param("job_id")
	jobID, err := strconv.Atoi(jobIDStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid job id",
		})
	}

	res, err := h.JobService.JobStatus(c.Request().Context(), uint(jobID))
	if err != nil {
		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) updatePassword(c echo.Context) error {
	userIDHeader := c.Request().Header.Get("X-User-ID")
	if userIDHeader == "" {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "missing user id"})
	}

	userID, err := strconv.ParseUint(userIDHeader, 10, 64)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid user id"})
	}

	var body struct {
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	if err := c.Bind(&body); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	res, err := h.ContributorService.UpdatePassword(c.Request().Context(), contributor.UpdatePasswordRequest{
		ID:          types.ID(userID),
		OldPassword: body.OldPassword,
		NewPassword: body.NewPassword,
	})
	if err != nil {
		if vErr, ok := err.(validator.Error); ok {
			return c.JSON(vErr.StatusCode(), vErr)
		}
		if eResp, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eResp), eResp)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) getFailRecords(c echo.Context) error {
	jobIdStr := c.Param("job_id")
	jobID, err := strconv.Atoi(jobIdStr)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "invalid job id",
		})
	}

	records, err := h.JobService.GetFailRecords(c.Request().Context(), uint(jobID))
	if err != nil {
		if eRes, ok := err.(errmsg.ErrorResponse); ok {
			return c.JSON(statuscode.MapToHTTPStatusCode(eRes), eRes)
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusOK, records)
}
