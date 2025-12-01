package http

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/authapp/service/tokenservice"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
	types "github.com/gocasters/rankr/type"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	authService  auth.Service
	tokenService *tokenservice.AuthService
}

func NewHandler(authSrv auth.Service, tokenSrv *tokenservice.AuthService) Handler {
	return Handler{
		authService:  authSrv,
		tokenService: tokenSrv,
	}
}

func (h Handler) login(c echo.Context) error {
	var req auth.LoginRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}

	res, err := h.authService.Login(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}

	return c.JSON(http.StatusOK, res)
}

func (h Handler) verifyToken(c echo.Context) error {
	token := extractBearerToken(c.Request())
	if token == "" {
		var body struct {
			Token string `json:"token"`
		}
		if err := c.Bind(&body); err != nil {
			return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request"})
		}
		token = body.Token
	}
	if token == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "token is required"})
	}

	claims, err := h.tokenService.VerifyToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	response := echo.Map{
		"user_id": claims.UserID,
		"role":    claims.Role,
	}
	if claims.RegisteredClaims.ExpiresAt != nil {
		response["expires_at"] = claims.RegisteredClaims.ExpiresAt.Time.Format(time.RFC3339)
	}
	if claims.RegisteredClaims.IssuedAt != nil {
		response["issued_at"] = claims.RegisteredClaims.IssuedAt.Time.Format(time.RFC3339)
	}
	return c.JSON(http.StatusOK, response)
}

func (h Handler) listRoles(c echo.Context) error {
	page, pageSize, err := parsePagination(c.QueryParam("page"), c.QueryParam("page_size"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	res, err := h.authService.ListRoles(c.Request().Context(), auth.ListRoleRequest{
		Page:     page,
		PageSize: pageSize,
	})
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h Handler) getRole(c echo.Context) error {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	res, err := h.authService.GetRole(c.Request().Context(), auth.GetRoleRequest{RoleID: id})
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h Handler) createRole(c echo.Context) error {
	var req auth.CreateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	res, err := h.authService.CreateRole(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusCreated, res)
}

func (h Handler) updateRole(c echo.Context) error {
	var req auth.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	res, err := h.authService.UpdateRole(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h Handler) deleteRole(c echo.Context) error {
	id, err := parseIDParam(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": err.Error()})
	}
	res, err := h.authService.DeleteRole(c.Request().Context(), auth.DeleteRoleRequest{RoleID: id})
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h Handler) addPermissionToRole(c echo.Context) error {
	var req auth.AddPermissionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	res, err := h.authService.AddPermissionToRole(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h Handler) removePermissionFromRole(c echo.Context) error {
	var req auth.RemovePermissionRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "invalid request body"})
	}
	res, err := h.authService.RemovePermissionFromRole(c.Request().Context(), req)
	if err != nil {
		return h.handleError(c, err)
	}
	return c.JSON(http.StatusOK, res)
}

func (h Handler) healthCheck(c echo.Context) error {
	return c.JSON(http.StatusOK, echo.Map{"status": "ok"})
}

func (h Handler) handleError(c echo.Context, err error) error {
	if vErr, ok := err.(validator.Error); ok {
		return c.JSON(vErr.StatusCode(), vErr)
	}
	if eResp, ok := err.(errmsg.ErrorResponse); ok {
		return c.JSON(statuscode.MapToHTTPStatusCode(eResp), eResp)
	}
	return c.JSON(http.StatusInternalServerError, echo.Map{"error": err.Error()})
}

func parseIDParam(raw string) (types.ID, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return types.ID(id), nil
}

func parsePagination(rawPage, rawPageSize string) (int, int, error) {
	page := 0
	pageSize := 0
	var err error
	if rawPage != "" {
		if page, err = strconv.Atoi(rawPage); err != nil {
			return 0, 0, err
		}
	}
	if rawPageSize != "" {
		if pageSize, err = strconv.Atoi(rawPageSize); err != nil {
			return 0, 0, err
		}
	}
	return page, pageSize, nil
}

func extractBearerToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if len(authz) > 7 && authz[:7] == "Bearer " {
		return authz[7:]
	}
	return ""
}
