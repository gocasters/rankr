package auth

import (
	"time"

	types "github.com/gocasters/rankr/type"
)

// Role describes an access role alongside its permissions.
type Role struct {
	ID          types.ID     `json:"id"`
	Name        string       `json:"name"`
	Description string       `json:"description"`
	Permissions []Permission `json:"permissions,omitempty"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedAt   time.Time    `json:"updated_at"`
}

// Permission defines a particular capability granted to a role.
type Permission struct {
	ID          types.ID  `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RolePermission expresses the relation between a role and a permission.
type RolePermission struct {
	RoleID       types.ID  `json:"role_id"`
	PermissionID types.ID  `json:"permission_id"`
	CreatedAt    time.Time `json:"created_at"`
}
