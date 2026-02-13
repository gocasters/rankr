package contributor

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/gocasters/rankr/pkg/cachemanager"
	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/gocasters/rankr/pkg/statuscode"

	"github.com/gocasters/rankr/type"

	"golang.org/x/crypto/bcrypt"
)

type Repository interface {
	GetContributorByID(ctx context.Context, id types.ID) (*Contributor, error)
	GetContributorByGitHubUsername(ctx context.Context, username string) (*Contributor, error)
	CreateContributor(ctx context.Context, contributor Contributor) (*Contributor, error)
	UpdateProfileContributor(ctx context.Context, contributor Contributor) (*Contributor, error)
	UpdatePassword(ctx context.Context, id types.ID, hashedPassword string) error

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

	if req.PrivacyMode == "" {
		req.PrivacyMode = PrivacyModeReal
	}

	hashedPassword, err := hashPassword(req.Password)
	if err != nil {
		return CreateContributorResponse{}, err
	}

	now := time.Now()

	contributor := Contributor{
		GitHubID:       req.GitHubID,
		GitHubUsername: req.GitHubUsername,
		Password:       hashedPassword,
		Role:           string(role.User),
		DisplayName:    req.DisplayName,
		ProfileImage:   req.ProfileImage,
		Bio:            req.Bio,
		PrivacyMode:    req.PrivacyMode,
		CreatedAt:      now,
		UpdatedAt:      now, //todo move to repository
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

func (s Service) Upsert(ctx context.Context, req UpsertContributorRequest) (UpsertContributorResponse, error) {
	if err := s.validator.ValidateUpsertContributorRequest(req); err != nil {
		return UpsertContributorResponse{}, err
	}

	c, err := s.repository.GetContributorByGitHubUsername(ctx, req.GitHubUsername)
	if err != nil {
		if !errors.Is(err, ErrNotFoundGithubUsername) {
			return UpsertContributorResponse{}, errmsg.ErrorResponse{
				Message:         "failed to get contributor",
				Errors:          map[string]interface{}{"error": err.Error()},
				InternalErrCode: statuscode.IntCodeUnExpected,
			}
		}
	}

	if c == nil {
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

	upRequest.ID = c.ID
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

func (s Service) GetContributorCredentials(ctx context.Context, id types.ID) (*Contributor, error) {
	if id == 0 {
		return nil, fmt.Errorf("contributor id is required")
	}

	contrib, err := s.repository.GetContributorByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return contrib, nil
}

func (s Service) UpdatePassword(ctx context.Context, req UpdatePasswordRequest) (UpdatePasswordResponse, error) {
	if err := s.validator.ValidateUpdatePasswordRequest(ctx, req); err != nil {
		return UpdatePasswordResponse{}, err
	}

	contrib, err := s.repository.GetContributorByID(ctx, req.ID)
	if err != nil {
		logger.L().Error("contributor_update_password", "error", err)
		return UpdatePasswordResponse{}, err
	}

	if !passwordMatches(contrib.Password, req.OldPassword) {
		return UpdatePasswordResponse{}, errmsg.ErrorResponse{
			Message:         "old password is incorrect",
			InternalErrCode: statuscode.IntCodeNotAuthorize,
		}
	}

	hashed, err := hashPassword(req.NewPassword)
	if err != nil {
		return UpdatePasswordResponse{}, err
	}

	if err := s.repository.UpdatePassword(ctx, req.ID, hashed); err != nil {
		logger.L().Error("contributor_update_password", "error", err)
		return UpdatePasswordResponse{}, err
	}

	return UpdatePasswordResponse{Success: true}, nil
}

func (s Service) VerifyPassword(ctx context.Context, req VerifyPasswordRequest) (VerifyPasswordResponse, error) {
	if req.Password == "" {
		return VerifyPasswordResponse{}, fmt.Errorf("password is required")
	}
	if req.ID == 0 && req.GitHubUsername == "" {
		return VerifyPasswordResponse{}, fmt.Errorf("id or github_username is required")
	}

	var (
		contrib *Contributor
		err     error
	)

	if req.ID != 0 {
		contrib, err = s.repository.GetContributorByID(ctx, req.ID)
	} else {
		contrib, err = s.repository.GetContributorByGitHubUsername(ctx, req.GitHubUsername)
	}
	if err != nil {
		return VerifyPasswordResponse{}, err
	}

	valid := passwordMatches(contrib.Password, req.Password)
	return VerifyPasswordResponse{Valid: valid, ID: types.ID(contrib.ID), Role: contrib.Role}, nil
}

func hashPassword(password string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("failed to hash password: %w", err)
	}

	return string(hashed), nil
}

func passwordMatches(hashedOrPlain, provided string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedOrPlain), []byte(provided))
	if err == nil {
		return true
	}
	if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
		return false
	}

	return hashedOrPlain == provided
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
