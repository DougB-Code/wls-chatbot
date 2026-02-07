// sqlite_test.go verifies SQLite datastore initialization and recovery behavior.
// internal/core/adapters/datastore/sqlite_test.go
package datastore

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TestOpenSQLiteInitializesFreshDatabase verifies a new datastore path is fully initialized.
func TestOpenSQLiteInitializesFreshDatabase(t *testing.T) {

	databasePath := newTestDatabasePath(t, filepath.Join("fresh", "app.db"))
	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	assertDatastoreReady(t, db)

	providerCount := mustCountRows(t, db, "SELECT COUNT(*) FROM catalog_providers")
	if providerCount < minimumDefaultProviderCount {
		t.Fatalf("expected at least %d providers, got %d", minimumDefaultProviderCount, providerCount)
	}

	seedModelCount := mustCountRows(t, db, "SELECT COUNT(*) FROM model_catalog_entries WHERE source = 'seed'")
	if seedModelCount == 0 {
		t.Fatalf("expected seeded models, got 0")
	}
}

// TestOpenSQLiteConfiguresRuntimePragmas verifies required SQLite runtime pragmas are enabled.
func TestOpenSQLiteConfiguresRuntimePragmas(t *testing.T) {

	databasePath := newTestDatabasePath(t, filepath.Join("pragmas", "app.db"))
	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var foreignKeys int
	if err := db.QueryRow("PRAGMA foreign_keys").Scan(&foreignKeys); err != nil {
		t.Fatalf("read foreign_keys pragma: %v", err)
	}
	if foreignKeys != 1 {
		t.Fatalf("expected foreign_keys=1, got %d", foreignKeys)
	}

	var busyTimeout int
	if err := db.QueryRow("PRAGMA busy_timeout").Scan(&busyTimeout); err != nil {
		t.Fatalf("read busy_timeout pragma: %v", err)
	}
	if busyTimeout != sqliteBusyTimeoutMillis {
		t.Fatalf("expected busy_timeout=%d, got %d", sqliteBusyTimeoutMillis, busyTimeout)
	}

	var tempStore int
	if err := db.QueryRow("PRAGMA temp_store").Scan(&tempStore); err != nil {
		t.Fatalf("read temp_store pragma: %v", err)
	}
	if tempStore != 2 {
		t.Fatalf("expected temp_store=2 (MEMORY), got %d", tempStore)
	}
}

// TestOpenSQLiteRepairsPartialSchemaLoad verifies missing tables are repaired after interrupted setup.
func TestOpenSQLiteRepairsPartialSchemaLoad(t *testing.T) {

	databasePath := newTestDatabasePath(t, "partial-schema.db")
	raw := openRawSQLiteDB(t, databasePath)
	_, err := raw.Exec(`
		CREATE TABLE catalog_providers (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL UNIQUE,
			display_name TEXT NOT NULL,
			adapter_type TEXT NOT NULL,
			trust_mode TEXT NOT NULL,
			base_url TEXT,
			created_at INTEGER NOT NULL,
			updated_at INTEGER NOT NULL
		)
	`)
	if err != nil {
		t.Fatalf("create partial table: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw sqlite: %v", err)
	}

	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite after partial schema: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	assertDatastoreReady(t, db)

	missingTableCount := 0
	for _, tableName := range requiredSchemaTables {
		exists, tableErr := tableExists(db, tableName)
		if tableErr != nil {
			t.Fatalf("check table %s: %v", tableName, tableErr)
		}
		if !exists {
			missingTableCount++
		}
	}
	if missingTableCount != 0 {
		t.Fatalf("expected all required tables after recovery, missing=%d", missingTableCount)
	}
}

