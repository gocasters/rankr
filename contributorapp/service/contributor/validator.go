package contributor

import (
	"context"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/pkg/validator"
	types "github.com/gocasters/rankr/type"
	"mime/multipart"
	"net/http"
	"strings"

	validation "github.com/go-ozzo/ozzo-validation/v4"
)

const (
	ErrValidationRequired      = "is required"
	ErrValidationEnumPrivacy   = "must be 'real' or 'anonymous'"
	ErrValidationPositive      = "must be 'public' or 'private'"
	ErrValidationLength3To100  = "must be between 3 and 100 characters"
	ErrValidationInvalidIDType = "ID must be uint64"
)

type ValidateConfig struct {
	HttpFileType []string `koanf:"http_file_type"`
}

type ValidatorContributorRepository interface {
}

type Validator struct {
	config ValidateConfig
}

func NewValidator(cfg ValidateConfig) Validator {
	return Validator{config: cfg}
}

func (v Validator) ValidateCreateContributorRequest(ctx context.Context, req CreateContributorRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.GitHubID, validation.Required.Error(ErrValidationRequired), validation.Min(int64(1)).Error(ErrValidationPositive)),
		validation.Field(&req.GitHubUsername, validation.Required.Error(ErrValidationRequired), validation.Length(3, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.DisplayName, validation.Length(0, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.ProfileImage, validation.Length(0, 255)),
		validation.Field(&req.Bio, validation.Length(0, 500)),
		validation.Field(
			&req.PrivacyMode,
			validation.Required.Error(ErrValidationRequired),
			validation.In(PrivacyModeReal, PrivacyModeAnonymous).Error(ErrValidationEnumPrivacy),
		),
	)
}

func (v Validator) ValidateUpdateProfileRequest(ctx context.Context, req UpdateProfileRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.ID, validation.Required.Error(ErrValidationRequired), validation.By(checkID)),
		validation.Field(&req.GitHubID, validation.Required.Error(ErrValidationRequired), validation.Min(int64(1)).Error(ErrValidationPositive)),
		validation.Field(&req.GitHubUsername, validation.Required.Error(ErrValidationRequired), validation.Length(3, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.DisplayName, validation.Length(0, 100).Error(ErrValidationLength3To100)),
		validation.Field(&req.ProfileImage, validation.Length(0, 255)),
		validation.Field(&req.Bio, validation.Length(0, 500)),
		validation.Field(
			&req.PrivacyMode,
			validation.Required.Error(ErrValidationRequired),
			validation.In(PrivacyModeReal, PrivacyModeAnonymous).Error(ErrValidationEnumPrivacy),
		),
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

func (v Validator) ValidateUpsertContributorRequest(req UpsertContributorRequest) error {
	if err := validation.ValidateStruct(&req,
		validation.Field(&req.GitHubUsername, validation.Required.Error(ErrValidationRequired))); err != nil {
		return validator.NewError(err, validator.Flat, "invalid request")
	}

	return nil
}

func (v Validator) ImportJobRequestValidate(req ImportContributorRequest) error {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.File, validation.Required, validation.By(v.validateFile)),
		validation.Field(&req.FileName, validation.Required),
	); err != nil {
		return validator.NewError(err, validator.Flat, "invalid request")
	}

	return nil
}

func (v Validator) validateFile(value interface{}) error {
	file, ok := value.(multipart.File)
	if !ok {
		return fmt.Errorf("invalid file type")
	}

	if file == nil {
		return fmt.Errorf("file missing")
	}

	buffer := make([]byte, 512)
	_, err := file.Read(buffer)
	if err != nil {
		return fmt.Errorf("cannot read uploaded file")
	}

	if _, err := file.Seek(0, 0); err != nil {
		return fmt.Errorf("failed to reset file pointer")
	}

	mime := http.DetectContentType(buffer)

	for _, m := range v.config.HttpFileType {
		if strings.HasPrefix(mime, m) {
			return nil
		}
	}

	return fmt.Errorf("invalid MIME file type: %s", mime)
}
