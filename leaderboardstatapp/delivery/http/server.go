package http

import (
	httpserver "github.com/gocasters/rankr/pkg/httpserver"
	"log/slog"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
	logger     *slog.Logger
}

func New(server httpserver.Server, handler Handler, logger *slog.Logger) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
		logger:     logger,
	}
}

func (s Server) Serve() error {
	s.RegisterRoutes()
	if err := s.HTTPServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s Server) RegisterRoutes() {
	// TODO- Add tracing middleware

	// create v1 group
	v1 := s.HTTPServer.GetRouter().Group("/v1")
	v1.GET("/health-check", s.healthCheck)

	// contributor group
	contributorGroup := v1.Group("/contributors")
	contributorGroup.GET("/:id/stats", s.Handler.GetContributorStats)
	//contributorGroup.GET("/:id/rank", s.Handler.GetContributorRank)
}
