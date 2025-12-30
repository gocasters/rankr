package contributor

import (
	"context"
	"github.com/gocasters/rankr/pkg/cachemanager"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/statuscode"
	"github.com/gocasters/rankr/type"
)

type Repository interface {
	GetContributorByID(ctx context.Context, id types.ID) (*Contributor, error)
	CreateContributor(ctx context.Context, contributor Contributor) (*Contributor, error)
	UpdateProfileContributor(ctx context.Context, contributor Contributor) (*Contributor, error)
	GetContributorByGitHubUsername(ctx context.Context, githubUsername string) (int64, bool, error)
	FindByVCSUsernames(ctx context.Context, provider VcsProvider, usernames []string) ([]*Contributor, error)
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

func (s Service) Upsert(ctx context.Context, req UpsertContributorRequest) (UpsertContributorResponse, error) {
	if err := s.validator.ValidateUpsertContributorRequest(req); err != nil {
		return UpsertContributorResponse{}, err
	}

	id, exists, err := s.GetContributorByGithubUsername(ctx, req.GitHubUsername)
	if err != nil {
		return UpsertContributorResponse{}, errmsg.ErrorResponse{
			Message:         "failed to get contributor",
			Errors:          map[string]interface{}{"error": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	if !exists {
		var reqRequest Contributor

		reqRequest.GitHubUsername = req.GitHubUsername
		reqRequest.GitHubID = req.GitHubID
		reqRequest.PrivacyMode = req.PrivacyMode
		reqRequest.Bio = req.Bio
		reqRequest.ProfileImage = req.ProfileImage
		reqRequest.DisplayName = req.DisplayName
		reqRequest.Email = req.Email
		reqRequest.CreatedAt = req.CreateAt

		resCreate, err := s.repository.CreateContributor(ctx, reqRequest)
		if err != nil {
			return UpsertContributorResponse{}, errmsg.ErrorResponse{
				Message:         "failed to create contributor:" + err.Error(),
				Errors:          map[string]interface{}{"error": err.Error()},
				InternalErrCode: statuscode.IntCodeUnExpected,
			}
		}

		return UpsertContributorResponse{ID: types.ID(resCreate.ID), IsNew: true}, nil
	}

	var upRequest Contributor

	upRequest.ID = id
	upRequest.GitHubID = req.GitHubID
	upRequest.GitHubUsername = req.GitHubUsername
	upRequest.DisplayName = req.DisplayName
	upRequest.Bio = req.Bio
	upRequest.PrivacyMode = req.PrivacyMode
	upRequest.ProfileImage = req.ProfileImage
	upRequest.Email = req.Email
	upRequest.CreatedAt = req.CreateAt

	_, err = s.repository.UpdateProfileContributor(ctx, upRequest)
	if err != nil {
		return UpsertContributorResponse{}, errmsg.ErrorResponse{
			Message:         "failed to update contributor",
			Errors:          map[string]interface{}{"error": err.Error()},
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	return UpsertContributorResponse{ID: types.ID(upRequest.ID), IsNew: false}, nil
}

func (s Service) GetContributorsByVCS(ctx context.Context, req GetContributorsByVCSRequest) (GetContributorsByVCSResponse, error) {
	contributors, err := s.repository.FindByVCSUsernames(ctx, req.VcsProvider, req.Usernames)
	if err != nil {
		logger.L().Error("get_contributors_by_vcs", "error", err)
		return GetContributorsByVCSResponse{}, err
	}

	mappings := make([]ContributorMapping, 0, len(contributors))
	for _, c := range contributors {
		mappings = append(mappings, ContributorMapping{
			ContributorID: c.ID,
			VcsUsername:   c.GitHubUsername,
			VcsUserID:     c.GitHubID,
		})
	}

	return GetContributorsByVCSResponse{
		VcsProvider:  req.VcsProvider,
		Contributors: mappings,
	}, nil
}
