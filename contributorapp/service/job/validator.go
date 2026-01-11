package job

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gocasters/rankr/pkg/validator"
	"mime/multipart"
	"net/http"
	"strings"
)

type ValidatorJobRepo interface {
	ImportJobRequestValidate(req ImportContributorRequest) error
}

type ValidateConfig struct {
	HttpFileType []string `koanf:"http_file_type"`
}

type Validator struct {
	config ValidateConfig
}

func NewValidator(cfg ValidateConfig) ValidatorJobRepo {
	return Validator{config: cfg}
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
		if strings.HasPrefix(m, mime) || strings.HasPrefix(mime, strings.Split(m, ";")[0]) {
			return nil
		}
	}

	return fmt.Errorf("invalid MIME file type: %s", mime)
}
