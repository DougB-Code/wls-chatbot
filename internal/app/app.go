// app.go defines the application facade used by all transport adapters.
// internal/app/app.go
package app

import (
	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
	imageports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/image/ports"
	modelinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/model/ports"
	providerfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/app/provider"
)

// App groups feature capabilities behind one application facade.
type App struct {
	Providers     *providerfeature.Orchestrator
	Models        modelinterfaces.ProviderModelInterface
	Images        imageports.ImageInterface
	Chat          chatports.ChatInterface
	Conversations *chatfeature.Orchestrator
}
