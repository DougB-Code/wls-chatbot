// chat_service.go executes provider-backed stateless chat completions.
// internal/features/ai/chat/app/chat/chat_service.go
package chat

import (
	"context"
	"fmt"
	"strings"

	aiinterfaces "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
	providercore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/core"
	providergateway "github.com/MadeByDoug/wls-chatbot/internal/features/ai/providers/ports/gateway"
)

// ChatService handles chat operations for transport adapters.
type ChatService struct {
	registry providercore.ProviderRegistry
	secrets  providercore.SecretStore
}

var _ aiinterfaces.ChatInterface = (*ChatService)(nil)

// NewChatService creates a chat backend service.
func NewChatService(registry providercore.ProviderRegistry, secrets providercore.SecretStore) *ChatService {

	return &ChatService{
		registry: registry,
		secrets:  secrets,
	}
}

// Chat streams chat completion chunks.
func (s *ChatService) Chat(ctx context.Context, request aiinterfaces.ChatRequest) (<-chan aiinterfaces.ChatChunk, error) {

	providerName := strings.TrimSpace(request.ProviderName)
	if providerName == "" {
		return nil, fmt.Errorf("provider name required")
	}
	if s.registry == nil {
		return nil, fmt.Errorf("provider registry not configured")
	}

	prov := s.registry.Get(providerName)
	if prov == nil {
		return nil, fmt.Errorf("provider not found: %s", providerName)
	}
	if err := s.configureProvider(providerName, prov); err != nil {
		return nil, err
	}

	messages, err := toProviderMessages(request.Messages)
	if err != nil {
		return nil, err
	}

	chunks, err := prov.Chat(ctx, messages, toProviderChatOptions(request))
	if err != nil {
		return nil, err
	}
	return bridgeChatChunks(chunks), nil
}

// configureProvider hydrates required secret credentials before chat execution.
func (s *ChatService) configureProvider(providerName string, provider providercore.Provider) error {

	credentials := make(providercore.ProviderCredentials)
	for _, field := range provider.CredentialFields() {
		if !field.Secret {
			continue
		}
		if s.secrets == nil {
			if field.Required {
				return fmt.Errorf("secret store not configured")
			}
			continue
		}

		value, err := s.secrets.GetProviderSecret(providerName, field.Name)
		if err != nil || strings.TrimSpace(value) == "" {
			if field.Required {
				return fmt.Errorf("missing required credential: %s", field.Name)
			}
			continue
		}
		credentials[field.Name] = value
	}
	if len(credentials) == 0 {
		return nil
	}
	return provider.Configure(providercore.ProviderConfig{Credentials: credentials})
}

// toProviderMessages converts transport chat messages to provider messages.
func toProviderMessages(messages []aiinterfaces.ChatMessage) ([]providergateway.ProviderMessage, error) {

	converted := make([]providergateway.ProviderMessage, 0, len(messages))
	for _, message := range messages {
		if strings.TrimSpace(message.Content) == "" {
			continue
		}
		role, err := toProviderRole(message.Role)
		if err != nil {
			return nil, err
		}
		converted = append(converted, providergateway.ProviderMessage{
			Role:    role,
			Content: message.Content,
		})
	}
	return converted, nil
}

// toProviderRole converts transport chat roles into provider roles.
func toProviderRole(role aiinterfaces.ChatRole) (providergateway.Role, error) {

	switch role {
	case aiinterfaces.ChatRoleUser:
		return providergateway.RoleUser, nil
	case aiinterfaces.ChatRoleAssistant:
		return providergateway.RoleAssistant, nil
	case aiinterfaces.ChatRoleSystem:
		return providergateway.RoleSystem, nil
	case aiinterfaces.ChatRoleTool:
		return providergateway.RoleTool, nil
	default:
		return "", fmt.Errorf("unsupported chat role: %s", role)
	}
}

// toProviderChatOptions converts transport chat options to provider options.
func toProviderChatOptions(request aiinterfaces.ChatRequest) providergateway.ChatOptions {

	return providergateway.ChatOptions{
		Model:       request.ModelName,
		Temperature: request.Options.Temperature,
		MaxTokens:   request.Options.MaxTokens,
		Stream:      request.Options.Stream,
		Tools:       toProviderTools(request.Options.Tools),
		StopWords:   append([]string(nil), request.Options.StopWords...),
	}
}

// toProviderTools converts transport tool declarations to provider tool declarations.
func toProviderTools(tools []aiinterfaces.ChatTool) []providergateway.Tool {

	if len(tools) == 0 {
		return nil
	}

	converted := make([]providergateway.Tool, 0, len(tools))
	for _, tool := range tools {
		converted = append(converted, providergateway.Tool{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  tool.InputSchema,
		})
	}
	return converted
}

// bridgeChatChunks converts provider streaming chunks into transport chunks.
func bridgeChatChunks(chunks <-chan providergateway.Chunk) <-chan aiinterfaces.ChatChunk {

	out := make(chan aiinterfaces.ChatChunk)
	go func() {
		defer close(out)
		for chunk := range chunks {
			out <- toChatChunk(chunk)
		}
	}()
	return out
}

// toChatChunk converts one provider chunk into one transport chunk.
func toChatChunk(chunk providergateway.Chunk) aiinterfaces.ChatChunk {

	result := aiinterfaces.ChatChunk{
		Content:      chunk.Content,
		Model:        chunk.Model,
		FinishReason: chunk.FinishReason,
	}

	if len(chunk.ToolCalls) > 0 {
		toolCalls := make([]aiinterfaces.ChatToolCall, 0, len(chunk.ToolCalls))
		for _, call := range chunk.ToolCalls {
			toolCalls = append(toolCalls, aiinterfaces.ChatToolCall{
				ID:        call.ID,
				Name:      call.Name,
				Arguments: call.Arguments,
			})
		}
		result.ToolCalls = toolCalls
	}

	if chunk.Usage != nil {
		result.Usage = &aiinterfaces.ChatUsage{
			InputTokens:  chunk.Usage.PromptTokens,
			OutputTokens: chunk.Usage.CompletionTokens,
			TotalTokens:  chunk.Usage.TotalTokens,
		}
	}
	if chunk.Error != nil {
		result.Error = chunk.Error.Error()
	}

	return result
}
