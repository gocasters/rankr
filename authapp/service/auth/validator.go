package auth

import (
	"fmt"

	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/go-ozzo/ozzo-validation/v4/is"
	types "github.com/gocasters/rankr/type"
)

type validator struct {
	repo Repository
}

func NewValidator(repo Repository) Validator {
	return validator{repo: repo}
}

func (v validator) ValidateCreate(req CreateGrantRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.Subject, validation.Required, validation.Length(1, 128)),
		validation.Field(&req.Object, validation.Required, validation.Length(1, 128)),
		validation.Field(&req.Action, validation.Required, validation.Length(1, 128)),
		validation.Field(&req.Field, validation.By(validateFieldValues)),
	)
}

func (v validator) ValidateUpdate(req UpdateGrantRequest) error {
	if err := v.ValidateID(types.ID(req.ID)); err != nil {
		return err
	}

	if req.Subject == "" && req.Object == "" && req.Action == "" && req.Field == nil {
		return validation.Errors{
			"update": validation.NewError("validation", "at least one field must be provided"),
		}
	}

	return validation.ValidateStruct(&req,
		validation.Field(&req.Subject, validation.When(req.Subject != "", validation.Length(1, 128))),
		validation.Field(&req.Object, validation.When(req.Object != "", validation.Length(1, 128))),
		validation.Field(&req.Action, validation.When(req.Action != "", validation.Length(1, 128))),
		validation.Field(&req.Field, validation.When(req.Field != nil, validation.By(validateFieldValues))),
	)
}

func (v validator) ValidateID(id types.ID) error {
	return validation.Validate(id,
		validation.Required,
		validation.By(func(value interface{}) error {
			if typed, ok := value.(types.ID); ok {
				if typed == 0 {
					return fmt.Errorf("id must be greater than zero")
				}
				return nil
			}
			return fmt.Errorf("invalid id type")
		}),
	)
}

func validateFieldValues(value interface{}) error {
	fields, _ := value.([]string)
	for _, f := range fields {
		if err := validation.Validate(f,
			validation.Required,
			validation.Length(1, 128),
			is.PrintableASCII,
		); err != nil {
			return err
		}
	}
	return nil
}
