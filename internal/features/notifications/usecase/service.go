// service.go manages notification lifecycle operations.
// internal/features/notifications/usecase/service.go
package notifications

import (
	"fmt"
	"strings"

	notificationdomain "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/domain"
)

// Service coordinates notification persistence.
type Service struct {
	repo Repository
}

// NewService creates a notification service with the provided repository.
func NewService(repo Repository) *Service {

	return &Service{repo: repo}
}

// Create persists a new notification and returns the stored record.
func (s *Service) Create(payload NotificationPayload) (*Notification, error) {

	if s == nil || s.repo == nil {
		return nil, fmt.Errorf("notifications: repo required")
	}

	message := strings.TrimSpace(payload.Message)
	if message == "" {
		return nil, fmt.Errorf("notifications: message required")
	}

	kind := TypeInfo
	if strings.EqualFold(strings.TrimSpace(payload.Type), string(TypeError)) {
		kind = TypeError
	}

	title := strings.TrimSpace(payload.Title)
	notification := notificationdomain.NewNotification(kind, title, message)
	return s.repo.Create(notification)
}

// List returns the stored notifications in descending order.
func (s *Service) List() []Notification {

	if s == nil || s.repo == nil {
		return []Notification{}
	}

	list, err := s.repo.List()
	if err != nil {
		return []Notification{}
	}

	result := make([]Notification, 0, len(list))
	for _, notification := range list {
		if notification == nil {
			continue
		}
		result = append(result, *notification)
	}

	return result
}

// Delete removes a notification by id.
func (s *Service) Delete(id int64) error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("notifications: repo required")
	}
	if id <= 0 {
		return fmt.Errorf("notifications: id required")
	}

	return s.repo.Delete(id)
}

// Clear removes all notifications.
func (s *Service) Clear() error {

	if s == nil || s.repo == nil {
		return fmt.Errorf("notifications: repo required")
	}

	return s.repo.Clear()
}
