// events.go registers strongly typed provider event signals for frontend transport.
// internal/features/ai/providers/app/provider/events.go
package provider

import coreevents "github.com/MadeByDoug/wls-chatbot/internal/core/events"

var SignalProvidersUpdated = coreevents.MustRegister[coreevents.EmptyPayload]("providers:updated")
