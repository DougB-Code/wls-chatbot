// catalog_test.go verifies canonical model catalog loading and validation.
// pkg/models/catalog_test.go
package models

import "testing"

// TestLoadEmbeddedReturnsValidCatalog ensures the embedded canonical catalog is structurally valid.
func TestLoadEmbeddedReturnsValidCatalog(t *testing.T) {

	catalog, err := LoadEmbedded()
	if err != nil {
		t.Fatalf("load embedded catalog: %v", err)
	}
	if len(catalog.Families) == 0 {
		t.Fatalf("expected at least one family in canonical catalog")
	}
	if len(catalog.Providers) == 0 {
		t.Fatalf("expected at least one provider in canonical catalog")
	}
	if len(catalog.Models) == 0 {
		t.Fatalf("expected at least one model in canonical catalog")
	}
}
