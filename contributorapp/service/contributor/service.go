package contributor

import (
	"context"
	"github.com/gocasters/rankr/pkg/cachemanager"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/type"
)

type Repository interface {
	GetContributorByID(ctx context.Context, id types.ID) (*Contributor, error)
	CreateContributor(ctx context.Context, contributor Contributor) (*Contributor, error)
	UpdateProfileContributor(ctx context.Context, contributor Contributor) (*Contributor, error)
	GetContributorByGitHubUsername(ctx context.Context, githubUsername string) (int64, bool, error)
}

type Service struct {
	repository     Repository
	validator      Validator
	CacheManager   cachemanager.CacheManager
	forceAcceptOtp int
}

func NewService(
	repo Repository,
	cm cachemanager.CacheManager,
	validator Validator,
) Service {
	return Service{
		repository:   repo,
		validator:    validator,
		CacheManager: cm,
	}
}

func (s Service) GetProfile(ctx context.Context, id types.ID) (GetProfileResponse, error) {
	contributor, err := s.repository.GetContributorByID(ctx, id)
	if err != nil {
		logger.L().Error("contributor_get_profile", "error", err)
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

func (s Service) CreateContributor(ctx context.Context, req CreateContributorRequest) (CreateContributorResponse, error) {

	vErr := s.validator.ValidateCreateContributorRequest(ctx, req)
	if vErr != nil {
		return CreateContributorResponse{}, vErr
	}

	contributor := Contributor{
		GitHubID:       req.GitHubID,
		GitHubUsername: req.GitHubUsername,
		DisplayName:    req.DisplayName,
		ProfileImage:   req.ProfileImage,
		Bio:            req.Bio,
		PrivacyMode:    req.PrivacyMode,
	}

	createdContributor, err := s.repository.CreateContributor(ctx, contributor)
	if err != nil {
		return CreateContributorResponse{}, err
	}

	return CreateContributorResponse{
		ID: types.ID(createdContributor.ID),
	}, nil
}

func (s Service) UpdateProfile(ctx context.Context, req UpdateProfileRequest) (UpdateProfileResponse, error) {
	if err := s.validator.ValidateUpdateProfileRequest(ctx, req); err != nil {
		return UpdateProfileResponse{}, err
	}

	if _, err := s.repository.GetContributorByID(ctx, req.ID); err != nil {
		logger.L().Error("contributor_update_profile", "error", err)
		return UpdateProfileResponse{}, err
	}

	contributor := Contributor{
		ID:             int64(req.ID),
		GitHubID:       req.GitHubID,
		GitHubUsername: req.GitHubUsername,
		DisplayName:    req.DisplayName,
		ProfileImage:   req.ProfileImage,
		Bio:            req.Bio,
		PrivacyMode:    req.PrivacyMode,
	}

	resContributor, err := s.repository.UpdateProfileContributor(ctx, contributor)
	if err != nil {
		logger.L().Error("contributor_update_profile", "error", err)
		return UpdateProfileResponse{}, err
	}

	return UpdateProfileResponse{
		ID:             resContributor.ID,
		GitHubID:       resContributor.GitHubID,
		GitHubUsername: resContributor.GitHubUsername,
		DisplayName:    resContributor.DisplayName,
		ProfileImage:   resContributor.ProfileImage,
		Bio:            resContributor.Bio,
		PrivacyMode:    resContributor.PrivacyMode,
		CreatedAt:      resContributor.CreatedAt,
		UpdatedAt:      resContributor.UpdatedAt,
	}, nil
}

func (s Service) GetContributorByGithubUsername(ctx context.Context, githubUsername string) (int64, bool, error) {

	id, exists, err := s.repository.GetContributorByGitHubUsername(ctx, githubUsername)
	if err != nil {
		logger.L().Error("contributor-get-by-id", "error", err)
		return 0, false, err
	}

	if !exists {
		return 0, false, nil
	}

	return id, true, nil
}
