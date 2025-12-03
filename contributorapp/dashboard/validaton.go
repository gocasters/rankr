package dashboard

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/gocasters/rankr/pkg/validator"
	"mime/multipart"
	"net/http"
	"strings"
)

type Config struct {
	HttpFileType []string `koanf:"http_file_type"`
}

type Validate struct {
	config Config
}

func NewValidate(cfg Config) Validate {
	return Validate{config: cfg}
}

func (v Validate) ImportJobRequestValidate(req ImportJobRequest) error {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.File, validation.Required, validation.By(v.validateFile)),
		validation.Field(&req.FileName, validation.Required),
	); err != nil {
		return validator.NewError(err, validator.Flat, "invalid request")
	}

	return nil
}

func (v Validate) validateFile(value interface{}) error {
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
