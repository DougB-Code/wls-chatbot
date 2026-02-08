// catalog_test.go verifies modeltest plans derived from the canonical model catalog.
// pkg/models/modeltest/catalog_test.go
package modeltest

import (
	"testing"

	modelcatalog "github.com/MadeByDoug/wls-chatbot/pkg/models"
)

// TestLoadEmbeddedCatalogPlanIncludesEveryModel ensures every canonical model is represented in the derived test plan.
func TestLoadEmbeddedCatalogPlanIncludesEveryModel(t *testing.T) {

	catalog, err := modelcatalog.LoadEmbedded()
	if err != nil {
		t.Fatalf("load embedded model catalog: %v", err)
	}
	plan, err := LoadEmbeddedCatalogPlan()
	if err != nil {
		t.Fatalf("load embedded catalog plan: %v", err)
	}

	plannedModels := make(map[string]int)
	for _, provider := range plan.Providers {
		if len(provider.Models) == 0 {
			t.Fatalf("provider %q has zero planned models", provider.Name)
		}
		for _, model := range provider.Models {
			if len(model.Capabilities) == 0 {
				t.Fatalf("provider %q model %q has no planned capabilities", provider.Name, model.ID)
			}
			plannedModels[model.ID]++
		}
	}

	for _, model := range catalog.Models {
		count := plannedModels[model.ID]
		if count == 0 {
			t.Fatalf("model %q missing from derived plan", model.ID)
		}
		if count > 1 {
			t.Fatalf("model %q appears %d times in derived plan", model.ID, count)
		}
	}
}
