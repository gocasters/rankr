package notification

import (
	"context"
	"errors"
	"fmt"
)

var (
	ErrNotFound  = errors.New("notification not found")
	ErrForbidden = errors.New("user does not have permission to access this resource")
)

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

type Service struct {
	repo Repository
}

// NewService returns a new Service backed by the provided Repository.
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

// Get retrieves a single notification after checking for ownership.
func (s *Service) Get(ctx context.Context, req GetRequest) (GetResponse, error) {
	notification, err := s.repo.Get(ctx, req.NotificationID)
	if err != nil {
		return GetResponse{}, ErrNotFound
	}

	if notification.UserID != req.UserID {
		return GetResponse{}, ErrForbidden
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

// MarkAsRead marks a notification as read after checking for ownership.
func (s *Service) MarkAsRead(ctx context.Context, req MarkAsReadRequest) (MarkAsReadResponse, error) {
	// First, verify the notification exists and the user owns it.
	notification, err := s.repo.Get(ctx, req.NotificationID)
	if err != nil {
		return MarkAsReadResponse{}, ErrNotFound
	}
	if notification.UserID != req.UserID {
		return MarkAsReadResponse{}, ErrForbidden
	}

	updatedNotification, err := s.repo.MarkAsRead(ctx, req.NotificationID)
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
	notification, err := s.repo.Get(ctx, req.NotificationID)
	if err != nil {
		return ErrNotFound
	}
	if notification.UserID != req.UserID {
		return ErrForbidden
	}

	if err := s.repo.Delete(ctx, req.NotificationID); err != nil {
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
