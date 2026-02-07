// sqlite_test.go verifies catalog repository input and auth metadata guardrails.
// internal/features/catalog/adapters/catalogrepo/sqlite_test.go
package catalogrepo

import (
	"context"
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

// TestSaveProviderInputsRejectsSecretLikeKeys enforces non-secret provider input persistence.
func TestSaveProviderInputsRejectsSecretLikeKeys(t *testing.T) {

	repo := newTestRepository(t)
	seedTestProvider(t, repo, "openai")

	err := repo.SaveProviderInputs("openai", map[string]string{
		"api_key": "sk-test",
	})
	if err == nil {
		t.Fatalf("expected error for secret-like provider input key")
	}
	if !strings.Contains(err.Error(), "secret-like input key") {
		t.Fatalf("expected secret-like input key error, got %v", err)
	}
}

// TestUpsertEndpointRejectsUnexpectedAuthJSON ensures endpoint auth metadata is schema-restricted.
func TestUpsertEndpointRejectsUnexpectedAuthJSON(t *testing.T) {

	repo := newTestRepository(t)
	provider := seedTestProvider(t, repo, "openai")

	_, err := repo.UpsertEndpoint(context.Background(), EndpointRecord{
		ProviderID:       provider.ID,
		DisplayName:      "Primary",
		AdapterType:      "openai",
		BaseURL:          "https://api.openai.com/v1",
		RouteKind:        "direct",
		OriginProvider:   "openai",
		OriginRouteLabel: "default",
		AuthJSON:         `{"api_key":"sk-test"}`,
	})
	if err == nil {
		t.Fatalf("expected invalid auth metadata error")
	}
	if !strings.Contains(err.Error(), "invalid auth metadata") {
		t.Fatalf("expected invalid auth metadata error, got %v", err)
	}
}

// TestUpsertEndpointAcceptsStructuredAuthJSON allows credential field metadata without secret values.
func TestUpsertEndpointAcceptsStructuredAuthJSON(t *testing.T) {

	repo := newTestRepository(t)
	provider := seedTestProvider(t, repo, "openai")

	authJSON, err := MarshalAuthJSON(map[string]interface{}{
		"credential_fields": []map[string]interface{}{
			{
				"name":     "api_key",
				"required": true,
				"secret":   true,
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal auth json: %v", err)
	}

	record, err := repo.UpsertEndpoint(context.Background(), EndpointRecord{
		ProviderID:       provider.ID,
		DisplayName:      "Primary",
		AdapterType:      "openai",
		BaseURL:          "https://api.openai.com/v1",
		RouteKind:        "direct",
		OriginProvider:   "openai",
		OriginRouteLabel: "default",
		AuthJSON:         authJSON,
	})
	if err != nil {
		t.Fatalf("upsert endpoint: %v", err)
	}
	if record.AuthJSON == "" {
		t.Fatalf("expected normalized auth json to be stored")
	}
}

// TestUpsertEndpointRejectsSecretLikeFieldMarkedNonSecret enforces secret flags on sensitive credential names.
func TestUpsertEndpointRejectsSecretLikeFieldMarkedNonSecret(t *testing.T) {

	repo := newTestRepository(t)
	provider := seedTestProvider(t, repo, "openai")

	_, err := repo.UpsertEndpoint(context.Background(), EndpointRecord{
		ProviderID:       provider.ID,
		DisplayName:      "Primary",
		AdapterType:      "openai",
		BaseURL:          "https://api.openai.com/v1",
		RouteKind:        "direct",
		OriginProvider:   "openai",
		OriginRouteLabel: "default",
		AuthJSON:         `{"credential_fields":[{"name":"api_key","required":true,"secret":false}]}`,
	})
	if err == nil {
		t.Fatalf("expected secret flag validation error")
	}
	if !strings.Contains(err.Error(), "must be marked secret") {
		t.Fatalf("expected secret flag validation error, got %v", err)
	}
}

// TestListEndpointsAllowsNullAuthJSON ensures legacy rows with NULL auth_json are readable.
func TestListEndpointsAllowsNullAuthJSON(t *testing.T) {

	repo := newTestRepository(t)
	provider := seedTestProvider(t, repo, "openai")

	_, err := repo.db.ExecContext(
		context.Background(),
		`INSERT INTO catalog_endpoints (id, provider_id, display_name, adapter_type, base_url, route_kind, origin_provider, origin_route_label, auth_json, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, NULL, ?, ?)`,
		"endpoint-null-auth",
		provider.ID,
		"Primary",
		"openai",
		"https://api.openai.com/v1",
		"direct",
		"openai",
		"default",
		int64(1),
		int64(1),
	)
	if err != nil {
		t.Fatalf("insert endpoint with null auth_json: %v", err)
	}

	endpoints, err := repo.ListEndpoints(context.Background())
	if err != nil {
		t.Fatalf("list endpoints: %v", err)
	}
	if len(endpoints) != 1 {
		t.Fatalf("expected one endpoint, got %d", len(endpoints))
	}
	if endpoints[0].AuthJSON != "" {
		t.Fatalf("expected empty auth json for NULL value, got %q", endpoints[0].AuthJSON)
	}
}

// newTestRepository creates a repository backed by a temporary SQLite database.
func newTestRepository(t *testing.T) *Repository {

	t.Helper()
	path := filepath.Join(t.TempDir(), "catalog.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	repo, err := NewRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}
	return repo
}

// seedTestProvider inserts a baseline provider for endpoint/input persistence tests.
func seedTestProvider(t *testing.T, repo *Repository, name string) ProviderRecord {

	t.Helper()
	record, err := repo.EnsureProvider(context.Background(), ProviderRecord{
		Name:        name,
		DisplayName: strings.ToUpper(name),
		AdapterType: "openai",
		TrustMode:   "user_managed",
		BaseURL:     "https://api.openai.com/v1",
	})
	if err != nil {
		t.Fatalf("ensure provider: %v", err)
	}
	return record
}
