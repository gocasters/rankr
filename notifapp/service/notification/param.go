package notification

import (
	types "github.com/gocasters/rankr/type"
	"time"
)

type CreateRequest struct {
	UserID  types.ID         `json:"user_id"`
	Message string           `json:"message"`
	Type    NotificationType `json:"type"`
}

func (c CreateRequest) mapToNotification() Notification {
	return Notification{
		UserID:    c.UserID,
		Message:   c.Message,
		Type:      c.Type,
		Status:    StatusUnread,
		CreatedAt: time.Now(),
	}
}

type GetRequest struct {
	UserID         types.ID `json:"user_id"`
	NotificationID types.ID `json:"notification_id"`
}

type ListRequest struct {
	UserID types.ID `json:"user_id"`
}

type MarkAsReadRequest struct {
	UserID         types.ID `json:"user_id"`
	NotificationID types.ID `json:"notification_id"`
}

type MarkAllAsReadRequest struct {
	UserID types.ID `json:"user_id"`
}

type DeleteRequest struct {
	UserID         types.ID `json:"user_id"`
	NotificationID types.ID `json:"notification_id"`
}

type CountUnreadRequest struct {
	UserID types.ID `json:"user_id"`
}

type CreateResponse struct {
	Notification Notification `json:"notification"`
}

type GetResponse struct {
	Notification Notification `json:"notification"`
}

type ListResponse struct {
	Notifications []Notification `json:"notifications"`
}

type MarkAsReadResponse struct {
	Notification Notification `json:"notification"`
}

type GetUnreadCountResponse struct {
	Count int `json:"count"`
}
