package http

import (
	"fmt"
	"strconv"

	httpserver "github.com/gocasters/rankr/pkg/httpserver"
)

type Server struct {
	HTTPServer httpserver.Server
	Handler    Handler
}

func New(server httpserver.Server, handler Handler) Server {
	return Server{
		HTTPServer: server,
		Handler:    handler,
	}
}

func (s Server) Serve() {
	s.RegisterRoutes()

	fmt.Printf("start echo server on %s\n", strconv.Itoa(s.HTTPServer.GetConfig().Port))
	if err := s.HTTPServer.Start(); err != nil {
		fmt.Println("router start error", err)
	}
}

func (s Server) RegisterRoutes() {
	// TODO- Add tracing middleware

	// create v1 group
	v1 := s.HTTPServer.GetRouter().Group("/v1")
	v1.GET("/health-check", s.healthCheck)

	// leaderboard group
	leaderboardGroup := v1.Group("/leaderboard")
	leaderboardGroup.GET("/", s.Handler.GetLeaderboards)
	//leaderboardGroup.GET("/:project", s.Handler.GetLeaderboard)

	// contributor group
	contributorGroup := v1.Group("/contributors")
	contributorGroup.GET("/:id/stats", s.Handler.GetContributorStats)
	//contributorGroup.GET("/:id/rank", s.Handler.GetContributorRank)
}
