package http

import (
	"github.com/gocasters/rankr/contributorapp/service/contributor"
	"log/slog"
)

type Handler struct {
	ContributorService contributor.Service
	Logger             *slog.Logger
}

func NewHandler(contributorSrv contributor.Service, logger *slog.Logger) Handler {
	return Handler{
		ContributorService: contributorSrv,
		Logger:             logger,
	}
}
