// orchestration.go exposes notification workflows for adapters.
// internal/features/notifications/usecase/orchestration.go
package notifications

import "fmt"

// Orchestrator coordinates notification workflows for adapters.
type Orchestrator struct {
	service *Service
}

// NewOrchestrator creates a notification orchestrator.
func NewOrchestrator(service *Service) *Orchestrator {

	return &Orchestrator{service: service}
}

// CreateNotification persists a notification and returns the stored record.
func (o *Orchestrator) CreateNotification(payload NotificationPayload) (*Notification, error) {

	if o == nil || o.service == nil {
		return nil, fmt.Errorf("notifications service not configured")
	}

	return o.service.Create(payload)
}

// ListNotifications returns stored notifications.
func (o *Orchestrator) ListNotifications() []Notification {

	if o == nil || o.service == nil {
		return []Notification{}
	}

	return o.service.List()
}

// DeleteNotification removes a notification by id.
func (o *Orchestrator) DeleteNotification(id int64) bool {

	if o == nil || o.service == nil {
		return false
	}

	return o.service.Delete(id) == nil
}

// ClearNotifications removes all notifications.
func (o *Orchestrator) ClearNotifications() bool {

	if o == nil || o.service == nil {
		return false
	}

	return o.service.Clear() == nil
}
