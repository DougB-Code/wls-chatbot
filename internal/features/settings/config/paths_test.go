// paths_test.go verifies application path resolution.
// internal/features/settings/config/paths_test.go
package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

// TestResolveAppDataDirRequiresName verifies empty app names fail.
func TestResolveAppDataDirRequiresName(t *testing.T) {

	if _, err := ResolveAppDataDir(""); err == nil {
		t.Fatalf("expected error for empty app name")
	}
}

// TestResolveAppDataDirUsesUserConfigDir verifies the app data dir uses the user config dir.
func TestResolveAppDataDirUsesUserConfigDir(t *testing.T) {

	root := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", root)

	path, err := ResolveAppDataDir("wls-test")
	if err != nil {
		t.Fatalf("resolve path: %v", err)
	}

	expected := filepath.Join(root, "wls-test")
	if path != expected {
		t.Fatalf("expected %s, got %s", expected, path)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected app data dir to exist: %v", err)
	}
}

// TestResolveAppDataDirErrorsWithoutConfigDir verifies missing config dirs return errors.
func TestResolveAppDataDirErrorsWithoutConfigDir(t *testing.T) {

	if runtime.GOOS == "windows" {
		t.Skip("user config dir env handling differs on windows")
	}

	t.Setenv("XDG_CONFIG_HOME", "")
	t.Setenv("HOME", "")

	if _, err := ResolveAppDataDir("wls-test"); err == nil {
		t.Fatalf("expected error when config dir is unavailable")
	}
}
