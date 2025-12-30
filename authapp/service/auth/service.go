package auth

import (
	"context"
	"errors"
	"time"

	errmsg "github.com/gocasters/rankr/pkg/err_msg"
	"github.com/gocasters/rankr/pkg/logger"
	"github.com/gocasters/rankr/pkg/role"
	"github.com/gocasters/rankr/pkg/statuscode"
	types "github.com/gocasters/rankr/type"
)

type Repository interface {
	Create(ctx context.Context, r Role) (types.ID, error)
	Get(ctx context.Context, roleID types.ID) (Role, error)
	Update(ctx context.Context, r Role) error
	Delete(ctx context.Context, roleID types.ID) error
	List(ctx context.Context, page, pageSize int) ([]Role, error)
	AddPermission(ctx context.Context, roleID, permissionID types.ID) error
	RemovePermission(ctx context.Context, roleID, permissionID types.ID) error
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
	IssueTokens(userID, role string) (string, string, error)
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
	access, refresh, err := s.tokenService.IssueTokens(contributorID.String(), assignedRole)
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

func (s Service) CreateRole(ctx context.Context, req CreateRoleRequest) (CreateRoleResponse, error) {
	if err := s.validator.ValidateCreateRoleRequest(req); err != nil {
		return CreateRoleResponse{}, err
	}

	role := Role{
		Name:        req.Name,
		Description: req.Description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	roleID, err := s.repository.Create(ctx, role)
	if err != nil {
		logger.L().Error("create_role_failed", "error", err)
		return CreateRoleResponse{}, errmsg.ErrorResponse{
			Message:         "failed to create role",
			InternalErrCode: statuscode.IntCodeUnExpected,
		}
	}

	return CreateRoleResponse{RoleID: roleID}, nil
}

func (s Service) GetRole(ctx context.Context, req GetRoleRequest) (GetRoleResponse, error) {
	if err := s.validator.ValidateGetRoleRequest(req); err != nil {
		return GetRoleResponse{}, err
	}

	role, err := s.repository.Get(ctx, req.RoleID)
	if err != nil {
		logger.L().Error("get_role_failed", "error", err)
		if errors.Is(err, statuscode.ErrRoleNotFound) {
			return GetRoleResponse{}, errmsg.ErrorResponse{
				Message:         "role not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"role_id": req.RoleID},
			}
		}
		return GetRoleResponse{}, errmsg.ErrorResponse{
			Message:         "failed to fetch role",
			InternalErrCode: statuscode.IntCodeUnExpected,
			Errors:          map[string]any{"repository_error": err.Error()},
		}
	}

	return GetRoleResponse{Role: role}, nil
}

func (s Service) UpdateRole(ctx context.Context, req UpdateRoleRequest) (UpdateRoleResponse, error) {
	if err := s.validator.ValidateUpdateRoleRequest(req); err != nil {
		return UpdateRoleResponse{}, err
	}

	role := Role{
		ID:          req.RoleID,
		Name:        req.Name,
		Description: req.Description,
		UpdatedAt:   time.Now(),
	}

	if err := s.repository.Update(ctx, role); err != nil {
		logger.L().Error("update_role_failed", "error", err)
		if errors.Is(err, statuscode.ErrRoleNotFound) {
			return UpdateRoleResponse{}, errmsg.ErrorResponse{
				Message:         "role not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"role_id": req.RoleID},
			}
		}
		return UpdateRoleResponse{}, errmsg.ErrorResponse{
			Message:         "failed to update role",
			InternalErrCode: statuscode.IntCodeUnExpected,
			Errors:          map[string]any{"repository_error": err.Error()},
		}
	}

	return UpdateRoleResponse{Success: true}, nil
}

func (s Service) DeleteRole(ctx context.Context, req DeleteRoleRequest) (DeleteRoleResponse, error) {
	if err := s.validator.ValidateDeleteRoleRequest(req); err != nil {
		return DeleteRoleResponse{}, err
	}

	if err := s.repository.Delete(ctx, req.RoleID); err != nil {
		logger.L().Error("delete_role_failed", "error", err)
		if errors.Is(err, statuscode.ErrRoleNotFound) {
			return DeleteRoleResponse{}, errmsg.ErrorResponse{
				Message:         "role not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"role_id": req.RoleID},
			}
		}
		return DeleteRoleResponse{}, errmsg.ErrorResponse{
			Message:         "failed to delete role",
			InternalErrCode: statuscode.IntCodeUnExpected,
			Errors:          map[string]any{"repository_error": err.Error()},
		}
	}

	return DeleteRoleResponse{Success: true}, nil
}

func (s Service) ListRoles(ctx context.Context, req ListRoleRequest) (ListRoleResponse, error) {
	if err := s.validator.ValidateListRoleRequest(req); err != nil {
		return ListRoleResponse{}, err
	}

	page := req.Page
	pageSize := req.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	roles, err := s.repository.List(ctx, page, pageSize)
	if err != nil {
		logger.L().Error("list_roles_failed", "error", err)
		return ListRoleResponse{}, errmsg.ErrorResponse{
			Message:         "failed to list roles",
			InternalErrCode: statuscode.IntCodeUnExpected,
			Errors:          map[string]any{"repository_error": err.Error()},
		}
	}

	return ListRoleResponse{Roles: roles}, nil
}

func (s Service) AddPermissionToRole(ctx context.Context, req AddPermissionRequest) (AddPermissionResponse, error) {
	if err := s.validator.ValidateAddPermissionRequest(req); err != nil {
		return AddPermissionResponse{}, err
	}

	if err := s.repository.AddPermission(ctx, req.RoleID, req.PermissionID); err != nil {
		logger.L().Error("add_permission_failed", "error", err)
		switch {
		case errors.Is(err, statuscode.ErrRoleNotFound):
			return AddPermissionResponse{}, errmsg.ErrorResponse{
				Message:         "role not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"role_id": req.RoleID},
			}
		case errors.Is(err, statuscode.ErrPermissionNotFound):
			return AddPermissionResponse{}, errmsg.ErrorResponse{
				Message:         "permission not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"permission_id": req.PermissionID},
			}
		default:
			return AddPermissionResponse{}, errmsg.ErrorResponse{
				Message:         "failed to add permission",
				InternalErrCode: statuscode.IntCodeUnExpected,
				Errors:          map[string]any{"repository_error": err.Error()},
			}
		}
	}

	return AddPermissionResponse{Success: true}, nil
}

func (s Service) RemovePermissionFromRole(ctx context.Context, req RemovePermissionRequest) (RemovePermissionResponse, error) {
	if err := s.validator.ValidateRemovePermissionRequest(req); err != nil {
		return RemovePermissionResponse{}, err
	}

	if err := s.repository.RemovePermission(ctx, req.RoleID, req.PermissionID); err != nil {
		logger.L().Error("remove_permission_failed", "error", err)
		switch {
		case errors.Is(err, statuscode.ErrRoleNotFound):
			return RemovePermissionResponse{}, errmsg.ErrorResponse{
				Message:         "role not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"role_id": req.RoleID},
			}
		case errors.Is(err, statuscode.ErrPermissionNotFound):
			return RemovePermissionResponse{}, errmsg.ErrorResponse{
				Message:         "permission not found",
				InternalErrCode: statuscode.IntCodeNotFound,
				Errors:          map[string]any{"permission_id": req.PermissionID},
			}
		default:
			return RemovePermissionResponse{}, errmsg.ErrorResponse{
				Message:         "failed to remove permission",
				InternalErrCode: statuscode.IntCodeUnExpected,
				Errors:          map[string]any{"repository_error": err.Error()},
			}
		}
	}

	return RemovePermissionResponse{Success: true}, nil
}
