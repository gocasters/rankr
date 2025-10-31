package notification

import (
	types "github.com/gocasters/rankr/type"
	"time"
)

type Notification struct {
	ID        types.ID           `json:"id"`
	UserID    types.ID           `json:"user_id"`
	Message   string             `json:"message"`
	Type      NotificationType   `json:"type"`
	Status    NotificationStatus `json:"status"`
	CreatedAt time.Time          `json:"created_at"`
	ReadAt    *time.Time         `json:"read_at"`
}
