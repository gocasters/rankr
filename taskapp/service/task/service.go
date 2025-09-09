package task

import (
	"log/slog"

	"github.com/gocasters/rankr/cachemanager"
)

type Repository interface {
}

type Service struct {
	repository     Repository
	validator      Validator
	logger         *slog.Logger
	forceAcceptOtp int
}

func NewService(
	repo Repository,
	validator Validator,
	logger *slog.Logger,
) Service {
	return Service{
		repository: repo,
		validator:  validator,
		logger:     logger,
	}
}
