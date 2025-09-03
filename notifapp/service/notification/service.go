package notification

import (
	"context"
	"fmt"
)

// Repository defines the methods required for persisting and retrieving notifications.
type Repository interface {
	Create(ctx context.Context, notification Notification) (Notification, error)
	Get(ctx context.Context, notificationID, userID string) (Notification, error)
	List(ctx context.Context, userID string) ([]Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID string) (Notification, error)
	MarkAllAsRead(ctx context.Context, userID string) error
	Delete(ctx context.Context, notificationID, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int, error)
}

type Service struct {
	repo Repository
}

func NewService(repo Repository) *Service {
	return &Service{repo: repo}
}

// Create creates a new notification.
func (s *Service) Create(ctx context.Context, req CreateRequest) (CreateResponse, error) {
	n := new(req.UserID, req.Message, req.Type)

	createdNotification, err := s.repo.Create(ctx, *n)
	if err != nil {
		return CreateResponse{}, fmt.Errorf("failed to create notification: %w", err)
	}

	return CreateResponse{Notification: createdNotification}, nil
}

// Get retrieves a single notification.
func (s *Service) Get(ctx context.Context, req GetRequest) (GetResponse, error) {
	notification, err := s.repo.Get(ctx, req.NotificationID, req.UserID)
	if err != nil {
		return GetResponse{},
			fmt.Errorf("failed to get notification: %w", err)
	}

	return GetResponse{Notification: notification}, nil
}

func (s *Service) List(ctx context.Context, req ListRequest) (ListResponse, error) {
	notifications, err := s.repo.List(ctx, req.UserID)
	if err != nil {
		return ListResponse{}, fmt.Errorf("failed to list notifications: %w", err)
	}

	return ListResponse{Notifications: notifications}, nil
}

// MarkAsRead marks a notification as read.
func (s *Service) MarkAsRead(ctx context.Context, req MarkAsReadRequest) (MarkAsReadResponse, error) {
	updatedNotification, err := s.repo.MarkAsRead(ctx, req.NotificationID, req.UserID)
	if err != nil {
		return MarkAsReadResponse{}, fmt.Errorf("failed to mark as read: %w", err)
	}

	return MarkAsReadResponse{Notification: updatedNotification}, nil
}

// MarkAllAsRead marks all of a user's notifications as read.
func (s *Service) MarkAllAsRead(ctx context.Context, req MarkAllAsReadRequest) error {
	if err := s.repo.MarkAllAsRead(ctx, req.UserID); err != nil {
		return fmt.Errorf("failed to mark all as read: %w", err)
	}
	return nil
}

// Delete removes a notification after checking for ownership.
func (s *Service) Delete(ctx context.Context, req DeleteRequest) error {
	if err := s.repo.Delete(ctx, req.NotificationID, req.UserID); err != nil {
		return fmt.Errorf("failed to delete notification: %w", err)
	}
	return nil
}

// GetUnreadCount gets the unread count for a user.
func (s *Service) GetUnreadCount(ctx context.Context, req CountUnreadRequest) (GetUnreadCountResponse, error) {
	count, err := s.repo.GetUnreadCount(ctx, req.UserID)
	if err != nil {
		return GetUnreadCountResponse{}, fmt.Errorf("failed to get unread count: %w", err)
	}
	return GetUnreadCountResponse{Count: count}, nil
}
