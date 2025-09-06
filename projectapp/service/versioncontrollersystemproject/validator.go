package versioncontrollersystemproject

import (
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gocasters/rankr/projectapp/constant"
)

var (
	ErrValidationRequiredLess255Char = "field is required and must be less than 255 characters"
	ErrValidationRequiredAndNotZero  = "field is required and cannot be empty"
	ErrInvalidURLFormat              = "invalid URL format"
)

type Validator struct {
}

func NewValidator() *Validator {
	return &Validator{}
}

func (v *Validator) ValidateCreateVersionControllerSystemProject(input CreateVersionControllerSystemProjectInput) error {
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	return validation.ValidateStruct(&input,
		validation.Field(&input.ProjectID,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
		),
		validation.Field(&input.Provider,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.In(constant.VcsProviderGitHub, constant.VcsProviderGitLab, constant.VcsProviderBitbucket).Error("invalid VCS provider"),
		),
		validation.Field(&input.ProviderRepoID,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
		),
		validation.Field(&input.Owner,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
		),
		validation.Field(&input.Name,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
		),
		validation.Field(&input.RemoteURL,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.Match(urlPattern).Error(ErrInvalidURLFormat),
			validation.Length(1, 2048).Error("remote URL must be less than 2048 characters"),
		),
		validation.Field(&input.Visibility,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.In(constant.VcsVisibilityPublic, constant.VcsVisibilityPrivate, constant.VcsVisibilityInternal).Error("invalid VCS visibility"),
		),
		validation.Field(&input.DefaultBranch,
			validation.When(input.DefaultBranch != nil,
				validation.Length(1, 255).Error("default branch must be less than 255 characters"),
			),
		),
		validation.Field(&input.InstallationID,
			validation.When(input.InstallationID != nil,
				validation.Length(1, 255).Error("installation ID must be less than 255 characters"),
			),
		),
	)
}

func (v *Validator) ValidateUpdateVcsRepo(input UpdateVersionControllerSystemProjectInput) error {
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	return validation.ValidateStruct(&input,
		validation.Field(&input.ID,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
		),
		validation.Field(&input.Owner,
			validation.When(input.Owner != nil,
				validation.Required.Error(ErrValidationRequiredAndNotZero),
				validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
			),
		),
		validation.Field(&input.Name,
			validation.When(input.Name != nil,
				validation.Required.Error(ErrValidationRequiredAndNotZero),
				validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
			),
		),
		validation.Field(&input.RemoteURL,
			validation.When(input.RemoteURL != nil,
				validation.Required.Error(ErrValidationRequiredAndNotZero),
				validation.Match(urlPattern).Error(ErrInvalidURLFormat),
				validation.Length(1, 2048).Error("remote URL must be less than 2048 characters"),
			),
		),
		validation.Field(&input.Visibility,
			validation.When(input.Visibility != nil,
				validation.Required.Error(ErrValidationRequiredAndNotZero),
				validation.In(constant.VcsVisibilityPublic, constant.VcsVisibilityPrivate, constant.VcsVisibilityInternal).Error("invalid VCS visibility"),
			),
		),
		validation.Field(&input.DefaultBranch,
			validation.When(input.DefaultBranch != nil && *input.DefaultBranch != nil,
				validation.Length(1, 255).Error("default branch must be less than 255 characters"),
			),
		),
		validation.Field(&input.InstallationID,
			validation.When(input.InstallationID != nil && *input.InstallationID != nil,
				validation.Length(1, 255).Error("installation ID must be less than 255 characters"),
			),
		),
	)
}
