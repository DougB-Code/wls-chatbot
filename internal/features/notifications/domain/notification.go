// notification.go defines notification records.
// internal/features/notifications/domain/notification.go
package domain

import "time"

// Type describes a notification category.
type Type string

const (
	TypeInfo  Type = "info"
	TypeError Type = "error"
)

// Notification captures a stored notification entry.
type Notification struct {
	ID        int64  `json:"id"`
	Type      Type   `json:"type"`
	Title     string `json:"title"`
	Message   string `json:"message"`
	CreatedAt int64  `json:"createdAt"`
	ReadAt    *int64 `json:"readAt,omitempty"`
}

// NewNotification constructs a new notification with a timestamp.
func NewNotification(kind Type, title, message string) *Notification {

	return &Notification{
		Type:      kind,
		Title:     title,
		Message:   message,
		CreatedAt: time.Now().UnixMilli(),
	}
}

// MarkRead stamps a notification as read.
func (n *Notification) MarkRead() {

	if n == nil {
		return
	}
	readAt := time.Now().UnixMilli()
	n.ReadAt = &readAt
}
