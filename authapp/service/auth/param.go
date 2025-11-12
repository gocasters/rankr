package auth

import (
	types "github.com/gocasters/rankr/type"
)

type CreateRoleRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateRoleResponse struct {
	RoleID types.ID `json:"role_id"`
}

type GetRoleRequest struct {
	RoleID types.ID `json:"role_id"`
}

type GetRoleResponse struct {
	Role Role `json:"role"`
}

type UpdateRoleRequest struct {
	RoleID      types.ID `json:"role_id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
}

type UpdateRoleResponse struct {
	Success bool `json:"success"`
}

type DeleteRoleRequest struct {
	RoleID types.ID `json:"role_id"`
}

type DeleteRoleResponse struct {
	Success bool `json:"success"`
}

type ListRoleRequest struct {
	Page     int `json:"page"`
	PageSize int `json:"page_size"`
}

type ListRoleResponse struct {
	Roles []Role `json:"roles"`
}

type AddPermissionRequest struct {
	RoleID       types.ID `json:"role_id"`
	PermissionID types.ID `json:"permission_id"`
}

type AddPermissionResponse struct {
	Success bool `json:"success"`
}

type RemovePermissionRequest struct {
	RoleID       types.ID `json:"role_id"`
	PermissionID types.ID `json:"permission_id"`
}

type RemovePermissionResponse struct {
	Success bool `json:"success"`
}