// TestOpenSQLiteRepairsPartialSeedLoad verifies missing seed data is recovered without deleting user data.
func TestOpenSQLiteRepairsPartialSeedLoad(t *testing.T) {

	databasePath := newTestDatabasePath(t, "partial-seed.db")
	raw := openRawSQLiteDB(t, databasePath)
	schemaSQL, err := sqlFiles.ReadFile("schema.sql")
	if err != nil {
		t.Fatalf("read embedded schema: %v", err)
	}
	if _, err := raw.Exec(string(schemaSQL)); err != nil {
		t.Fatalf("exec schema: %v", err)
	}

	now := time.Now().UnixMilli()
	_, err = raw.Exec(
		"INSERT INTO app_config (id, config_json, created_at, updated_at) VALUES (1, ?, ?, ?)",
		`{"theme":"dark","providers":[]}`,
		now,
		now,
	)
	if err != nil {
		t.Fatalf("insert partial app config: %v", err)
	}

	_, err = raw.Exec(
		`INSERT INTO catalog_providers
		(id, name, display_name, adapter_type, trust_mode, base_url, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"prov_custom",
		"custom",
		"Custom Provider",
		"openai",
		"user_managed",
		"https://custom.example/v1",
		now,
		now,
	)
	if err != nil {
		t.Fatalf("insert custom provider: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw sqlite: %v", err)
	}

	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite after partial seed: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	assertDatastoreReady(t, db)

	customProviderCount := mustCountRows(
		t,
		db,
		"SELECT COUNT(*) FROM catalog_providers WHERE id = 'prov_custom' AND name = 'custom'",
	)
	if customProviderCount != 1 {
		t.Fatalf("expected custom provider to be preserved, got %d rows", customProviderCount)
	}

	providerCount := mustCountRows(t, db, "SELECT COUNT(*) FROM catalog_providers")
	if providerCount < minimumDefaultProviderCount+1 {
		t.Fatalf("expected default providers plus custom provider, got %d", providerCount)
	}
}

// TestOpenSQLitePreservesUserAddedModels verifies user model rows survive re-open initialization checks.
func TestOpenSQLitePreservesUserAddedModels(t *testing.T) {

	databasePath := newTestDatabasePath(t, "user-models.db")
	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}

	seedModelsBefore := mustCountRows(t, db, "SELECT COUNT(*) FROM model_catalog_entries WHERE source = 'seed'")
	now := time.Now().UnixMilli()
	customEndpointID := "ep_openai_user_custom"
	customEntryID := "entry_openai_custom_user_model"
	customModelID := "custom-user-model-v1"

	_, err = db.Exec(
		`INSERT INTO catalog_endpoints (
			id, provider_id, display_name, adapter_type, base_url, route_kind,
			origin_provider, origin_route_label, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		customEndpointID,
		"prov_openai",
		"OpenAI Custom Endpoint",
		"openai",
		"https://api.openai.com/v1",
		"llm",
		"openai",
		"user_custom",
		now,
		now,
	)
	if err != nil {
		_ = db.Close()
		t.Fatalf("insert custom endpoint: %v", err)
	}

	_, err = db.Exec(
		`INSERT INTO model_catalog_entries (
			id, endpoint_id, model_id, display_name, first_seen_at, last_seen_at,
			availability_state, approved, missed_refreshes, source, metadata_json
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		customEntryID,
		customEndpointID,
		customModelID,
		customModelID,
		now,
		now,
		"available",
		1,
		0,
		"user",
		`{"owner":"user"}`,
	)
	if err != nil {
		_ = db.Close()
		t.Fatalf("insert custom model entry: %v", err)
	}
	if err := db.Close(); err != nil {
		t.Fatalf("close sqlite before reopen: %v", err)
	}

	reopenedDB, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("reopen sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = reopenedDB.Close()
	})

	assertDatastoreReady(t, reopenedDB)

	customModelCount := mustCountRows(
		t,
		reopenedDB,
		"SELECT COUNT(*) FROM model_catalog_entries WHERE id = ? AND source = 'user'",
		customEntryID,
	)
	if customModelCount != 1 {
		t.Fatalf("expected custom user model to be preserved, got %d rows", customModelCount)
	}

	seedModelsAfter := mustCountRows(t, reopenedDB, "SELECT COUNT(*) FROM model_catalog_entries WHERE source = 'seed'")
	if seedModelsAfter < seedModelsBefore {
		t.Fatalf("expected seed models to remain available, before=%d after=%d", seedModelsBefore, seedModelsAfter)
	}
}

// TestSeedModelsReconcilesEndpointFields verifies endpoint fields are updated on seed conflicts.
func TestSeedModelsReconcilesEndpointFields(t *testing.T) {

	databasePath := newTestDatabasePath(t, "seed-reconcile.db")
	db, err := OpenSQLite(databasePath)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	updatedBaseURL := "https://openai.example/v2"
	_, err = db.Exec(`UPDATE catalog_providers SET base_url = ? WHERE name = 'openai'`, updatedBaseURL)
	if err != nil {
		t.Fatalf("update provider base url: %v", err)
	}

	models, err := sqlFiles.ReadFile("seeds/models.yaml")
	if err != nil {
		t.Fatalf("read embedded models: %v", err)
	}
	if err := SeedModels(db, models); err != nil {
		t.Fatalf("reseed models: %v", err)
	}

	var baseURL string
	if err := db.QueryRow(`SELECT base_url FROM catalog_endpoints WHERE id = 'ep_openai_default'`).Scan(&baseURL); err != nil {
		t.Fatalf("load endpoint base url: %v", err)
	}
	if baseURL != updatedBaseURL {
		t.Fatalf("expected endpoint base_url=%q, got %q", updatedBaseURL, baseURL)
	}
}

// TestOpenSQLiteRejectsEmptyPath verifies path validation.
func TestOpenSQLiteRejectsEmptyPath(t *testing.T) {

	db, err := OpenSQLite("")
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatalf("expected error for empty path")
	}
	if !strings.Contains(err.Error(), "path required") {
		t.Fatalf("expected path required error, got %v", err)
	}
}

// TestOpenSQLiteRejectsParentAsFile verifies directory creation failures are reported.
func TestOpenSQLiteRejectsParentAsFile(t *testing.T) {

	databasePath := newTestDatabasePath(t, filepath.Join("parent-file", "blocked", "db.sqlite"))
	parentPath := filepath.Dir(databasePath)
	if err := os.MkdirAll(filepath.Dir(parentPath), 0o755); err != nil {
		t.Fatalf("mkdir parent: %v", err)
	}
	if err := os.WriteFile(parentPath, []byte("blocking file"), 0o644); err != nil {
		t.Fatalf("create blocking file: %v", err)
	}

	db, err := OpenSQLite(databasePath)
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatalf("expected error when parent path is a file")
	}
	if !strings.Contains(err.Error(), "ensure directory") {
		t.Fatalf("expected directory error, got %v", err)
	}
}

// TestOpenSQLiteRejectsIncompatibleExistingSchema verifies unrecoverable schema drift fails fast.
func TestOpenSQLiteRejectsIncompatibleExistingSchema(t *testing.T) {

	databasePath := newTestDatabasePath(t, "incompatible-schema.db")
	raw := openRawSQLiteDB(t, databasePath)
	_, err := raw.Exec(`CREATE TABLE catalog_providers (id TEXT PRIMARY KEY)`)
	if err != nil {
		t.Fatalf("create incompatible table: %v", err)
	}
	if err := raw.Close(); err != nil {
		t.Fatalf("close raw sqlite: %v", err)
	}

	db, err := OpenSQLite(databasePath)
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatalf("expected error for incompatible schema")
	}
	if !strings.Contains(err.Error(), "init failed") {
		t.Fatalf("expected init failed error, got %v", err)
	}
}

// TestOpenSQLiteRejectsCorruptDatabaseFile verifies invalid SQLite files fail initialization.
func TestOpenSQLiteRejectsCorruptDatabaseFile(t *testing.T) {

	databasePath := newTestDatabasePath(t, "corrupt.db")
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		t.Fatalf("mkdir database parent: %v", err)
	}
	if err := os.WriteFile(databasePath, []byte("this is not a sqlite database"), 0o644); err != nil {
		t.Fatalf("write corrupt database file: %v", err)
	}

	db, err := OpenSQLite(databasePath)
	if err == nil {
		if db != nil {
			_ = db.Close()
		}
		t.Fatalf("expected error for corrupt database file")
	}
}

// newTestDatabasePath creates a temporary database path and preserves artifacts on failure.
func newTestDatabasePath(t *testing.T, relativePath string) string {

	t.Helper()
	rootDir, err := os.MkdirTemp("", "wls-chatbot-datastore-*")
	if err != nil {
		t.Fatalf("create temp directory: %v", err)
	}
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("preserving failed test artifact: %s", rootDir)
			return
		}
		_ = os.RemoveAll(rootDir)
	})
	return filepath.Join(rootDir, relativePath)
}

// openRawSQLiteDB opens a raw sqlite connection for controlled pre-test setup.
func openRawSQLiteDB(t *testing.T, databasePath string) *sql.DB {

	t.Helper()
	parentDir := filepath.Dir(databasePath)
	if err := os.MkdirAll(parentDir, 0o755); err != nil {
		t.Fatalf("mkdir database parent directory: %v", err)
	}
	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		t.Fatalf("open raw sqlite: %v", err)
	}
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Fatalf("ping raw sqlite: %v", err)
	}
	return db
}

// assertDatastoreReady verifies schema and seed readiness checks pass.
func assertDatastoreReady(t *testing.T, db *sql.DB) {

	t.Helper()
	needsInit, err := shouldInitializeDatabase(db)
	if err != nil {
		t.Fatalf("check initialization readiness: %v", err)
	}
	if needsInit {
		t.Fatalf("expected datastore to be fully initialized")
	}
}

// mustCountRows executes a count query and fails the test on query errors.
func mustCountRows(t *testing.T, db *sql.DB, query string, args ...any) int {

	t.Helper()
	count, err := countRows(db, query, args...)
	if err != nil {
		t.Fatalf("count rows with query %q: %v", query, err)
	}
	return count
}
