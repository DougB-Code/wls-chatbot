// store.go defines configuration persistence contracts.
// internal/features/settings/config/store.go
package config

import "errors"

// ErrConfigNotFound signals the configuration has not been stored yet.
var ErrConfigNotFound = errors.New("config not found")

// Store persists application configuration.
type Store interface {
	Load() (AppConfig, error)
	Save(AppConfig) error
}
