package auth

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	types "github.com/gocasters/rankr/type"
)

type Validator interface {
	ValidateCreateRoleRequest(req CreateRoleRequest) error
	ValidateGetRoleRequest(req GetRoleRequest) error
	ValidateUpdateRoleRequest(req UpdateRoleRequest) error
	ValidateDeleteRoleRequest(req DeleteRoleRequest) error
	ValidateListRoleRequest(req ListRoleRequest) error
	ValidateAddPermissionRequest(req AddPermissionRequest) error
	ValidateRemovePermissionRequest(req RemovePermissionRequest) error
	ValidateID(id types.ID) error
}

type validator struct {
	repo Repository
}

// NewValidator returns a Validator implementation backed by the given repository.
func NewValidator(repo Repository) Validator {
	return validator{repo: repo}
}

func (v validator) ValidateCreateRoleRequest(req CreateRoleRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Name,
			validation.Required,
			validation.Length(1, 50),
			is.PrintableASCII,
		),
		validation.Field(&req.Description,
			validation.Length(0, 255),
			is.PrintableASCII,
		),
	)
}

func (v validator) ValidateGetRoleRequest(req GetRoleRequest) error {
	return v.ValidateID(req.RoleID)
}

func (v validator) ValidateUpdateRoleRequest(req UpdateRoleRequest) error {
	if err := v.ValidateID(req.RoleID); err != nil {
		return err
	}

	if req.Name == "" && req.Description == "" {
		return validation.Errors{
			"update": validation.NewError("validation", "at least one field must be provided"),
		}
	}

	return validation.ValidateStruct(&req,
		validation.Field(&req.Name,
			validation.When(req.Name != "", validation.Length(1, 50), is.PrintableASCII),
		),
		validation.Field(&req.Description,
			validation.When(req.Description != "", validation.Length(0, 255), is.PrintableASCII),
		),
	)
}

func (v validator) ValidateDeleteRoleRequest(req DeleteRoleRequest) error {
	return v.ValidateID(req.RoleID)
}

func (v validator) ValidateListRoleRequest(req ListRoleRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Page, validation.Min(0)),
		validation.Field(&req.PageSize, validation.Min(0), validation.Max(100)),
	)
}

func (v validator) ValidateAddPermissionRequest(req AddPermissionRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.RoleID, validation.By(v.validateIDValue)),
		validation.Field(&req.PermissionID, validation.By(v.validateIDValue)),
	)
}

func (v validator) ValidateRemovePermissionRequest(req RemovePermissionRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.RoleID, validation.By(v.validateIDValue)),
		validation.Field(&req.PermissionID, validation.By(v.validateIDValue)),
	)
}

func (v validator) ValidateID(id types.ID) error {
	return validation.Validate(id,
		validation.Required,
		validation.By(v.validateIDValue),
	)
}

func (v validator) validateIDValue(value interface{}) error {
	if typed, ok := value.(types.ID); ok {
		if typed == 0 {
			return fmt.Errorf("id must be greater than zero")
		}
		return nil
	}
	return fmt.Errorf("invalid id type")
}
