package notification

type CreateRequest struct {
	UserID  string           `json:"user_id"`
	Message string           `json:"message"`
	Type    NotificationType `json:"type"`
}

type GetRequest struct {
	UserID         string `json:"user_id"`
	NotificationID string `json:"notification_id"`
}

type ListRequest struct {
	UserID string `json:"user_id"`
}

type MarkAsReadRequest struct {
	UserID         string `json:"user_id"`
	NotificationID string `json:"notification_id"`
}

type MarkAllAsReadRequest struct {
	UserID string `json:"user_id"`
}

type DeleteRequest struct {
	UserID         string `json:"user_id"`
	NotificationID string `json:"notification_id"`
}

type CountUnreadRequest struct {
	UserID string `json:"user_id"`
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
