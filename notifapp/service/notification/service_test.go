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

	t.Run("should return ErrNotFound for a non-existent notification", func(t *testing.T) {
		req := GetRequest{UserID: "user-A", NotificationID: "999"}
		_, err := service.Get(ctx, req)

		// The current service wraps the error, so we need to check for the wrapped error
		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to get notification")
		assert.Contains(t, err.Error(), "not found in mock")
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
		require.Error(t, getErr)
		assert.Contains(t, getErr.Error(), "failed to get notification")
		assert.Contains(t, getErr.Error(), "not found in mock")
	})
}

func TestGetUnreadCount(t *testing.T) {
	service, _ := setupTestCase()
	ctx := context.Background()

	t.Run("should return correct unread count for a user", func(t *testing.T) {
		req := CountUnreadRequest{UserID: "user-A"}
		resp, err := service.GetUnreadCount(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 1, resp.Count)
	})

	t.Run("should return zero for a user with no unread notifications", func(t *testing.T) {
		// Mark user-B's only notification as read to test this case
		_, _ = service.MarkAsRead(ctx, MarkAsReadRequest{UserID: "user-B", NotificationID: "3"})

		req := CountUnreadRequest{UserID: "user-B"}
		resp, err := service.GetUnreadCount(ctx, req)

		require.NoError(t, err)
		assert.Equal(t, 0, resp.Count)
	})

}

func TestMarkAllAsRead(t *testing.T) {
	service, mockRepo := setupTestCase()
	ctx := context.Background()

	t.Run("should mark all notifications as read for a user", func(t *testing.T) {
		// Verify initial state - user-A should have 1 unread notification
		countReq := CountUnreadRequest{UserID: "user-A"}
		countResp, err := service.GetUnreadCount(ctx, countReq)
		require.NoError(t, err)
		assert.Equal(t, 1, countResp.Count)

		// Mark all as read
		req := MarkAllAsReadRequest{UserID: "user-A"}
		err = service.MarkAllAsRead(ctx, req)
		require.NoError(t, err)

		// Verify all notifications are now read
		countResp, err = service.GetUnreadCount(ctx, countReq)
		require.NoError(t, err)
		assert.Equal(t, 0, countResp.Count)

		// Verify the specific notification that was unread is now read
		getReq := GetRequest{UserID: "user-A", NotificationID: "1"}
		getResp, err := service.Get(ctx, getReq)
		require.NoError(t, err)
		assert.Equal(t, StatusRead, getResp.Notification.Status)
		assert.NotNil(t, getResp.Notification.ReadAt)
	})

	t.Run("should only affect the specified user's notifications", func(t *testing.T) {
		// Reset test data
		mockRepo.notifications = map[string]Notification{
			"1": {ID: "1", UserID: "user-A", Message: "Hello A (unread)", Status: StatusUnread},
			"2": {ID: "2", UserID: "user-A", Message: "Hello A (read)", Status: StatusRead},
			"3": {ID: "3", UserID: "user-B", Message: "Hello B (unread)", Status: StatusUnread},
		}

		// Mark all as read for user-A
		req := MarkAllAsReadRequest{UserID: "user-A"}
		err := service.MarkAllAsRead(ctx, req)
		require.NoError(t, err)

		// Verify user-A has no unread notifications
		countReqA := CountUnreadRequest{UserID: "user-A"}
		countRespA, err := service.GetUnreadCount(ctx, countReqA)
		require.NoError(t, err)
		assert.Equal(t, 0, countRespA.Count)

		// Verify user-B still has unread notifications
		countReqB := CountUnreadRequest{UserID: "user-B"}
		countRespB, err := service.GetUnreadCount(ctx, countReqB)
		require.NoError(t, err)
		assert.Equal(t, 1, countRespB.Count)
	})

	t.Run("should handle user with no notifications", func(t *testing.T) {
		req := MarkAllAsReadRequest{UserID: "user-nonexistent"}
		err := service.MarkAllAsRead(ctx, req)
		require.NoError(t, err) // Should not error even if user has no notifications
	})

	t.Run("should handle user with no unread notifications", func(t *testing.T) {
		// Mark all as read for user-A first
		req := MarkAllAsReadRequest{UserID: "user-A"}
		err := service.MarkAllAsRead(ctx, req)
		require.NoError(t, err)

		// Try to mark all as read again - should not error
		err = service.MarkAllAsRead(ctx, req)
		require.NoError(t, err)

		// Verify count is still 0
		countReq := CountUnreadRequest{UserID: "user-A"}
		countResp, err := service.GetUnreadCount(ctx, countReq)
		require.NoError(t, err)
		assert.Equal(t, 0, countResp.Count)
	})
}
