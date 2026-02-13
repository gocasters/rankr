package auth

import (
	"context"

	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/gocasters/rankr/pkg/statuscode"
	types "github.com/gocasters/rankr/type"
)

type Repository interface {
	ListPermissionsByRoleName(ctx context.Context, roleName string) ([]string, error)
}

type Service struct {
	repository        Repository
	validator         Validator
	contributorClient contributorCredentialsProvider
	tokenService      tokenIssuer
}

type contributorCredentialsProvider interface {
	VerifyPassword(ctx context.Context, username string, password string) (types.ID, string, bool, error)
}

type tokenIssuer interface {
	IssueTokens(userID, role string, access []string) (string, string, error)
}

func NewService(repo Repository, validator Validator, contributorClient contributorCredentialsProvider, tokenSrv tokenIssuer) Service {
	return Service{
		repository:        repo,
		validator:         validator,
		contributorClient: contributorClient,
		tokenService:      tokenSrv,
	}
}

func (s Service) Login(ctx context.Context, req LoginRequest) (LoginResponse, error) {
	if err := s.validator.ValidateLoginRequest(req); err != nil {
		return LoginResponse{}, err
	}

	contributorID, assignedRole, valid, err := s.contributorClient.VerifyPassword(ctx, req.ContributorName, req.Password)
	if err != nil {
		logger.L().Warn("login_verify_password_failed", "error", err)
		return LoginResponse{}, errmsg.ErrorResponse{
			Message:         "username or password is incorrect",
			InternalErrCode: statuscode.IntCodeNotAuthorize,
		}
	}

	if !valid {
		return LoginResponse{}, errmsg.ErrorResponse{
			Message:         "username or password is incorrect",
			InternalErrCode: statuscode.IntCodeNotAuthorize,
		}
	}

	if assignedRole == "" {
		assignedRole = string(role.User)
	}
	if parsedRole, ok := role.Parse(assignedRole); ok {
		assignedRole = string(parsedRole)
	} else {
		assignedRole = string(role.User)
	}

	accessList, err := s.repository.ListPermissionsByRoleName(ctx, assignedRole)
	if err != nil {
		logger.L().Error("list_permissions_failed", "error", err, "role", assignedRole)
		return LoginResponse{}, errmsg.ErrorResponse{
			Message:         "failed to issue tokens",
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	access, refresh, err := s.tokenService.IssueTokens(contributorID.String(), assignedRole, accessList)
	if err != nil {
		logger.L().Error("login_issue_tokens_failed", "error", err)
		return LoginResponse{}, errmsg.ErrorResponse{
			Message:         "failed to issue tokens",
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	return LoginResponse{
		AccessToken:  access,
		RefreshToken: refresh,
	}, nil
}
