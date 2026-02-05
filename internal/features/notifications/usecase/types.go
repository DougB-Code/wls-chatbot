// types.go re-exports notification domain types and ports.
// internal/features/notifications/usecase/types.go
package notifications

import (
	notificationdomain "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/domain"
	notificationports "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/ports"
)

type Notification = notificationdomain.Notification

type NotificationType = notificationdomain.Type

type Repository = notificationports.NotificationRepository

// NotificationPayload describes input needed to create a notification.
type NotificationPayload struct {
	Type    string `json:"type"`
	Title   string `json:"title"`
	Message string `json:"message"`
}

const (
	TypeInfo  = notificationdomain.TypeInfo
	TypeError = notificationdomain.TypeError
)
