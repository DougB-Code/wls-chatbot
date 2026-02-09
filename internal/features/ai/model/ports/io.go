// io.go defines model feature I/O ports used by app services.
// internal/features/ai/model/ports/io.go
package ports

import "database/sql"

// FileSystem defines file access required by model app services.
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	FileExists(path string) (bool, error)
}

// AppDataDirResolver resolves platform-specific application data directories.
type AppDataDirResolver interface {
	ResolveAppDataDir(appName string) (string, error)
}

// ModelSeeder imports model payloads into persistent storage.
type ModelSeeder interface {
	SeedModels(db *sql.DB, payload []byte) error
}
