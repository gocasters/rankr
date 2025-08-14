package handler

import (
    "net/http"
    "strconv"

    "github.com/gocasters/rankr/domain/contributor"
    "github.com/labstack/echo/v4"
)

type ContributorHandler struct {
    service *contributor.Service
}

func NewContributorHandler(service *contributor.Service) *ContributorHandler {
    return &ContributorHandler{service: service}
}

func (h *ContributorHandler) Create(c echo.Context) error {
    var req contributor.ContributorCreate
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
    }
    
    result, err := h.service.CreateContributor(c.Request().Context(), req.Username, req.Email, req.DisplayName)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusCreated, result)
}

func (h *ContributorHandler) GetByID(c echo.Context) error {
    id := c.Param("id")
    result, err := h.service.GetContributor(c.Request().Context(), id)
    if err != nil {
        return c.JSON(http.StatusNotFound, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, result)
}

func (h *ContributorHandler) Update(c echo.Context) error {
    id := c.Param("id")
    var req contributor.ContributorUpdate
    if err := c.Bind(&req); err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid input"})
    }
    
    result, err := h.service.UpdateContributor(c.Request().Context(), id, &req)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, result)
}

func (h *ContributorHandler) Delete(c echo.Context) error {
    id := c.Param("id")
    err := h.service.DeleteContributor(c.Request().Context(), id)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, map[string]string{"message": "contributor deleted"})
}

func (h *ContributorHandler) List(c echo.Context) error {
    // Parse query parameters
    limitStr := c.QueryParam("limit")
    offsetStr := c.QueryParam("offset")
    
    limit := 10 // default limit
    offset := 0 // default offset
    
    if limitStr != "" {
        if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
            limit = l
        }
    }
    
    if offsetStr != "" {
        if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
            offset = o
        }
    }
    
    result, err := h.service.ListContributors(c.Request().Context(), limit, offset)
    if err != nil {
        return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
    }
    return c.JSON(http.StatusOK, result)
}
