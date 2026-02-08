// sqlite.go opens and configures the shared SQLite datastore.
// internal/core/adapters/datastore/sqlite.go
package datastore

import (
	"database/sql"
	"embed"
	"fmt"
	"os"
	"path/filepath"

	modelcatalog "github.com/MadeByDoug/wls-chatbot/pkg/models"
	"github.com/rs/zerolog/log"
	_ "modernc.org/sqlite"
)

//go:embed schema.sql seed.sql
var sqlFiles embed.FS

const minimumDefaultProviderCount = 5
const sqliteBusyTimeoutMillis = 5000

var requiredSchemaTables = []string{
	"app_config",
	"catalog_providers",
	"provider_inputs",
	"catalog_endpoints",
	"model_catalog_entries",
	"model_capabilities",
	"model_capabilities_input_modalities",
	"model_capabilities_output_modalities",
	"model_system_profile",
	"model_system_tags",
	"model_user_addenda",
	"model_user_tags",
	"roles",
	"role_required_input_modalities",
	"role_required_output_modalities",
	"role_assignments",
	"chat_conversations",
	"chat_messages",
	"chat_message_blocks",
}

// OpenSQLite opens the SQLite database at the provided path.
func OpenSQLite(path string) (*sql.DB, error) {

	if path == "" {
		return nil, fmt.Errorf("open sqlite: path required")
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("open sqlite: ensure directory: %w", err)
	}

	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)

	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode = WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec("PRAGMA synchronous = NORMAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec(fmt.Sprintf("PRAGMA busy_timeout = %d", sqliteBusyTimeoutMillis)); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}
	if _, err := db.Exec("PRAGMA temp_store = MEMORY"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	needsInit, err := shouldInitializeDatabase(db)
	if err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("open sqlite: check initialization state: %w", err)
	}
	if needsInit {
		log.Info().Str("path", path).Msg("Initializing SQLite datastore")
		if err := initDatabase(db); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("open sqlite: init failed: %w", err)
		}

		needsInit, err = shouldInitializeDatabase(db)
		if err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("open sqlite: verify initialization state: %w", err)
		}
		if needsInit {
			_ = db.Close()
			return nil, fmt.Errorf("open sqlite: init failed: datastore remains incomplete")
		}
	}

	return db, nil
}

// shouldInitializeDatabase determines whether schema and seed data are complete.
func shouldInitializeDatabase(db *sql.DB) (bool, error) {

	for _, tableName := range requiredSchemaTables {
		exists, err := tableExists(db, tableName)
		if err != nil {
			return false, err
		}
		if !exists {
			return true, nil
		}
	}

	appConfigRows, err := countRows(db, "SELECT COUNT(*) FROM app_config WHERE id = 1")
	if err != nil {
		return false, err
	}
	if appConfigRows == 0 {
		return true, nil
	}

	defaultProviders, err := countRows(db, "SELECT COUNT(*) FROM catalog_providers")
	if err != nil {
		return false, err
	}
	if defaultProviders < minimumDefaultProviderCount {
		return true, nil
	}

	seedModels, err := countRows(db, "SELECT COUNT(*) FROM model_catalog_entries WHERE source = 'seed'")
	if err != nil {
		return false, err
	}
	return seedModels == 0, nil
}

// tableExists returns whether a SQLite table exists.
func tableExists(db *sql.DB, tableName string) (bool, error) {

	var count int
	if err := db.QueryRow(
		"SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name = ?",
		tableName,
	).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

// countRows executes a scalar count query.
func countRows(db *sql.DB, query string, args ...any) (int, error) {

	var count int
	if err := db.QueryRow(query, args...).Scan(&count); err != nil {
		return 0, err
	}
	return count, nil
}

// initDatabase executes the schema and seed scripts.
func initDatabase(db *sql.DB) error {

	log.Debug().Msg("Applying schema initialization script")
	schema, err := sqlFiles.ReadFile("schema.sql")
	if err != nil {
		return fmt.Errorf("read schema: %w", err)
	}
	if _, err := db.Exec(string(schema)); err != nil {
		return fmt.Errorf("exec schema: %w", err)
	}

	log.Debug().Msg("Applying seed initialization script")
	seed, err := sqlFiles.ReadFile("seed.sql")
	if err != nil {
		return fmt.Errorf("read seed: %w", err)
	}
	if _, err := db.Exec(string(seed)); err != nil {
		return fmt.Errorf("exec seed: %w", err)
	}

	log.Debug().Msg("Applying model seed initialization script")
	if err := SeedModels(db, modelcatalog.EmbeddedYAML()); err != nil {
		return fmt.Errorf("seed models: %w", err)
	}
	log.Debug().Msg("SQLite datastore initialization completed")

	return nil
}
