// local.go provides local filesystem and platform app-data adapters for model services.
// internal/features/ai/model/adapters/io/local.go
package io

import (
	"fmt"
	"os"

	modelports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/ports"
	"github.com/MadeByDoug/wls-chatbot/internal/platform"
)

// LocalFileSystem implements model file-system operations with os package calls.
type LocalFileSystem struct{}

var _ modelports.FileSystem = (*LocalFileSystem)(nil)

// NewLocalFileSystem creates a local filesystem adapter.
func NewLocalFileSystem() *LocalFileSystem {

	return &LocalFileSystem{}
}

// ReadFile reads bytes from the requested file path.
func (*LocalFileSystem) ReadFile(path string) ([]byte, error) {

	return os.ReadFile(path)
}

// FileExists reports whether a file exists at the requested path.
func (*LocalFileSystem) FileExists(path string) (bool, error) {

	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, fmt.Errorf("stat model file: %w", err)
}

// PlatformAppDataDirResolver resolves app-data directories using platform conventions.
type PlatformAppDataDirResolver struct{}

var _ modelports.AppDataDirResolver = (*PlatformAppDataDirResolver)(nil)

// NewPlatformAppDataDirResolver creates an app-data directory resolver adapter.
func NewPlatformAppDataDirResolver() *PlatformAppDataDirResolver {

	return &PlatformAppDataDirResolver{}
}

// ResolveAppDataDir resolves the application data directory for an app name.
func (*PlatformAppDataDirResolver) ResolveAppDataDir(appName string) (string, error) {

	return platform.ResolveAppDataDir(appName)
}
