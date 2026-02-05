// notification_repository.go declares notification persistence ports.
// internal/features/notifications/ports/notification_repository.go
package ports

import "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/domain"

// NotificationRepository persists notifications.
type NotificationRepository interface {
	Create(notification *domain.Notification) (*domain.Notification, error)
	List() ([]*domain.Notification, error)
}
