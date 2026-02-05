// sqlite.go persists notifications in SQLite.
// internal/features/notifications/adapters/notificationrepo/sqlite.go
package notificationrepo

import (
	"database/sql"
	"fmt"

	notificationdomain "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/domain"
	notificationports "github.com/MadeByDoug/wls-chatbot/internal/features/notifications/ports"
)

const notificationSchema = `
CREATE TABLE IF NOT EXISTS notifications (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	type TEXT NOT NULL,
	title TEXT NOT NULL,
	message TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	read_at INTEGER
);

CREATE INDEX IF NOT EXISTS idx_notifications_created_at
ON notifications (created_at DESC);
`

// Repository stores notifications in SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed notification repository.
func NewRepository(db *sql.DB) (*Repository, error) {

	if db == nil {
		return nil, fmt.Errorf("notification repo: db required")
	}

	if _, err := db.Exec(notificationSchema); err != nil {
		return nil, fmt.Errorf("notification repo: ensure schema: %w", err)
	}

	return &Repository{db: db}, nil
}

var _ notificationports.NotificationRepository = (*Repository)(nil)

// Create stores a notification record.
func (r *Repository) Create(notification *notificationdomain.Notification) (*notificationdomain.Notification, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("notification repo: db required")
	}
	if notification == nil {
		return nil, fmt.Errorf("notification repo: notification required")
	}

	var readAt interface{} = nil
	if notification.ReadAt != nil {
		readAt = *notification.ReadAt
	}

	result, err := r.db.Exec(
		`INSERT INTO notifications (type, title, message, created_at, read_at)
		 VALUES (?, ?, ?, ?, ?)`,
		notification.Type,
		notification.Title,
		notification.Message,
		notification.CreatedAt,
		readAt,
	)
	if err != nil {
		return nil, fmt.Errorf("notification repo: create: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("notification repo: create id: %w", err)
	}

	copy := *notification
	copy.ID = id
	return &copy, nil
}

// List returns stored notifications ordered by created time.
func (r *Repository) List() ([]*notificationdomain.Notification, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("notification repo: db required")
	}

	rows, err := r.db.Query(
		`SELECT id, type, title, message, created_at, read_at
		 FROM notifications
		 ORDER BY created_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("notification repo: list: %w", err)
	}
	defer func() {
		_ = rows.Close()
	}()

	results := make([]*notificationdomain.Notification, 0)
	for rows.Next() {
		var notification notificationdomain.Notification
		var readAt sql.NullInt64
		if err := rows.Scan(
			&notification.ID,
			&notification.Type,
			&notification.Title,
			&notification.Message,
			&notification.CreatedAt,
			&readAt,
		); err != nil {
			return nil, fmt.Errorf("notification repo: scan: %w", err)
		}
		if readAt.Valid {
			value := readAt.Int64
			notification.ReadAt = &value
		}
		results = append(results, &notification)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("notification repo: rows: %w", err)
	}

	return results, nil
}
