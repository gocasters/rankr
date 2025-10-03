package http

import "github.com/gocasters/rankr/pkg/httpserver"

type Server struct {
	HTTPServer *httpserver.Server
	Handler    Handler
}

func New(httpServer *httpserver.Server, handler Handler) Server {
	return Server{HTTPServer: httpServer, Handler: handler}
}

func (s Server) RegisterRoutes() {
	v1 := s.HTTPServer.GetRouter().Group("/v1/userprofile")

	v1.GET("/health_check", s.healthCheck)

	v1.GET("/profile", s.profile)
}
