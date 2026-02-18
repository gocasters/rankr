package http

import (
	"net/http"
	"strings"
	"time"

	authmiddleware "github.com/gocasters/rankr/authapp/delivery/http/middleware"
	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/authapp/service/tokenservice"
	"github.com/gocasters/rankr/pkg/authhttp"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/pkg/validator"
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

func (h Handler) me(c echo.Context) error {
	claims, ok := authmiddleware.AccessClaimsFromContext(c)
	if !ok || claims == nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}

	if requiredPermission := requiredPermissionFromRequest(c.Request()); requiredPermission != "" &&
		!role.HasPermission(claims.Access, requiredPermission) {
		return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden"})
	}

	setUserResponseHeaders(c, claims)
	return c.JSON(http.StatusOK, buildMeResponse(claims))
}

func (h Handler) refreshToken(c echo.Context) error {
	refresh := authhttp.ExtractRefreshToken(c.Request())
	if refresh == "" {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "refresh token is required"})
	}

	claims, err := h.tokenService.VerifyRefreshToken(refresh)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid refresh token"})
	}
	if _, ok := role.Parse(claims.Role); !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid role"})
	}

	accessToken, refreshToken, issueErr := h.tokenService.IssueTokens(claims.UserID, claims.Role, claims.Access)
	if issueErr != nil {
		return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to issue tokens"})
	}

	return c.JSON(http.StatusOK, echo.Map{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
	})
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

func requiredPermissionFromRequest(req *http.Request) role.Permission {
	originalURI := req.Header.Get("X-Original-URI")
	originalMethod := req.Header.Get("X-Original-Method")
	if originalURI == "" && originalMethod == "" {
		return role.Permission("")
	}

	originalHost := req.Header.Get("X-Original-Host")
	if originalHost == "" {
		originalHost = req.Host
	}
	return role.RequiredPermission(originalMethod, originalURI, originalHost)
}

func buildMeResponse(claims *tokenservice.UserClaims) echo.Map {
	response := echo.Map{
		"user_id": claims.UserID,
		"role":    claims.Role,
		"access":  claims.Access,
	}
	if claims.RegisteredClaims.ExpiresAt != nil {
		response["expires_at"] = claims.RegisteredClaims.ExpiresAt.Time.Format(time.RFC3339)
	}
	if claims.RegisteredClaims.IssuedAt != nil {
		response["issued_at"] = claims.RegisteredClaims.IssuedAt.Time.Format(time.RFC3339)
	}
	return response
}

func setUserResponseHeaders(c echo.Context, claims *tokenservice.UserClaims) {
	c.Response().Header().Set("X-User-ID", claims.UserID)
	c.Response().Header().Set("X-Role", claims.Role)
	if len(claims.Access) > 0 {
		c.Response().Header().Set("X-Access", strings.Join(claims.Access, ","))
	}
	if encoded, err := authhttp.EncodeUserInfo(claims.UserID); err == nil {
		c.Response().Header().Set("X-User-Info", encoded)
	}
}
