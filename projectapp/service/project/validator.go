package project

import (
	"context"
	"errors"
	"regexp"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

var (
	ErrValidationRequiredLess100Char = "field is required and must be less than 100 characters"
	ErrValidationRequiredLess255Char = "field is required and must be less than 255 characters"
	ErrValidationRequiredAndNotZero  = "field is required and cannot be empty"
	ErrInvalidSlugFormat             = "slug must contain only lowercase letters, numbers, and hyphens"
	ErrInvalidURLFormat              = "invalid URL format"
	ErrSlugAlreadyExists             = errors.New("slug already exists")
)

type ProjectRepository interface {
	FindBySlug(ctx context.Context, slug string) (*ProjectEntity, error)
}

type Validator struct {
	projectRepo ProjectRepository
}

func NewValidator(projectRepo ProjectRepository) *Validator {
	return &Validator{
		projectRepo: projectRepo,
	}
}

func (v *Validator) ValidateCreateProject(ctx context.Context, input CreateProjectInput) error {
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

func (v *Validator) ValidateUpdateProject(ctx context.Context, input UpdateProjectInput) error {
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
			validation.When(input.Description != nil,
				validation.Length(0, 1000).Error("description must be less than 1000 characters"),
			),
		),
		validation.Field(&input.DesignReferenceURL,
			validation.When(input.DesignReferenceURL != nil && *input.DesignReferenceURL != "",
				validation.Match(urlPattern).Error(ErrInvalidURLFormat),
				validation.Length(0, 2048).Error("URL must be less than 2048 characters"),
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
