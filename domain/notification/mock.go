package notification

import (
	"context"
	"strconv"
	"time"
)

type mockRepository struct {
	notifications map[string]Notification
	idCounter     int
}

// Compile-time check that mockRepository implements Repository.
var _ Repository = (*mockRepository)(nil)

func NewMockRepository() *mockRepository {
	return &mockRepository{
		notifications: make(map[string]Notification),
		idCounter:     0,
	}
}

func (m *mockRepository) Create(ctx context.Context, notification Notification) (Notification, error) {
	m.idCounter++
	id := strconv.Itoa(m.idCounter)
	notification.ID = id
	notification.CreatedAt = time.Now()
	notification.Status = StatusUnread
	notification.ReadAt = nil

	m.notifications[id] = notification

	return notification, nil
}

func (m *mockRepository) Get(ctx context.Context, notificationID string) (Notification, error) {
	if n, ok := m.notifications[notificationID]; ok {
		return n, nil
	}
	return Notification{}, ErrNotFound
}

func (m *mockRepository) List(ctx context.Context, userID string) ([]Notification, error) {
	ns := make([]Notification, 0)
	for _, n := range m.notifications {
		if n.UserID == userID {
			ns = append(ns, n)
		}
	}
	return ns, nil
}

func (m *mockRepository) MarkAsRead(ctx context.Context, notificationID string) (Notification, error) {
	if n, ok := m.notifications[notificationID]; ok {
		now := time.Now()
		n.Status = StatusRead
		n.ReadAt = &now
		m.notifications[notificationID] = n
		return n, nil
	}
	return Notification{}, ErrNotFound
}

func (m *mockRepository) MarkAllAsRead(ctx context.Context, userID string) error {
	now := time.Now()
	for id, n := range m.notifications {
		if n.UserID == userID && n.Status == StatusUnread {
			n.ReadAt = &now
			n.Status = StatusRead
			m.notifications[id] = n
		}
	}
	return nil
}

func (m *mockRepository) Delete(ctx context.Context, notificationID string) error {
	if _, ok := m.notifications[notificationID]; !ok {
		return ErrNotFound
	}
	delete(m.notifications, notificationID)
	return nil
}

func (m *mockRepository) GetUnreadCount(ctx context.Context, userID string) (int, error) {
	count := 0
	for _, n := range m.notifications {
		if n.UserID == userID && n.Status == StatusUnread {
			count++
		}
	}
	return count, nil
}
