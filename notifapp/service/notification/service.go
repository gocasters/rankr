package notification

import (
	"context"
	"github.com/gocasters/rankr/pkg/logger"
	types "github.com/gocasters/rankr/type"
)

// Repository defines the methods required for persisting and retrieving notifications.
type Repository interface {
	Get(ctx context.Context, notificationID, userID types.ID) (Notification, error)
	List(ctx context.Context, userID types.ID) ([]Notification, error)
	GetUnreadCount(ctx context.Context, userID types.ID) (int, error)
	Create(ctx context.Context, notification Notification) (Notification, error)
	MarkAsRead(ctx context.Context, notificationID, userID types.ID) (Notification, error)
	MarkAllAsRead(ctx context.Context, userID types.ID) error
	Delete(ctx context.Context, notificationID, userID types.ID) error
}

type Service struct {
	repo      Repository
	validator Validate
}

func NewService(repo Repository, validator Validate) Service {
	return Service{repo: repo, validator: validator}
}

// Create creates a new notification.
func (s Service) Create(ctx context.Context, req CreateRequest) (CreateResponse, error) {

	if err := s.validator.CreateNotificationValidate(req); err != nil {
		return CreateResponse{}, err
	}

	notify := req.createRequestMapToNotification()

	createdNotification, err := s.repo.Create(ctx, notify)
	if err != nil {
		logger.L().Error("notifapp/create-notification", "error", err)
		return CreateResponse{}, err
	}

	createResponse := createdNotification.notificationMapToCreateResponse()

	return createResponse, nil
}

// Get retrieves a single notification.
func (s Service) Get(ctx context.Context, req GetRequest) (GetResponse, error) {

	if err := s.validator.GetNotificationValidate(req); err != nil {
		return GetResponse{}, err
	}

	notification, err := s.repo.Get(ctx, req.NotificationID, req.UserID)
	if err != nil {
		logger.L().Error("notifapp/get-notification", "error", err)
		return GetResponse{}, err
	}

	return GetResponse{Notification: notification}, nil
}

func (s Service) List(ctx context.Context, req ListRequest) (ListResponse, error) {

	if err := s.validator.ListNotificationsValidate(req); err != nil {
		return ListResponse{}, err
	}

	notifications, err := s.repo.List(ctx, req.UserID)
	if err != nil {
		logger.L().Error("notifapp/list-notifications", "error", err)
		return ListResponse{}, err
	}

	return ListResponse{Notifications: notifications}, nil
}

// MarkAsRead marks a notification as read.
func (s Service) MarkAsRead(ctx context.Context, req MarkAsReadRequest) (MarkAsReadResponse, error) {

	if err := s.validator.MarkAsReadNotificationValidate(req); err != nil {
		return MarkAsReadResponse{}, err
	}

	updatedNotification, err := s.repo.MarkAsRead(ctx, req.NotificationID, req.UserID)
	if err != nil {
		logger.L().Error("notifapp/mark-as-read-notification", "error", err)
		return MarkAsReadResponse{}, err
	}

	return MarkAsReadResponse{Notification: updatedNotification}, nil
}

// MarkAllAsRead marks all of a user's notifications as read.
func (s Service) MarkAllAsRead(ctx context.Context, req MarkAllAsReadRequest) error {

	if err := s.validator.MarkAllAsReadNotificationValidate(req); err != nil {
		return err
	}

	if err := s.repo.MarkAllAsRead(ctx, req.UserID); err != nil {
		logger.L().Error("notifapp/mark-all-as-read-notification", "error", err)
		return err
	}

	return nil
}

// Delete removes a notification after checking for ownership.
func (s Service) Delete(ctx context.Context, req DeleteRequest) error {

	if err := s.validator.DeleteNotificationValidate(req); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, req.NotificationID, req.UserID); err != nil {
		logger.L().Error("notifapp/delete-notification", "error", err)
		return err
	}

	return nil
}

// GetUnreadCount gets the unread count for a user.
func (s Service) GetUnreadCount(ctx context.Context, req CountUnreadRequest) (GetUnreadCountResponse, error) {

	if err := s.validator.GetUnreadCountNotificationValidate(req); err != nil {
		return GetUnreadCountResponse{}, err
	}

	count, err := s.repo.GetUnreadCount(ctx, req.UserID)
	if err != nil {
		logger.L().Error("notifapp/get-unread-count-notification", "error", err)
		return GetUnreadCountResponse{}, err
	}

	return GetUnreadCountResponse{Count: count}, nil
}
