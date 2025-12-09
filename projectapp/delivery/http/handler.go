package http

import (
	"log/slog"

	"github.com/gocasters/rankr/projectapp/constant"
	"github.com/gocasters/rankr/projectapp/service/project"
	"github.com/gocasters/rankr/projectapp/service/versioncontrollersystemproject"
	"github.com/labstack/echo/v4"
)

type Handler struct {
	projectService                        project.Service
	versionControllerSystemProjectService versioncontrollersystemproject.Service
	logger                                *slog.Logger
}

func NewHandler(
	projectService project.Service,
	VersionControllerSystemProjectService versioncontrollersystemproject.Service,
	logger *slog.Logger,
) Handler {
	return Handler{
		projectService:                        projectService,
		versionControllerSystemProjectService: VersionControllerSystemProjectService,
		logger:                                logger,
	}
}

func (h Handler) createProject(ctx echo.Context) error {
	var input project.CreateProjectInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	response, err := h.projectService.CreateProject(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to create response", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to create response"})
	}

	return ctx.JSON(201, response)
}

func (h Handler) GetProjectById(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "response ID is required"})
	}

	response, err := h.projectService.GetProject(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get response", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to get response"})
	}

	if response == nil {
		return ctx.JSON(404, echo.Map{"error": "response not found"})
	}

	return ctx.JSON(200, response)
}

func (h Handler) listProjects(ctx echo.Context) error {
	var input project.ListProjectsInput
	if err := echo.QueryParamsBinder(ctx).
		Int32("page_size", &input.PageSize).
		Int32("offset", &input.Offset).
		BindError(); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid pagination parameters"})
	}

	projects, err := h.projectService.ListProjects(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to list projects", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to list projects"})
	}

	return ctx.JSON(200, projects)
}

func (h Handler) UpdateProject(ctx echo.Context) error {
	var input project.UpdateProjectInput
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

func (h Handler) CreateVersionControllerSystemProject(ctx echo.Context) error {
	var input versioncontrollersystemproject.CreateVersionControllerSystemProjectInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	repo, err := h.versionControllerSystemProjectService.CreateVersionControllerSystemProject(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to create VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to create VCS repo"})
	}

	return ctx.JSON(201, repo)
}

func (h Handler) GetVersionControllerSystemProjectById(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "VCS repo ID is required"})
	}

	repo, err := h.versionControllerSystemProjectService.GetVcsReposByProject(ctx.Request().Context(), id)
	if err != nil {
		h.logger.Error("failed to get VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to get VCS repo"})
	}

	return ctx.JSON(200, repo)
}

func (h Handler) ListVersionControllerSystemProjects(ctx echo.Context) error {
	repos, err := h.versionControllerSystemProjectService.ListVcsRepo(ctx.Request().Context())
	if err != nil {
		h.logger.Error("failed to list VCS repos", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to list VCS repos"})
	}

	return ctx.JSON(200, repos)
}

func (h Handler) UpdateVersionControllerSystemProject(ctx echo.Context) error {
	var input versioncontrollersystemproject.UpdateVersionControllerSystemProjectInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(400, echo.Map{"error": "invalid input"})
	}

	repo, err := h.versionControllerSystemProjectService.UpdateVcsRepo(ctx.Request().Context(), input)
	if err != nil {
		h.logger.Error("failed to update VCS repo", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to update VCS repo"})
	}

	return ctx.JSON(200, repo)
}

func (h Handler) DeleteVersionControllerSystemProject(ctx echo.Context) error {
	id := ctx.Param("id")
	if id == "" {
		return ctx.JSON(400, echo.Map{"error": "VCS repo ID is required"})
	}

	if err := h.versionControllerSystemProjectService.DeleteVcsRepo(ctx.Request().Context(), id); err != nil {
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

	repo, err := h.versionControllerSystemProjectService.GetVcsRepoByProviderID(ctx.Request().Context(), constant.VcsProvider(provider), providerRepoID, projectID)
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

	repos, err := h.versionControllerSystemProjectService.GetVcsReposByProject(ctx.Request().Context(), projectID)
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

	repos, err := h.versionControllerSystemProjectService.GetVcsReposByProject(ctx.Request().Context(), projectID)
	if err != nil {
		h.logger.Error("failed to list VCS repos by project", slog.Any("error", err))
		return ctx.JSON(500, echo.Map{"error": "failed to list VCS repos by project"})
	}

	return ctx.JSON(200, repos)
}
