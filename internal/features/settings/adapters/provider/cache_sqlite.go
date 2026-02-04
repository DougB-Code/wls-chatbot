// cache_sqlite.go persists provider resource caches in SQLite.
// internal/features/settings/adapters/provider/cache_sqlite.go
package provider

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/MadeByDoug/wls-chatbot/internal/features/settings/ports"
)

const providerCacheSchema = `
CREATE TABLE IF NOT EXISTS provider_cache (
	id INTEGER PRIMARY KEY CHECK (id = 1),
	cache_json TEXT NOT NULL,
	updated_at INTEGER NOT NULL
);`

// Cache manages provider cache persistence in SQLite.
type Cache struct {
	db *sql.DB
}

// NewCache creates a new SQLite-backed cache instance.
func NewCache(db *sql.DB) (*Cache, error) {

	if db == nil {
		return nil, fmt.Errorf("provider cache: db required")
	}

	if _, err := db.Exec(providerCacheSchema); err != nil {
		return nil, fmt.Errorf("provider cache: ensure schema: %w", err)
	}

	return &Cache{db: db}, nil
}

var _ ports.ProviderCache = (*Cache)(nil)

// Load retrieves the cached resources.
func (c *Cache) Load() (ports.ProviderCacheSnapshot, error) {

	if c == nil || c.db == nil {
		return nil, fmt.Errorf("provider cache: db required")
	}

	row := c.db.QueryRow("SELECT cache_json FROM provider_cache WHERE id = 1")
	var data string
	if err := row.Scan(&data); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return make(ports.ProviderCacheSnapshot), nil
		}
		return nil, fmt.Errorf("provider cache: load: %w", err)
	}

	var snapshot ports.ProviderCacheSnapshot
	if err := json.Unmarshal([]byte(data), &snapshot); err != nil {
		return nil, fmt.Errorf("provider cache: decode: %w", err)
	}
	if snapshot == nil {
		snapshot = make(ports.ProviderCacheSnapshot)
	}
	return snapshot, nil
}

// Save persists the resources to the cache store.
func (c *Cache) Save(snapshot ports.ProviderCacheSnapshot) error {

	if c == nil || c.db == nil {
		return fmt.Errorf("provider cache: db required")
	}

	if snapshot == nil {
		snapshot = make(ports.ProviderCacheSnapshot)
	}

	data, err := json.Marshal(snapshot)
	if err != nil {
		return fmt.Errorf("provider cache: encode: %w", err)
	}

	now := time.Now().UnixMilli()
	_, err = c.db.Exec(
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
