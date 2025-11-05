package http

import (
	"errors"
	"net/http"
	"strconv"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gocasters/rankr/authapp/service/auth"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	GrantService auth.Service
}

func NewHandler(grantSrv auth.Service) Handler {
	return Handler{
		GrantService: grantSrv,
	}
}

func (h Handler) getGrant(c echo.Context) error {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	grant, err := h.GrantService.GetGrant(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, auth.ErrGrantNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "grant not found"})
		}
		if isValidationError(err) {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}

		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to fetch grant"})
	}

	return c.JSON(http.StatusOK, grant)
}

func (h Handler) createGrant(c echo.Context) error {
	var req auth.CreateGrantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	grant, err := h.GrantService.CreateGrant(c.Request().Context(), req)
	if err != nil {
		if isValidationError(err) {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to create grant"})
	}

	return c.JSON(http.StatusCreated, grant)
}

func (h Handler) updateGrant(c echo.Context) error {
	var req auth.UpdateGrantRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request body"})
	}

	grant, err := h.GrantService.UpdateGrant(c.Request().Context(), req)
	if err != nil {
		if errors.Is(err, auth.ErrGrantNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "grant not found"})
		}
		if isValidationError(err) {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to update grant"})
	}

	return c.JSON(http.StatusOK, grant)
}

func (h Handler) deleteGrant(c echo.Context) error {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	if err := h.GrantService.DeleteGrant(c.Request().Context(), id); err != nil {
		if errors.Is(err, auth.ErrGrantNotFound) {
			return c.JSON(http.StatusNotFound, map[string]string{"error": "grant not found"})
		}
		if isValidationError(err) {
			return c.JSON(http.StatusBadRequest, map[string]any{"error": err.Error()})
		}
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to delete grant"})
	}

	return c.JSON(http.StatusOK, map[string]string{"message": "grant deleted successfully"})
}

func (h Handler) listGrants(c echo.Context) error {
	grants, err := h.GrantService.ListGrants(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "failed to list grants"})
	}

	res := make([]auth.GrantResponse, 0, len(grants))
	for _, grant := range grants {
		res = append(res, auth.GrantResponse{
			ID:        grant.ID,
			Subject:   grant.Subject,
			Object:    grant.Object,
			Action:    grant.Action,
			Field:     grant.Field,
			CreatedAt: grant.CreatedAt,
			UpdatedAt: grant.UpdatedAt,
		})
	}

	return c.JSON(http.StatusOK, auth.ListGrantsResponse{
		Grants: res,
		Total:  len(res),
	})
}

func (h Handler) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func parseIDParam(raw string) (types.ID, error) {
	if raw == "" {
		return 0, errors.New("id is required")
	}
	value, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, errors.New("id must be a positive integer")
	}
	return types.ID(value), nil
}

func isValidationError(err error) bool {
	_, ok := err.(validation.Errors)
	if ok {
		return true
	}
	_, ok = err.(*validation.Errors)
	return ok
}
