package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/notifapp/service/notification"
	"time"
)

func (r Repository) Create(ctx context.Context, notify notification.Notification) (notification.Notification, error) {

	query := `INSERT INTO notifications (user_id, message, type, status, created_at)
			  VALUES ($1,$2,$3,$4,$5)
			  RETURNING id;`
	var id int64

	err := r.db.Pool.QueryRow(ctx, query,
		notify.UserID,
		notify.Message,
		notify.Type,
		notify.Status,
		notify.CreatedAt,
	).Scan(&id)

	if err != nil {
		return notification.Notification{}, fmt.Errorf("failed create notification: %w", err)
	}

	if notify.CreatedAt.IsZero() {
		notify.CreatedAt = time.Now()
	}
	notify.ID = id

	return notify, nil
}

func (r Repository) MarkAsRead(ctx context.Context, notificationID, userID int64) (notification.Notification, error) {

	query := `
	UPDATE notifications
	SET status=$1, read_at=$2
	WHERE id=$3 AND user_id=$4 AND deleted_at IS NULL 
	 RETURNING id, user_id, message, type, status, created_at, read_at;
`
	var notify notification.Notification

	err := r.db.Pool.QueryRow(ctx, query,
		notification.StatusRead,
		time.Now(),
		notificationID,
		userID,
	).Scan(
		&notify.ID,
		&notify.UserID,
		&notify.Message,
		&notify.Type,
		&notify.Status,
		&notify.CreatedAt,
		&notify.ReadAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return notification.Notification{}, ErrNotificationNotFound
		}

		return notification.Notification{}, fmt.Errorf("failed to notification mark as read: %w", err)
	}

	return notify, nil
}

func (r Repository) MarkAllAsRead(ctx context.Context, userID int64) error {

	query := `
	UPDATE notifications
	SET status=$1, read_at=$2
	WHERE user_id=$3 AND status=$4 AND deleted_at IS NULL ;`

	_, err := r.db.Pool.Exec(ctx, query,
		notification.StatusRead,
		time.Now(),
		userID,
		notification.StatusUnread,
	)

	if err != nil {
		return fmt.Errorf("failed to mark all notifications as read for user %d: %w", userID, err)
	}

	return nil
}

func (r Repository) Delete(ctx context.Context, notificationID, userID int64) error {

	query := `
	UPDATE notifications
	SET deleted_at=$1
	WHERE user_id=$2 AND id=$3 AND deleted_at IS NULL ;`

	cmdTag, err := r.db.Pool.Exec(ctx, query,
		time.Now(),
		userID,
		notificationID,
	)
	if err != nil {
		return fmt.Errorf("failed to soft delete notification: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return ErrNotificationNotFound
	}

	return nil
}
