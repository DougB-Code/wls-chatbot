// sqlite.go persists app configuration in SQLite.
// internal/adapters/configstore/sqlite.go
package configstore

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MadeByDoug/wls-chatbot/internal/app/config"
)

const appConfigSchema = `
CREATE TABLE IF NOT EXISTS app_config (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	config_json TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL
);`

// SQLiteStore persists application configuration in SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a SQLite-backed configuration store.
func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {

	if db == nil {
		return nil, fmt.Errorf("config store: db required")
	}

	if _, err := db.Exec(appConfigSchema); err != nil {
		return nil, fmt.Errorf("config store: ensure schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

var _ config.Store = (*SQLiteStore)(nil)

// Load returns the stored application configuration.
func (s *SQLiteStore) Load() (config.AppConfig, error) {

	if s == nil || s.db == nil {
		return config.AppConfig{}, fmt.Errorf("config store: db required")
	}

	row := s.db.QueryRow("SELECT config_json FROM app_config WHERE id = 1")
	var data string
	if err := row.Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return config.AppConfig{}, config.ErrConfigNotFound
		}
		return config.AppConfig{}, fmt.Errorf("config store: load: %w", err)
	}

	var cfg config.AppConfig
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return config.AppConfig{}, fmt.Errorf("config store: decode: %w", err)
	}

	return cfg, nil
}

// Save writes the application configuration to the store.
func (s *SQLiteStore) Save(cfg config.AppConfig) error {

	if s == nil || s.db == nil {
		return fmt.Errorf("config store: db required")
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("config store: encode: %w", err)
	}

	now := time.Now().UnixMilli()
	_, err = s.db.Exec(
		`INSERT INTO app_config (id, config_json, created_at, updated_at)
		 VALUES (1, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		  config_json = excluded.config_json,
		  updated_at = excluded.updated_at`,
		string(data),
		now,
		now,
	)
	if err != nil {
		return fmt.Errorf("config store: save: %w", err)
	}

	return nil
}
