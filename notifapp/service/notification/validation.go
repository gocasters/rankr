package notification

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	types "github.com/gocasters/rankr/type"
)

type Validate struct{}

func NewValidation() Validate {
	return Validate{}
}

func (v Validate) CreateNotificationValidate(req CreateRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.Message, validation.Required),
		validation.Field(&req.Type, validation.Required, validation.By(validNotifyType)))
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

func (v Validate) GetNotificationValidate(req GetRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.NotificationID, validation.Required, validation.By(validID)))
}

func (v Validate) ListNotificationsValidate(req ListRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)))
}

func (v Validate) MarkAsReadNotificationValidate(req MarkAsReadRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.NotificationID, validation.Required, validation.By(validID)))
}

func (v Validate) MarkAllAsReadNotificationValidate(req MarkAllAsReadRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)))
}

func (v Validate) DeleteNotificationValidate(req DeleteRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)),
		validation.Field(&req.NotificationID, validation.Required, validation.By(validID)),
	)
}

func (v Validate) GetUnreadCountNotificationValidate(req CountUnreadRequest) error {
	return validation.ValidateStruct(&req,
		validation.Field(&req.UserID, validation.Required, validation.By(validID)))
}
