package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/gocasters/rankr/notifapp/service/notification"
)

func (r Repository) Get(ctx context.Context, notificationID, userID int64) (notification.Notification, error) {

	query := `SELECT id, user_id, message, type, status, created_at, read_at
              FROM notifications 
              WHERE id=$1 AND user_id=$2 AND deleted_at IS NULL;`

	var notify notification.Notification
	var readAt sql.NullTime
	err := r.db.Pool.QueryRow(ctx, query, notificationID, userID).Scan(
		&notify.ID,
		&notify.UserID,
		&notify.Message,
		&notify.Type,
		&notify.Status,
		&notify.CreatedAt,
		&readAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return notification.Notification{}, ErrNotificationNotFound
		}

		return notification.Notification{}, fmt.Errorf("failed to get notification: %w", err)
	}

	if readAt.Valid {
		notify.ReadAt = &readAt.Time
	}

	return notify, nil
}

func (r Repository) List(ctx context.Context, userID int64) ([]notification.Notification, error) {

	query := `SELECT id, user_id, message, type, status, created_at, read_at
             FROM notifications 
             WHERE user_id=$1 AND deleted_at IS NULL 
             ORDER BY created_at DESC;`

	var notifies []notification.Notification

	rows, err := r.db.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query notifications: %w", err)
	}

	defer rows.Close()

	for rows.Next() {
		var notify notification.Notification
		var readAt sql.NullTime
		if err := rows.Scan(
			&notify.ID,
			&notify.UserID,
			&notify.Message,
			&notify.Type,
			&notify.Status,
			&notify.CreatedAt,
			&readAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan row in list of notification: %w", err)
		}

		if readAt.Valid {
			notify.ReadAt = &readAt.Time
		}

		notifies = append(notifies, notify)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error to reading notifications: %w", err)
	}

	if len(notifies) == 0 {
		return nil, ErrNotificationNotFound
	}

	return notifies, nil
}

func (r Repository) GetUnreadCount(ctx context.Context, userID int64) (int, error) {

	query := `SELECT COUNT(*)
			  FROM notifications 
			  WHERE user_id=$1 AND status=$2 AND deleted_at IS NULL;`

	var count int
	err := r.db.Pool.QueryRow(ctx, query, userID, notification.StatusUnread).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}
