package http

import (
    "net/http"
    "github.com/gocasters/rankr/authapp/service"
    "github.com/gocasters/rankr/authapp/repository"
    "github.com/labstack/echo/v4"
    "strings"
)

type AuthHandler struct {
    authService   *service.AuthService
    roleRepo      *repository.RoleRepository
}

func NewAuthHandler(e *echo.Echo, svc *service.AuthService, repo *repository.RoleRepository) {
    h := &AuthHandler{authService: svc, roleRepo: repo}

    e.POST("/issue", h.IssueToken)
    e.POST("/verify", h.VerifyToken)
}

func (h *AuthHandler) IssueToken(c echo.Context) error {
    var req struct {
        UserID string `json:"user_id"`
    }
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
    }

    role, err := h.roleRepo.GetRoleByUserID(c.Request().Context(), req.UserID)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not fetch role"})
    }

    token, err := h.authService.IssueToken(req.UserID, role)
    if err != nil {
        return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not issue token"})
    }

    return c.JSON(http.StatusOK, map[string]string{"access_token": token})
}

func (h *AuthHandler) VerifyToken(c echo.Context) error {
    var req struct {
        Token string `json:"token"`
    }
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
    }

    // Allow Bearer token via Authorization header
    if strings.TrimSpace(req.Token) == "" {
        authz := c.Request().Header.Get("Authorization")
        if len(authz) >= 7 && strings.EqualFold(authz[0:7], "Bearer ") {
            req.Token = strings.TrimSpace(authz[7:])
        }
    }
    req.Token = strings.TrimSpace(req.Token)
    if req.Token == "" {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "token is required"})
    }

    claims, err := h.authService.VerifyToken(req.Token)

    if err != nil {
                // RFC 6750 guidance
                c.Response().Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
               return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
    }


    return c.JSON(http.StatusOK, claims)
}

