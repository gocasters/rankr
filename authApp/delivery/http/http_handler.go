package http

import (
    "net/http"
    "github.com/gocasters/rankr/authApp/service"
    "github.com/gocasters/rankr/authApp/repository"
    "github.com/labstack/echo/v4"
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

    claims, err := h.authService.VerifyToken(req.Token)
    if err != nil {
        return c.JSON(http.StatusUnauthorized, map[string]string{"error": "invalid token"})
    }

    return c.JSON(http.StatusOK, claims)
}

