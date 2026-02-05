// orchestration.go exposes notification workflows for adapters.
// internal/features/notifications/usecase/orchestration.go
package notifications

// Orchestrator coordinates notification workflows for adapters.
type Orchestrator struct {
	service *Service
}

// NewOrchestrator creates a notification orchestrator.
func NewOrchestrator(service *Service) *Orchestrator {

	return &Orchestrator{service: service}
}

// CreateNotification persists a notification and returns the stored record.
func (o *Orchestrator) CreateNotification(payload NotificationPayload) *Notification {

	if o == nil || o.service == nil {
		return nil
	}

	notification, err := o.service.Create(payload)
	if err != nil {
		return nil
	}
	return notification
}

// ListNotifications returns stored notifications.
func (o *Orchestrator) ListNotifications() []Notification {

	if o == nil || o.service == nil {
		return []Notification{}
	}

	return o.service.List()
}
