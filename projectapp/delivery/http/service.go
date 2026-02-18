package http

import (
	"context"
	"log/slog"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/projectapp/service/project"
	"github.com/gocasters/rankr/projectapp/service/versioncontrollersystemproject"
)

type Server struct {
	HTTPServer                            *httpserver.Server
	ProjectService                        project.Service
	VersionControllerSystemProjectService versioncontrollersystemproject.Service
	Handler                               Handler
	logger                                *slog.Logger
}

func New(server *httpserver.Server, handler Handler, logger *slog.Logger, projectSvc project.Service, vcsProject versioncontrollersystemproject.Service) Server {
	return Server{
		HTTPServer:                            server,
		Handler:                               handler,
		logger:                                logger,
		ProjectService:                        projectSvc,
		VersionControllerSystemProjectService: vcsProject,
	}
}

func (s Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}
	return nil
}

func (s Server) Stop(ctx context.Context) error {
	return s.HTTPServer.Stop(ctx)
}

func (s Server) RegisterRoutes() {
	router := s.HTTPServer.GetRouter()

	v1 := router.Group("/v1")

	projectGroup := v1.Group("/projects")

	v1.GET("/health-check", s.healthCheck)

	projectGroup.GET("", s.Handler.listProjects)

	projectGroup.POST("", s.Handler.createProject)

	projectGroup.GET("/:id", s.Handler.GetProjectById)

	projectGroup.PATCH("/:id", s.Handler.UpdateProject)

	projectGroup.DELETE("/:id", s.Handler.DeleteProject)

	versionControllerSystemProjectGroup := v1.Group("/vcs-repos")
	versionControllerSystemProjectGroup.POST("/", s.Handler.CreateVersionControllerSystemProject)
	versionControllerSystemProjectGroup.GET("/:id", s.Handler.GetVersionControllerSystemProjectById)
	versionControllerSystemProjectGroup.GET("/", s.Handler.ListVersionControllerSystemProjects)
	versionControllerSystemProjectGroup.PUT("/:id", s.Handler.UpdateVersionControllerSystemProject)
	versionControllerSystemProjectGroup.DELETE("/:id", s.Handler.DeleteVersionControllerSystemProject)
}
