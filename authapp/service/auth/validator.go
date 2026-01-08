package auth

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
)

type Validator interface {
	ValidateLoginRequest(req LoginRequest) error
}

type validator struct {
	repo Repository
}

// NewValidator returns a Validator implementation backed by the given repository.
func NewValidator(repo Repository) Validator {
	return validator{repo: repo}
}

func (v validator) ValidateLoginRequest(req LoginRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Password, validation.Required),
		validation.Field(&req.ContributorName, validation.Required, validation.Length(3, 100)),
	)
}
