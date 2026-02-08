// dependencies.go defines dependency contracts consumed by CLI command adapters.
// internal/ui/adapters/cli/dependencies.go
package cli

import (
	"fmt"

	commonadapter "github.com/MadeByDoug/wls-chatbot/internal/ui/adapters/common"
)

// Dependencies groups construction functions required by the CLI adapter.
type Dependencies struct {
	*commonadapter.Dependencies
}

// validate returns an error when required CLI adapter dependencies are missing.
func (d Dependencies) validate() error {

	if d.Dependencies == nil {
		return fmt.Errorf("cli adapter: common dependencies required")
	}
	if err := d.Dependencies.ValidateCore(); err != nil {
		return fmt.Errorf("cli adapter: %w", err)
	}
	return nil
}
