package contributor

import (
	"context"
	"errors"
	types "github.com/gocasters/rankr/type"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrValidationRequired      = "is required"
	ErrValidationEnumPrivacy   = "must be 'real' or 'anonymous'"
	ErrValidationPositive      = "must be 'public' or 'private'"
	ErrValidationLength3To100  = "must be between 3 and 100 characters"
	ErrValidationLength6To72   = "must be between 6 and 72 characters"
	ErrValidationInvalidIDType = "ID must be uint64"
)

type ValidatorContributorRepository interface {
}

type Validator struct {
	repo ValidatorContributorRepository
}

func NewValidator(repo ValidatorContributorRepository) Validator {
	return Validator{repo: repo}
}

func (v Validator) ValidateCreateContributorRequest(ctx context.Context, req CreateContributorRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(
			&req.GitHubID,
			validation.When(req.GitHubID != 0, validation.Min(int64(1)).Error(ErrValidationPositive)),
		),
		validation.Field(&req.GitHubUsername, validation.Required.Error(ErrValidationRequired), validation.Length(3, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.Password, validation.Required.Error(ErrValidationRequired), validation.Length(6, 72).Error(ErrValidationLength6To72)),
		validation.Field(&req.DisplayName, validation.Length(0, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.ProfileImage, validation.Length(0, 255)),
		validation.Field(&req.Bio, validation.Length(0, 500)),
		validation.Field(
			&req.PrivacyMode,
			validation.When(req.PrivacyMode != "",
				validation.In(PrivacyModeReal, PrivacyModeAnonymous).Error(ErrValidationEnumPrivacy),
			),
		),
	)
}

func (v Validator) ValidateUpdateProfileRequest(ctx context.Context, req UpdateProfileRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validation.Required.Error(ErrValidationRequired), validation.By(checkID)),
		validation.Field(
			&req.GitHubID,
			validation.When(req.GitHubID != 0, validation.Min(int64(1)).Error(ErrValidationPositive)),
		),
		validation.Field(&req.GitHubUsername, validation.Required.Error(ErrValidationRequired), validation.Length(3, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.DisplayName, validation.Length(0, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.ProfileImage, validation.Length(0, 255)),
		validation.Field(&req.Bio, validation.Length(0, 500)),
		validation.Field(
			&req.PrivacyMode,
			validation.When(req.PrivacyMode != "",
				validation.In(PrivacyModeReal, PrivacyModeAnonymous).Error(ErrValidationEnumPrivacy),
			),
		),
	)
}

func (v Validator) ValidateUpdatePasswordRequest(_ context.Context, req UpdatePasswordRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validation.Required.Error(ErrValidationRequired), validation.By(checkID)),
		validation.Field(&req.OldPassword, validation.Required.Error(ErrValidationRequired), validation.Length(6, 72).Error(ErrValidationLength6To72)),
		validation.Field(&req.NewPassword, validation.Required.Error(ErrValidationRequired), validation.Length(6, 72).Error(ErrValidationLength6To72)),
	)
}

func checkID(value interface{}) error {
	val, ok := value.(types.ID)
	if !ok {
		return errors.New(ErrValidationInvalidIDType)
	}

	if err := val.Validate(); err != nil {
		return err
	}

	return nil
}
