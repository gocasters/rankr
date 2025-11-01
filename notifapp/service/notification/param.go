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

func (c CreateRequest) createRequestMapToNotification() Notification {
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
	ID        types.ID           `json:"id"`
	UserID    types.ID           `json:"user_id"`
	Message   string             `json:"message"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
}

func (n Notification) notificationMapToCreateResponse() CreateResponse {
	return CreateResponse{
		ID:        n.ID,
		UserID:    n.UserID,
		Message:   n.Message,
		Type:      n.Type,
		Status:    n.Status,
		CreatedAt: n.CreatedAt,
	}
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
