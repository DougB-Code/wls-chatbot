// dependencies.go defines dependency contracts consumed by AI CLI command adapters.
// internal/ui/adapters/cli/ai/dependencies.go
package ai

import (
	"fmt"

	commonadapter "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/common"
)

// Dependencies groups construction functions required by AI CLI adapters.
type Dependencies struct {
	*commonadapter.Dependencies
}

// validate returns an error when required AI CLI adapter dependencies are missing.
func (d Dependencies) validate() error {

	if d.Dependencies == nil {
		return fmt.Errorf("cli ai adapter: common dependencies required")
	}
	if err := d.Dependencies.ValidateCore(); err != nil {
		return fmt.Errorf("cli ai adapter: %w", err)
	}
	return nil
}
