// credential_policy_test.go verifies secret-like credential key classification.
// internal/features/providers/interfaces/core/credential_policy_test.go
package core

import "testing"

// TestIsSensitiveCredentialName returns true for known secret key patterns.
func TestIsSensitiveCredentialName(t *testing.T) {

	testCases := []struct {
		name     string
		key      string
		expected bool
	}{
		{name: "api key constant", key: CredentialAPIKey, expected: true},
		{name: "token constant", key: CredentialToken, expected: true},
		{name: "cloudflare token constant", key: CredentialCloudflareToken, expected: true},
		{name: "mixed case api key", key: "Api_Key", expected: true},
		{name: "private key", key: "ssh_private_key", expected: true},
		{name: "password", key: "db_password", expected: true},
		{name: "account id", key: CredentialAccountID, expected: false},
		{name: "gateway id", key: CredentialGatewayID, expected: false},
		{name: "openrouter referer", key: CredentialOpenRouterReferer, expected: false},
		{name: "openrouter title", key: CredentialOpenRouterTitle, expected: false},
		{name: "blank", key: "   ", expected: false},
	}

	for _, testCase := range testCases {
		got := IsSensitiveCredentialName(testCase.key)
		if got != testCase.expected {
			t.Fatalf("%s: expected %v, got %v for key %q", testCase.name, testCase.expected, got, testCase.key)
		}
	}
}
