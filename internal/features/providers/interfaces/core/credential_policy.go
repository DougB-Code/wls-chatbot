// credential_policy.go classifies provider input keys that must be treated as secrets.
// internal/features/providers/interfaces/core/credential_policy.go
package core

import "strings"

var sensitiveCredentialNameTokens = []string{
	"api_key",
	"apikey",
	"token",
	"secret",
	"password",
	"private_key",
	"access_key",
	"client_secret",
}

// IsSensitiveCredentialName reports whether a credential key is secret-like and must not be persisted as plain input.
func IsSensitiveCredentialName(name string) bool {

	normalized := strings.ToLower(strings.TrimSpace(name))
	if normalized == "" {
		return false
	}

	switch normalized {
	case CredentialAPIKey, CredentialToken, CredentialCloudflareToken:
		return true
	}

	for _, token := range sensitiveCredentialNameTokens {
		if strings.Contains(normalized, token) {
			return true
		}
	}

	return false
}
