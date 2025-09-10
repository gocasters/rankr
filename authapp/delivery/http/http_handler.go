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
        // Prefer Bearer token via Authorization header
        var token string
        if authz := strings.TrimSpace(c.Request().Header.Get("Authorization")); strings.HasPrefix(strings.ToLower(authz), "bearer ") {
            token = strings.TrimSpace(authz[len("Bearer "):])
        }
        // Fallback to JSON body only if header missing
        if token == "" {
            var req struct {
                Token string `json:"token"`
            }
            if err := c.Bind(&req); err != nil && c.Request().ContentLength > 0 {
                c.Response().Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
                c.Response().Header().Set("Cache-Control", "no-store")
                c.Response().Header().Set("Pragma", "no-cache")
               return c.JSON(stdhttp.StatusBadRequest, map[string]string{"error": "invalid request"})
            }
            token = strings.TrimSpace(req.Token)
        }
        if token == "" {
            c.Response().Header().Set("WWW-Authenticate", `Bearer error="invalid_request"`)
            c.Response().Header().Set("Cache-Control", "no-store")
            c.Response().Header().Set("Pragma", "no-cache")
            return c.JSON(stdhttp.StatusBadRequest, map[string]string{"error": "token is required"})
        }

    claims, err := h.authService.VerifyToken(token)

    if err != nil {
                // RFC 6750 guidance
                // RFC 6750 guidance
                c.Response().Header().Set("WWW-Authenticate", `Bearer error="invalid_token"`)
                c.Response().Header().Set("Cache-Control", "no-store")
                c.Response().Header().Set("Pragma", "no-cache")
                return c.JSON(stdhttp.StatusUnauthorized, map[string]string{"error": "invalid token"})
    }


    return c.JSON(stdhttp.StatusOK, claims)
}

