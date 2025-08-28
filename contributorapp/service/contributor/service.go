package contributor

import (
	"context"
	"github.com/gocasters/rankr/cachemanager"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/types"
	"log/slog"
)

type Repository interface {
	GetContributorByID(ctx context.Context, ID types.ID) (*Contributor, error)
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

func (s Service) GetProfile(ctx context.Context, req GetProfileRequest) (GetProfileResponse, error) {
	contributor, err := s.repository.GetContributorByID(ctx, req.ID)
	if err != nil {
		s.logger.Error("contributor_get_profile", "error", err)
		return GetProfileResponse{}, errmsg.ErrorResponse{
			Message: err.Error(),
			Errors: map[string]interface{}{
				"contributor_get_profile": err.Error(),
			},
		}
	}

	return GetProfileResponse{
		ID:           contributor.ID,
		GitHubID:     contributor.GitHubID,
		DisplayName:  contributor.DisplayName,
		ProfileImage: contributor.ProfileImage,
		Bio:          contributor.Bio,
		PrivacyMode:  contributor.PrivacyMode,
		CreatedAt:    contributor.CreatedAt,
	}, nil
}
