package notification

import (
	"errors"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	types "github.com/gocasters/rankr/type"
)

var (
	ErrMsgInvalidInput = errors.New("invalid input")
)

type Validate struct{}

func NewValidation() Validate {
	return Validate{}
}

func (v Validate) CreateNotificationValidate(req CreateRequest) (map[string]string, error) {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.Message, validation.Required),
		validation.Field(&req.Type, validation.Required, validation.By(validNotifyType)),
	); err != nil {
		fieldErr := make(map[string]string)
		errV, ok := err.(validation.Errors)

		if ok {
			for key, value := range errV {
				if value != nil {
					fieldErr[key] = value.Error()
				}
			}
		}

		return fieldErr, ErrMsgInvalidInput
	}

	return nil, nil
}

func validID(value interface{}) error {
	id, _ := value.(types.ID)
	if !id.IsValid() {
		return fmt.Errorf("invalid id %d", id)
	}

	return nil
}

func validNotifyType(value interface{}) error {
	t, _ := value.(NotificationType)
	if !t.IsValid() {
		return fmt.Errorf("notification type must be 'info', 'warning', 'error' or 'success'")
	}

	return nil
}

func (v Validate) GetNotificationValidate(req GetRequest) (map[string]string, error) {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.NotificationID, validation.Required, validation.By(validID)),
	); err != nil {
		fieldErr := make(map[string]string)

		errV, ok := err.(validation.Errors)
		if ok {
			for key, value := range errV {
				if value != nil {
					fieldErr[key] = value.Error()
				}
			}
		}

		return fieldErr, ErrMsgInvalidInput
	}

	return nil, nil
}

func (v Validate) ListNotificationsValidate(req ListRequest) (map[string]string, error) {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
	); err != nil {
		fieldErr := make(map[string]string)

		errV, ok := err.(validation.Errors)
		if ok {
			for key, value := range errV {
				if value != nil {
					fieldErr[key] = value.Error()
				}
			}
		}

		return fieldErr, ErrMsgInvalidInput
	}

	return nil, nil
}

func (v Validate) MarkAsReadNotificationValidate(req MarkAsReadRequest) (map[string]string, error) {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.NotificationID, validation.Required, validation.By(validID)),
	); err != nil {
		fieldErr := make(map[string]string)

		errV, ok := err.(validation.Errors)
		if ok {
			for key, value := range errV {
				if value != nil {
					fieldErr[key] = value.Error()
				}
			}
		}

		return fieldErr, ErrMsgInvalidInput
	}

	return nil, nil
}

func (v Validate) MarkAllAsReadNotificationValidate(req MarkAllAsReadRequest) (map[string]string, error) {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
	); err != nil {
		fieldErr := make(map[string]string)

		errV, ok := err.(validation.Errors)
		if ok {
			for key, value := range errV {
				if value != nil {
					fieldErr[key] = value.Error()
				}
			}
		}

		return fieldErr, ErrMsgInvalidInput
	}

	return nil, nil
}

func (v Validate) DeleteNotificationValidate(req DeleteRequest) (map[string]string, error) {

	if err := validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.NotificationID, validation.Required, validation.By(validID)),
	); err != nil {
		fieldErr := make(map[string]string)

		errV, ok := err.(validation.Errors)
		if ok {
			for key, value := range errV {
				if value != nil {
					fieldErr[key] = value.Error()
				}
			}
		}

		return fieldErr, ErrMsgInvalidInput
	}

	return nil, nil
}
