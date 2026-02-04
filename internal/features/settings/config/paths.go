// paths.go resolves OS-specific application paths.
// internal/features/settings/config/paths.go
package config

import (
	"fmt"
	"os"
	"path/filepath"
)

// ResolveAppDataDir resolves and creates the OS config directory for the app.
func ResolveAppDataDir(appName string) (string, error) {

	if appName == "" {
		return "", fmt.Errorf("resolve app data dir: app name required")
	}

	configDir, err := os.UserConfigDir()
	if err != nil || configDir == "" {
		return "", fmt.Errorf("resolve user config dir: %w", err)
	}

	appDir := filepath.Join(configDir, appName)
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		return "", fmt.Errorf("create app data dir: %w", err)
	}

	return appDir, nil
}
