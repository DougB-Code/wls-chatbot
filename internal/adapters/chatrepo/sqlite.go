// sqlite.go persists chat conversations in SQLite.
// internal/adapters/chatrepo/sqlite.go
package chatrepo

import (
	"database/sql"
	"encoding/json"
	"fmt"

	chatdomain "github.com/MadeByDoug/wls-chatbot/internal/core/domain/chat"
	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

const conversationSchema = `
CREATE TABLE IF NOT EXISTS conversations (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	is_archived INTEGER NOT NULL,
	data_json TEXT NOT NULL
);`

// Repository stores conversations in SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed chat repository.
func NewRepository(db *sql.DB) (*Repository, error) {

	if db == nil {
		return nil, fmt.Errorf("chat repo: db required")
	}

	if _, err := db.Exec(conversationSchema); err != nil {
		return nil, fmt.Errorf("chat repo: ensure schema: %w", err)
	}

	return &Repository{db: db}, nil
}

var _ ports.ChatRepository = (*Repository)(nil)

// Create saves a new conversation.
func (r *Repository) Create(conv *chatdomain.Conversation) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("chat repo: db required")
	}
	if conv == nil || conv.ID == "" {
		return fmt.Errorf("chat repo: conversation required")
	}

	data, err := json.Marshal(conv)
	if err != nil {
		return fmt.Errorf("chat repo: encode: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO conversations (id, title, created_at, updated_at, is_archived, data_json)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		conv.ID,
		conv.Title,
		conv.CreatedAt,
		conv.UpdatedAt,
		boolToInt(conv.IsArchived),
		string(data),
	)
	if err != nil {
		return fmt.Errorf("chat repo: create: %w", err)
	}

	return nil
}

// Get returns a conversation by ID.
func (r *Repository) Get(id string) (*chatdomain.Conversation, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("chat repo: db required")
	}
	if id == "" {
		return nil, nil
	}

	row := r.db.QueryRow("SELECT data_json FROM conversations WHERE id = ?", id)
	var data string
	if err := row.Scan(&data); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("chat repo: get: %w", err)
	}

	var conv chatdomain.Conversation
	if err := json.Unmarshal([]byte(data), &conv); err != nil {
		return nil, fmt.Errorf("chat repo: decode: %w", err)
	}

	return &conv, nil
}

// List returns all conversations.
func (r *Repository) List() ([]*chatdomain.Conversation, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("chat repo: db required")
	}

	rows, err := r.db.Query("SELECT data_json FROM conversations ORDER BY updated_at DESC")
	if err != nil {
		return nil, fmt.Errorf("chat repo: list: %w", err)
	}
	defer rows.Close()

	convs := []*chatdomain.Conversation{}
	for rows.Next() {
		var data string
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("chat repo: list scan: %w", err)
		}
		var conv chatdomain.Conversation
		if err := json.Unmarshal([]byte(data), &conv); err != nil {
			return nil, fmt.Errorf("chat repo: list decode: %w", err)
		}
		convs = append(convs, &conv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat repo: list rows: %w", err)
	}

	return convs, nil
}

// Update persists a conversation update.
func (r *Repository) Update(conv *chatdomain.Conversation) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("chat repo: db required")
	}
	if conv == nil || conv.ID == "" {
		return fmt.Errorf("chat repo: conversation required")
	}

	data, err := json.Marshal(conv)
	if err != nil {
		return fmt.Errorf("chat repo: encode: %w", err)
	}

	_, err = r.db.Exec(
		`INSERT INTO conversations (id, title, created_at, updated_at, is_archived, data_json)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		  title = excluded.title,
		  created_at = excluded.created_at,
		  updated_at = excluded.updated_at,
		  is_archived = excluded.is_archived,
		  data_json = excluded.data_json`,
		conv.ID,
		conv.Title,
		conv.CreatedAt,
		conv.UpdatedAt,
		boolToInt(conv.IsArchived),
		string(data),
	)
	if err != nil {
		return fmt.Errorf("chat repo: update: %w", err)
	}

	return nil
}

// Delete removes a conversation by ID.
func (r *Repository) Delete(id string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("chat repo: db required")
	}
	if id == "" {
		return nil
	}

	_, err := r.db.Exec("DELETE FROM conversations WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("chat repo: delete: %w", err)
	}

	return nil
}

// boolToInt maps booleans into SQLite-friendly integers.
func boolToInt(value bool) int {

	if value {
		return 1
	}
	return 0
}
