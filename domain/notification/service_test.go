package notification

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestCase() (*Service, *mockRepository) {
	mockRepo := NewMockRepository()
	// Pre-populate with some data.
	mockRepo.notifications = map[string]Notification{
		"1": {ID: "1", UserID: "user-A", Message: "Hello A (unread)", Status: StatusUnread},
		"2": {ID: "2", UserID: "user-A", Message: "Hello A (read)", Status: StatusRead},
		"3": {ID: "3", UserID: "user-B", Message: "Hello B (unread)", Status: StatusUnread},
	}
	mockRepo.idCounter = 3

	service := NewService(mockRepo)
	return service, mockRepo
}

func TestCreateNotification(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()
	req := CreateRequest{
		UserID:  "user-C",
		Message: "A new message",
		Type:    TypeInfo,
	}

	resp, err := service.Create(ctx, req)

	require.NoError(t, err)
	assert.Equal(t, "4", resp.Notification.ID)
	assert.Equal(t, "user-C", resp.Notification.UserID)
	assert.Equal(t, StatusUnread, resp.Notification.Status)
	assert.False(t, resp.Notification.CreatedAt.IsZero(), "CreatedAt should be set")
}

func TestGetNotification(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	t.Run("should get own notification successfully", func(t *testing.T) {
		req := GetRequest{UserID: "user-A", NotificationID: "1"}
		resp, err := service.Get(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, "1", resp.Notification.ID)
		assert.Equal(t, "user-A", resp.Notification.UserID)
	})

	t.Run("should return ErrForbidden when getting another user's notification", func(t *testing.T) {
		req := GetRequest{UserID: "user-A", NotificationID: "3"} // User-A tries to get User-B's notification
		_, err := service.Get(ctx, req)

		// Assert that the specific error is ErrForbidden
		require.Error(t, err)
		assert.Equal(t, ErrForbidden, err)
	})

	t.Run("should return ErrNotFound for a non-existent notification", func(t *testing.T) {
		req := GetRequest{UserID: "user-A", NotificationID: "999"}
		_, err := service.Get(ctx, req)

		// Assert that the specific error is ErrNotFound
		require.Error(t, err)
		assert.Equal(t, ErrNotFound, err)
	})
}

func TestListNotifications(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	t.Run("should list notifications for the correct user", func(t *testing.T) {
		req := ListRequest{UserID: "user-A"}
		resp, err := service.List(ctx, req)

		require.NoError(t, err)
		assert.Len(t, resp.Notifications, 2)
	})

	t.Run("should return an empty slice for a user with no notifications", func(t *testing.T) {
		req := ListRequest{UserID: "user-D"} // User with no notifications
		resp, err := service.List(ctx, req)

		require.NoError(t, err)
		assert.Len(t, resp.Notifications, 0)
	})
}

func TestMarkAsRead(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	t.Run("should mark own notification as read", func(t *testing.T) {
		req := MarkAsReadRequest{UserID: "user-A", NotificationID: "1"}
		resp, err := service.MarkAsRead(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, StatusRead, resp.Notification.Status)
		assert.NotNil(t, resp.Notification.ReadAt)
	})

	t.Run("should return ErrForbidden when marking another user's notification", func(t *testing.T) {
		req := MarkAsReadRequest{UserID: "user-A", NotificationID: "3"}
		_, err := service.MarkAsRead(ctx, req)

		require.Error(t, err)
		assert.Equal(t, ErrForbidden, err)
	})
}

func TestDelete(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	t.Run("should delete own notification", func(t *testing.T) {
		req := DeleteRequest{UserID: "user-A", NotificationID: "1"}
		err := service.Delete(ctx, req)

		require.NoError(t, err)

		// Verify it was actually deleted by trying to get it again
		_, getErr := service.Get(ctx, GetRequest{UserID: "user-A", NotificationID: "1"})
		assert.Equal(t, ErrNotFound, getErr)
	})

	t.Run("should return ErrForbidden when deleting another user's notification", func(t *testing.T) {
		req := DeleteRequest{UserID: "user-A", NotificationID: "3"}
		err := service.Delete(ctx, req)

		require.Error(t, err)
		assert.Equal(t, ErrForbidden, err)
	})
}

func TestGetUnreadCount(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	t.Run("should return correct unread count for a user", func(t *testing.T) {
		req := GetUnreadCountRequest{UserID: "user-A"}
		resp, err := service.GetUnreadCount(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Count)
	})

	t.Run("should return zero for a user with no unread notifications", func(t *testing.T) {
		_, err := service.MarkAsRead(ctx, MarkAsReadRequest{UserID: "user-B", NotificationID: "3"})
		require.NoError(t, err)

		req := GetUnreadCountRequest{UserID: "user-B"}
		resp, err := service.GetUnreadCount(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 0, resp.Count)
	})
}

func TestMarkAllAsRead(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	// user-A initially has 1 unread and 1 read
	countResp, err := service.GetUnreadCount(ctx, GetUnreadCountRequest{UserID: "user-A"})
	require.NoError(t, err)
	require.Equal(t, 1, countResp.Count)

	err = service.MarkAllAsRead(ctx, MarkAllAsReadRequest{UserID: "user-A"})
	require.NoError(t, err)

	// Validate all are read now
	countResp, err = service.GetUnreadCount(ctx, GetUnreadCountRequest{UserID: "user-A"})
	require.NoError(t, err)
	require.Equal(t, 0, countResp.Count)

	// Ensure user-B remains unaffected
	countRespB, err := service.GetUnreadCount(ctx, GetUnreadCountRequest{UserID: "user-B"})
	require.NoError(t, err)
	require.Equal(t, 1, countRespB.Count)
}
