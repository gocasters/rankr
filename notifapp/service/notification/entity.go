package notification

import "time"

type Notification struct {
	ID        string             `json:"id"`
	UserID    string             `json:"user_id"`
	Message   string             `json:"message"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	ReadAt    *time.Time         `json:"read_at"`
}

func new(userID, message string, nType NotificationType) *Notification {
	return &Notification{
		UserID:  userID,
		Message: message,
		Type:    nType,
		Status:  StatusUnread,
		ReadAt:  nil,
	}
}
