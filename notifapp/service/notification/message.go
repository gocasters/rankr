package notification

// NotificationType defines the type of notification.
type NotificationType string

const (
	TypeInfo    NotificationType = "info"
	TypeWarning NotificationType = "warning"
	TypeError   NotificationType = "error"
	TypeSuccess NotificationType = "success"
)

func (t NotificationType) IsValid() bool {
	switch t {
	case TypeInfo, TypeSuccess, TypeWarning, TypeError:
		return true
	}

	return false
}

// NotificationStatus defines the status of a notification.
type NotificationStatus string

const (
	StatusUnread NotificationStatus = "unread"
	StatusRead   NotificationStatus = "read"
)
