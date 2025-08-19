package notification

import "context"

// Repository defines the methods required for persisting and retrieving notifications.
type Repository interface {
	Create(ctx context.Context, notification Notification) (Notification, error)
	Get(ctx context.Context, notificationID string) (Notification, error)
	List(ctx context.Context, userID string) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID string) (Notification, error)
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, notificationID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}
