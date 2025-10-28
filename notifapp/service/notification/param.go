package notification

type CreateRequest struct {
	UserID  int64            `json:"user_id"`
	Message string           `json:"message"`
	Type    NotificationType `json:"type"`
}

type GetRequest struct {
	UserID         int64 `json:"user_id"`
	NotificationID int64 `json:"notification_id"`
}

type ListRequest struct {
	UserID int64 `json:"user_id"`
}

type MarkAsReadRequest struct {
	UserID         int64 `json:"user_id"`
	NotificationID int64 `json:"notification_id"`
}

type MarkAllAsReadRequest struct {
	UserID int64 `json:"user_id"`
}

type DeleteRequest struct {
	UserID         int64 `json:"user_id"`
	NotificationID int64 `json:"notification_id"`
}

type CountUnreadRequest struct {
	UserID int64 `json:"user_id"`
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
