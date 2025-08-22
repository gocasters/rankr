package services

import (
	"context"
	"errors"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	interfaces2 "github.com/gocasters/rankr/projectsapp/interfaces"
	"github.com/gocasters/rankr/projectsapp/types"
)

var (
	ErrValidationRequiredLess100Char = "field is required and must be less than 100 characters"
	ErrValidationRequiredLess255Char = "field is required and must be less than 255 characters"
	ErrValidationRequiredAndNotZero  = "field is required and cannot be empty"
	ErrInvalidSlugFormat             = "slug must contain only lowercase letters, numbers, and hyphens"
	ErrInvalidURLFormat              = "invalid URL format"
	ErrSlugAlreadyExists             = errors.New("slug already exists")
	ErrRepoAlreadyExists             = errors.New("repository already exists for this project and provider")
)

type Validator struct {
	projectRepo interfaces2.IProjectRepository
	vcsRepo     interfaces2.VcsRepoRepository
}

func NewValidator(projectRepo interfaces2.IProjectRepository, vcsRepo interfaces2.VcsRepoRepository) *Validator {
	return &Validator{
		projectRepo: projectRepo,
		vcsRepo:     vcsRepo,
	}
}

func (v *Validator) ValidateCreateProject(ctx context.Context, input interfaces2.CreateProjectInput) error {
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	return validation.ValidateStruct(&input,
		validation.Field(&input.Name,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
		),
		validation.Field(&input.Slug,
			validation.Required.Error(ErrValidationRequiredLess100Char),
			validation.Length(1, 100).Error(ErrValidationRequiredLess100Char),
			validation.Match(slugPattern).Error(ErrInvalidSlugFormat),
			validation.By(func(value interface{}) error {
				return v.checkSlugUniqueness(ctx, value)
			}),
		),
		validation.Field(&input.Description,
			validation.Length(0, 1000).Error("description must be less than 1000 characters"),
		),
		validation.Field(&input.DesignReferenceURL,
			validation.When(input.DesignReferenceURL != nil && *input.DesignReferenceURL != "",
				validation.Match(urlPattern).Error(ErrInvalidURLFormat),
				validation.Length(0, 2048).Error("URL must be less than 2048 characters"),
			),
		),
	)
}

func (v *Validator) ValidateUpdateProject(ctx context.Context, input interfaces2.UpdateProjectInput) error {
	slugPattern := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	return validation.ValidateStruct(&input,
		validation.Field(&input.ID,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
		),
		validation.Field(&input.Name,
			validation.When(input.Name != nil,
				validation.Required.Error(ErrValidationRequiredAndNotZero),
				validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
			),
		),
		validation.Field(&input.Slug,
			validation.When(input.Slug != nil,
				validation.Required.Error(ErrValidationRequiredLess100Char),
				validation.Length(1, 100).Error(ErrValidationRequiredLess100Char),
				validation.Match(slugPattern).Error(ErrInvalidSlugFormat),
				validation.By(func(value interface{}) error {
					return v.checkSlugUniquenessForUpdate(ctx, input.ID, value)
				}),
			),
		),
		validation.Field(&input.Description,
			validation.When(input.Description != nil && *input.Description != nil,
				validation.Length(0, 1000).Error("description must be less than 1000 characters"),
			),
		),
		validation.Field(&input.DesignReferenceURL,
			validation.When(input.DesignReferenceURL != nil && *input.DesignReferenceURL != nil && **input.DesignReferenceURL != "",
				validation.Match(urlPattern).Error(ErrInvalidURLFormat),
				validation.Length(0, 2048).Error("URL must be less than 2048 characters"),
			),
		),
	)
}

func (v *Validator) ValidateCreateVcsRepo(ctx context.Context, input interfaces2.CreateVcsRepoInput) error {
	urlPattern := regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

	return validation.ValidateStruct(&input,
		validation.Field(&input.ProjectID,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
		),
		validation.Field(&input.Provider,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.In(types.VcsProviderGitHub, types.VcsProviderGitLab, types.VcsProviderBitbucket).Error("invalid VCS provider"),
		),
		validation.Field(&input.ProviderRepoID,
			validation.Required.Error(ErrValidationRequiredAndNotZero),
			validation.Length(1, 255).Error(ErrValidationRequiredLess255Char),
			validation.By(func(value interface{}) error {
				return v.checkVcsRepoUniqueness(ctx, input.ProjectID, input.Provider, value)
			}),
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
			validation.In(types.VcsVisibilityPublic, types.VcsVisibilityPrivate, types.VcsVisibilityInternal).Error("invalid VCS visibility"),
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

func (v *Validator) ValidateUpdateVcsRepo(input interfaces2.UpdateVcsRepoInput) error {
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
				validation.In(types.VcsVisibilityPublic, types.VcsVisibilityPrivate, types.VcsVisibilityInternal).Error("invalid VCS visibility"),
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

func (v *Validator) checkSlugUniqueness(ctx context.Context, value interface{}) error {
	slug, ok := value.(string)
	if !ok {
		return errors.New("slug must be a string")
	}

	if existing, err := v.projectRepo.FindBySlug(ctx, slug); err == nil && existing != nil {
		return ErrSlugAlreadyExists
	}

	return nil
}

func (v *Validator) checkSlugUniquenessForUpdate(ctx context.Context, projectID string, value interface{}) error {
	slug, ok := value.(string)
	if !ok {
		return errors.New("slug must be a string")
	}

	if existing, err := v.projectRepo.FindBySlug(ctx, slug); err == nil && existing != nil && existing.ID != projectID {
		return ErrSlugAlreadyExists
	}

	return nil
}

func (v *Validator) checkVcsRepoUniqueness(ctx context.Context, projectID string, provider types.VcsProvider, value interface{}) error {
	providerRepoID, ok := value.(string)
	if !ok {
		return errors.New("provider repo ID must be a string")
	}

	if existing, err := v.vcsRepo.FindByProviderID(ctx, provider, providerRepoID, projectID); err == nil && existing != nil {
		return ErrRepoAlreadyExists
	}

	return nil
}
