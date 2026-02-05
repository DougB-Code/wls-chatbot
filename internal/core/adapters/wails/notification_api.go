// expose notification endpoints to the frontend via the bridge.
// internal/core/adapters/wails/notification_api.go
package wails

import "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/usecase"

// CreateNotification persists a notification payload.
func (b *Bridge) CreateNotification(payload notifications.NotificationPayload) *notifications.Notification {

	if b.notifications == nil {
		return nil
	}
	return b.notifications.CreateNotification(payload)
}

// ListNotifications returns stored notifications for the workspace.
func (b *Bridge) ListNotifications() []notifications.Notification {

	if b.notifications == nil {
		return nil
	}
	return b.notifications.ListNotifications()
}
