package http

import (
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gocasters/rankr/authapp/service/auth"
	"github.com/gocasters/rankr/authapp/service/tokenservice"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/role"
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

func (h Handler) me(c echo.Context) error {
	token := extractBearerToken(c.Request())
	if token == "" {
		return c.JSON(http.StatusBadRequest, echo.Map{"error": "token is required"})
	}

	claims, err := h.tokenService.VerifyToken(token)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid token"})
	}
	if _, ok := role.Parse(claims.Role); !ok {
		return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid role"})
	}

	originalURI := c.Request().Header.Get("X-Original-URI")
	originalMethod := c.Request().Header.Get("X-Original-Method")
	originalHost := c.Request().Header.Get("X-Original-Host")
	if originalHost == "" {
		originalHost = c.Request().Host
	}

	if originalURI != "" || originalMethod != "" {
		requiredPermission := role.RequiredPermission(originalMethod, originalURI, originalHost)
		if !role.HasPermission(claims.Access, requiredPermission) {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "forbidden"})
		}
	}

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

	if originalURI == "" && originalMethod == "" {
		refreshToken := extractRefreshToken(c.Request())
		if refreshToken == "" {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "refresh token is required"})
		}
		refreshClaims, refreshErr := h.tokenService.VerifyRefreshToken(refreshToken)
		if refreshErr != nil {
			return c.JSON(http.StatusUnauthorized, echo.Map{"error": "invalid refresh token"})
		}
		if refreshClaims.UserID != claims.UserID || refreshClaims.Role != claims.Role || !sameAccessList(refreshClaims.Access, claims.Access) {
			return c.JSON(http.StatusForbidden, echo.Map{"error": "refresh token does not match access token"})
		}

		accessToken, refreshToken, issueErr := h.tokenService.IssueTokens(claims.UserID, claims.Role, claims.Access)
		if issueErr != nil {
			return c.JSON(http.StatusInternalServerError, echo.Map{"error": "failed to issue tokens"})
		}
		response["access_token"] = accessToken
		response["refresh_token"] = refreshToken
	}

	c.Response().Header().Set("X-User-ID", claims.UserID)
	c.Response().Header().Set("X-Role", claims.Role)
	if len(claims.Access) > 0 {
		c.Response().Header().Set("X-Access", strings.Join(claims.Access, ","))
	}
	if encoded, err := encodeUserInfo(claims.UserID); err == nil {
		c.Response().Header().Set("X-User-Info", encoded)
	}

	return c.JSON(http.StatusOK, response)
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

func encodeUserInfo(userID string) (string, error) {
	parsedID, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(types.UserClaim{ID: types.ID(parsedID)})
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(payload), nil
}

func extractBearerToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if len(authz) > 7 && authz[:7] == "Bearer " {
		return authz[7:]
	}
	return ""
}

func extractRefreshToken(r *http.Request) string {
	if token := strings.TrimSpace(r.Header.Get("X-Refresh-Token")); token != "" {
		return token
	}
	if token := strings.TrimSpace(r.Header.Get("Refresh-Token")); token != "" {
		return token
	}
	if cookie, err := r.Cookie("refresh_token"); err == nil {
		if token := strings.TrimSpace(cookie.Value); token != "" {
			return token
		}
	}
	return ""
}

func sameAccessList(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	if len(a) == 0 {
		return true
	}
	counts := make(map[string]int, len(a))
	for _, item := range a {
		counts[item]++
	}
	for _, item := range b {
		counts[item]--
		if counts[item] < 0 {
			return false
		}
	}
	for _, remaining := range counts {
		if remaining != 0 {
			return false
		}
	}
	return true
}
