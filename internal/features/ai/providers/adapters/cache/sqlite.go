// sqlite.go persists provider resource cache snapshots in SQLite.
// internal/features/ai/providers/adapters/cache/sqlite.go
package cache

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/core"
)

const providerCacheSchema = `
CREATE TABLE IF NOT EXISTS provider_cache (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	cache_json TEXT NOT NULL,
	updated_at INTEGER NOT NULL
);`

// SQLiteStore manages provider cache persistence in SQLite.
type SQLiteStore struct {
	db *sql.DB
}

// NewSQLiteStore creates a SQLite-backed cache store.
func NewSQLiteStore(db *sql.DB) (*SQLiteStore, error) {

	if db == nil {
		return nil, fmt.Errorf("provider cache: db required")
	}

	if _, err := db.Exec(providerCacheSchema); err != nil {
		return nil, fmt.Errorf("provider cache: ensure schema: %w", err)
	}

	return &SQLiteStore{db: db}, nil
}

var _ providercore.ProviderCache = (*SQLiteStore)(nil)

// Load retrieves the cached resources.
func (s *SQLiteStore) Load() (providercore.ProviderCacheSnapshot, error) {

	if s == nil || s.db == nil {
		return nil, fmt.Errorf("provider cache: db required")
	}

	row := s.db.QueryRow("SELECT cache_json FROM provider_cache WHERE id = 1")
	var data string
	if err := row.Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make(providercore.ProviderCacheSnapshot), nil
		}
		return nil, fmt.Errorf("provider cache: load: %w", err)
	}

	var snapshot providercore.ProviderCacheSnapshot
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		return nil, fmt.Errorf("provider cache: decode: %w", err)
	}
	if snapshot == nil {
		snapshot = make(providercore.ProviderCacheSnapshot)
	}
	return snapshot, nil
}

// Save persists the resources to the cache store.
func (s *SQLiteStore) Save(snapshot providercore.ProviderCacheSnapshot) error {

	if s == nil || s.db == nil {
		return fmt.Errorf("provider cache: db required")
	}

	if snapshot == nil {
		snapshot = make(providercore.ProviderCacheSnapshot)
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("provider cache: encode: %w", err)
	}

	now := time.Now().UnixMilli()
	_, err = s.db.Exec(
		`INSERT INTO provider_cache (id, cache_json, updated_at)
		 VALUES (1, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		  cache_json = excluded.cache_json,
		  updated_at = excluded.updated_at`,
		string(data),
		now,
	)
	if err != nil {
		return fmt.Errorf("provider cache: save: %w", err)
	}

	return nil
}
