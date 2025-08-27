package contributor

import (
	"github.com/gocasters/rankr/cachemanager"
	"log/slog"
)

type Repository interface {
}

type Service struct {
	repository     Repository
	validator      Validator
	logger         *slog.Logger
	CacheManager   cachemanager.CacheManager
	forceAcceptOtp int
}

func NewService(
	repo Repository,
	cm cachemanager.CacheManager,
	validator Validator,
	logger *slog.Logger,
) Service {
	return Service{
		repository:   repo,
		validator:    validator,
		logger:       logger,
		CacheManager: cm,
	}
}
