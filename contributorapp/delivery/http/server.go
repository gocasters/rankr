package http

import (
	"context"

	"github.com/gocasters/rankr/contributorapp/delivery/http/middleware"
	echomiddleware "github.com/gocasters/rankr/pkg/echo_middleware"
	"github.com/gocasters/rankr/pkg/httpserver"
	"log/slog"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
	logger     *slog.Logger
	Middleware middleware.Middleware
}

func New(server httpserver.Server, handler Handler, logger *slog.Logger, mid middleware.Middleware) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
		logger:     logger,
		Middleware: mid,
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
	router.Use(
		echomiddleware.RequireClaimsWithConfig(
			echomiddleware.RequireClaimsConfig{
				Skipper: echomiddleware.SkipExactPaths("/v1/health-check"),
			},
		),
	)

	// Public health check
	router.GET("/v1/health-check", s.healthCheck)

	v1 := router.Group("/v1")
	v1.GET("/profile/:id", s.Handler.getProfile)
	v1.POST("/create", s.Handler.createContributor)
	v1.PUT("/update", s.Handler.updateProfile)

	v1.POST("/jobs/upload", s.Handler.uploadFile, s.Middleware.CheckFile)
	v1.GET("/jobs/status/:job_id", s.Handler.getJobStatus)
	v1.GET("/jobs/fail_records/:job_id", s.Handler.getFailRecords)

	v1.PUT("/password", s.Handler.updatePassword)
}
