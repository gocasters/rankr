package notification

import "time"

// NotificationType defines the type of notification.
type NotificationType string

const (
	TypeInfo    NotificationType = "info"
	TypeWarning NotificationType = "warning"
	TypeError   NotificationType = "error"
	TypeSuccess NotificationType = "success"
)

// NotificationStatus defines the status of a notification.
type NotificationStatus string

const (
	StatusUnread NotificationStatus = "unread"
	StatusRead   NotificationStatus = "read"
)

type Notification struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	Message   string             `json:"message"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	ReadAt    *time.Time         `json:"read_at"`
}

// new creates a Notification for the given user with the provided message and notification type.
// The returned Notification has Status set to StatusUnread and ReadAt set to nil.
// ID and CreatedAt are left as zero values and must be assigned by the caller or persistence layer.
func new(userID, message string, nType NotificationType) *Notification {
	return &Notification{
		UserID:  userID,
		Message: message,
		Type:    nType,
		Status:  StatusUnread,
		ReadAt:  nil,
	}
}
