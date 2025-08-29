package http

import (
	"log/slog"

	interfaces2 "github.com/gocasters/rankr/projectsapp/interfaces"
	"github.com/gocasters/rankr/projectsapp/services"
	"github.com/gocasters/rankr/projectsapp/types"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	projectService services.ProjectService
	vcsRepoService services.VcsRepoService
	logger         *slog.Logger
}

func NewHandler(
	projectService services.ProjectService,
	vcsRepoService services.VcsRepoService,
	logger *slog.Logger,
) Handler {
	return Handler{
		projectService: projectService,
		vcsRepoService: vcsRepoService,
		logger:         logger,
	}
}

func (h Handler) CreateProject(ctx echo.Context) error {
	var input interfaces2.CreateProjectInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	project, err := h.projectService.CreateProject(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to create project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to create project"})
	}

	return ctx.JSON(201, project)
}

func (h Handler) GetProjectById(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "project ID is required"})
	}

	project, err := h.projectService.GetProject(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to get project"})
	}

	if project == nil {
		return ctx.JSON(404, echo.Map{"error": "project not found"})
	}

	return ctx.JSON(200, project)
}

func (h Handler) ListProjects(ctx echo.Context) error {
	projects, err := h.projectService.ListProjects(ctx.Request().Context())
	if err != nil {
		h.logger.Error("failed to list projects", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to list projects"})
	}

	return ctx.JSON(200, projects)
}

func (h Handler) UpdateProject(ctx echo.Context) error {
	var input interfaces2.UpdateProjectInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	project, err := h.projectService.UpdateProject(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to update project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to update project"})
	}

	return ctx.JSON(200, project)
}

func (h Handler) DeleteProject(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "project ID is required"})
	}

	if err := h.projectService.DeleteProject(ctx.Request().Context(), id); err != nil {
		h.logger.Error("failed to delete project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to delete project"})
	}

	return ctx.NoContent(204)
}

func (h Handler) CreateVcsRepo(ctx echo.Context) error {
	var input interfaces2.CreateVcsRepoInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	repo, err := h.vcsRepoService.CreateVcsRepo(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to create VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to create VCS repo"})
	}

	return ctx.JSON(201, repo)
}

func (h Handler) GetVcsRepoById(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "VCS repo ID is required"})
	}

	repo, err := h.vcsRepoService.GetVcsRepo(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to get VCS repo"})
	}

	if repo == nil {
		return ctx.JSON(404, echo.Map{"error": "VCS repo not found"})
	}

	return ctx.JSON(200, repo)
}

func (h Handler) ListVcsRepos(ctx echo.Context) error {
	repos, err := h.vcsRepoService.ListVcsRepo(ctx.Request().Context())
	if err != nil {
		h.logger.Error("failed to list VCS repos", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to list VCS repos"})
	}

	return ctx.JSON(200, repos)
}

func (h Handler) UpdateVcsRepo(ctx echo.Context) error {
	var input interfaces2.UpdateVcsRepoInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	repo, err := h.vcsRepoService.UpdateVcsRepo(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to update VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to update VCS repo"})
	}

	return ctx.JSON(200, repo)
}

func (h Handler) DeleteVcsRepo(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "VCS repo ID is required"})
	}

	if err := h.vcsRepoService.DeleteVcsRepo(ctx.Request().Context(), id); err != nil {
		h.logger.Error("failed to delete VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to delete VCS repo"})
	}

	return ctx.NoContent(204)
}

func (h Handler) GetVcsRepoByProviderID(ctx echo.Context) error {
	provider := ctx.Param("provider")
	providerRepoID := ctx.Param("providerRepoID")
	projectID := ctx.Param("projectID")

	if provider == "" || providerRepoID == "" || projectID == "" {
		return ctx.JSON(400, echo.Map{"error": "provider, providerRepoID, and projectID are required"})
	}

	repo, err := h.vcsRepoService.GetVcsRepoByProviderID(ctx.Request().Context(), types.VcsProvider(provider), providerRepoID, projectID)
	if err != nil {
		h.logger.Error("failed to get VCS repo by provider ID", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to get VCS repo by provider ID"})
	}

	if repo == nil {
		return ctx.JSON(404, echo.Map{"error": "VCS repo not found"})
	}

	return ctx.JSON(200, repo)
}

func (h Handler) GetVcsReposByProject(ctx echo.Context) error {
	projectID := ctx.Param("projectID")
	if projectID == "" {
		return ctx.JSON(400, echo.Map{"error": "project ID is required"})
	}

	repos, err := h.vcsRepoService.GetVcsReposByProject(ctx.Request().Context(), projectID)
	if err != nil {
		h.logger.Error("failed to get VCS repos by project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to get VCS repos by project"})
	}

	return ctx.JSON(200, repos)
}

func (h Handler) ListVcsReposByProject(ctx echo.Context) error {
	projectID := ctx.Param("projectID")
	if projectID == "" {
		return ctx.JSON(400, echo.Map{"error": "project ID is required"})
	}

	repos, err := h.vcsRepoService.GetVcsReposByProject(ctx.Request().Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list VCS repos by project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to list VCS repos by project"})
	}

	return ctx.JSON(200, repos)
}
