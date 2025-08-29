package http

import (
	"context"
	"log/slog"

	"github.com/gocasters/rankr/pkg/httpserver"
	"github.com/gocasters/rankr/projectsapp/services"
)

type Server struct {
	HTTPServer     *httpserver.Server
	ProjectService services.ProjectService
	VcsRepoService services.VcsRepoService
	Handler        Handler
	logger         *slog.Logger
}

func New(server *httpserver.Server, handler Handler, logger *slog.Logger, project services.ProjectService, vcsProject services.VcsRepoService) Server {
	return Server{
		HTTPServer:     server,
		Handler:        handler,
		logger:         logger,
		ProjectService: project,
		VcsRepoService: vcsProject,
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
	return s.HTTPServer.StopWithTimeout()
}

func (s Server) RegisterRoutes() {
	v1 := s.HTTPServer.GetRouter().Group("/v1")
	v1.GET("/health-check", s.healthCheck)

	projectGroup := v1.Group("/projects")
	projectGroup.POST("/", s.Handler.CreateProject)
	projectGroup.GET("/:id", s.Handler.GetProjectById)
	projectGroup.GET("/", s.Handler.ListProjects)
	projectGroup.PUT("/:id", s.Handler.UpdateProject)
	projectGroup.DELETE("/:id", s.Handler.DeleteProject)

	vcsRepoGroup := v1.Group("/vcs-repos")
	vcsRepoGroup.POST("/", s.Handler.CreateVcsRepo)
	vcsRepoGroup.GET("/:id", s.Handler.GetVcsRepoById)
	vcsRepoGroup.GET("/", s.Handler.ListVcsRepos)
	vcsRepoGroup.PUT("/:id", s.Handler.UpdateVcsRepo)
	vcsRepoGroup.DELETE("/:id", s.Handler.DeleteVcsRepo)
}
