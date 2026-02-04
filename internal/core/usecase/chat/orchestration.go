// orchestrate chat workflows, streaming, and event emission.
// internal/core/usecase/chat/orchestration.go
package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/MadeByDoug/wls-chatbot/internal/core/ports"
)

// Orchestrator coordinates chat workflows and event emission.
type Orchestrator struct {
	service  *Service
	registry ports.ProviderRegistry
	secrets  ports.SecretStore
	emitter  ports.ChatEmitter
	stream   *streamManager
}

// NewOrchestrator creates a chat orchestrator with required dependencies.
func NewOrchestrator(chatService *Service, registry ports.ProviderRegistry, secrets ports.SecretStore, emitter ports.ChatEmitter) *Orchestrator {

	return &Orchestrator{
		service:  chatService,
		registry: registry,
		secrets:  secrets,
		emitter:  emitter,
		stream:   newStreamManager(),
	}
}

// CreateConversation creates a new conversation with the given settings.
func (o *Orchestrator) CreateConversation(providerName, model string) *Conversation {

	return o.service.CreateConversation(ConversationSettings{
		Provider: providerName,
		Model:    model,
	})
}

// SetActiveConversation sets the active conversation by ID.
func (o *Orchestrator) SetActiveConversation(id string) {

	o.service.SetActiveConversation(id)
}

// GetActiveConversation returns the currently active conversation.
func (o *Orchestrator) GetActiveConversation() *Conversation {

	id := o.service.ActiveConversationID()
	if id == "" {
		return nil
	}
	return o.service.GetConversation(id)
}

// GetConversation returns a conversation by ID.
func (o *Orchestrator) GetConversation(id string) *Conversation {

	return o.service.GetConversation(id)
}

// ListConversations returns summaries of all conversations.
func (o *Orchestrator) ListConversations() []ConversationSummary {

	return o.service.ListConversations()
}

// ListDeletedConversations returns summaries of archived conversations.
func (o *Orchestrator) ListDeletedConversations() []ConversationSummary {

	return o.service.ListDeletedConversations()
}

// UpdateConversationModel updates the model for a conversation.
func (o *Orchestrator) UpdateConversationModel(conversationID, model string) bool {

	return o.service.UpdateConversationModel(conversationID, model)
}

// DeleteConversation archives a conversation by ID.
func (o *Orchestrator) DeleteConversation(id string) bool {

	return o.service.DeleteConversation(id)
}

// RestoreConversation restores an archived conversation by ID.
func (o *Orchestrator) RestoreConversation(id string) bool {

	return o.service.RestoreConversation(id)
}

// PurgeConversation permanently deletes a conversation by ID.
func (o *Orchestrator) PurgeConversation(id string) bool {

	return o.service.PurgeConversation(id)
}

// SendMessage sends a user message and initiates a streaming response.
func (o *Orchestrator) SendMessage(ctx context.Context, conversationID, content string) (*Message, error) {

	conversationID = strings.TrimSpace(conversationID)
	content = strings.TrimSpace(content)
	if conversationID == "" {
		return nil, errors.New("conversation ID required")
	}
	if content == "" {
		return nil, errors.New("message content required")
	}

	userMsg := o.service.AddMessage(conversationID, RoleUser, content)
	if userMsg == nil {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}

	o.maybeAutoTitle(conversationID, userMsg)

	o.emitChatEvent(ports.ChatEvent{
		Type:           "chat.message",
		ConversationID: conversationID,
		MessageID:      userMsg.ID,
		Timestamp:      time.Now().UnixMilli(),
		Payload:        userMsg,
	})

	conv := o.service.GetConversation(conversationID)
	if conv == nil {
		return userMsg, nil
	}

	providerName := strings.TrimSpace(conv.Settings.Provider)
	if providerName == "" {
		return userMsg, nil
	}

	streamMsg := o.service.CreateStreamingMessage(conversationID, RoleAssistant)
	if streamMsg == nil {
		return userMsg, nil
	}

	o.emitChatEvent(ports.ChatEvent{
		Type:           "chat.stream.start",
		ConversationID: conversationID,
		MessageID:      streamMsg.ID,
		Timestamp:      time.Now().UnixMilli(),
		Payload:        streamMsg,
	})

	prov, err := o.ensureProviderConfigured(providerName)
	if err != nil {
		o.emitStreamError(conversationID, streamMsg.ID, err)
		metadata := o.buildMetadata(providerName, conv.Settings.Model, "error", nil, time.Now(), err)
		o.service.FinalizeMessage(conversationID, streamMsg.ID, metadata)
		return userMsg, nil
	}

	chatMessages := o.buildProviderMessages(conv, streamMsg.ID)
	opts := ports.ChatOptions{
		Model:       conv.Settings.Model,
		Temperature: conv.Settings.Temperature,
		MaxTokens:   conv.Settings.MaxTokens,
		Stream:      true,
	}

	ctx, cancel := context.WithCancel(ctx)
	o.stream.start(conversationID, streamMsg.ID, cancel)

	chunks, err := prov.Chat(ctx, chatMessages, opts)
	if err != nil {
		o.stream.clear(conversationID, streamMsg.ID)
		o.emitStreamError(conversationID, streamMsg.ID, err)
		metadata := o.buildMetadata(providerName, conv.Settings.Model, "error", nil, time.Now(), err)
		o.service.FinalizeMessage(conversationID, streamMsg.ID, metadata)
		return userMsg, nil
	}

	go o.consumeStream(conversationID, streamMsg.ID, providerName, conv.Settings.Model, chunks)

	return userMsg, nil
}

// StopStream cancels the currently running stream.
func (o *Orchestrator) StopStream() {

	o.stream.stop()
}

// emitChatEvent sends a chat event through the emitter if available.
func (o *Orchestrator) emitChatEvent(event ports.ChatEvent) {

	if o.emitter == nil {
		return
	}
	o.emitter.EmitChatEvent(event)
}

// emitStreamChunk publishes a streaming chunk event.
func (o *Orchestrator) emitStreamChunk(conversationID, messageID string, blockIndex int, content string) {

	o.emitChatEvent(ports.ChatEvent{
		Type:           "chat.stream.chunk",
		ConversationID: conversationID,
		MessageID:      messageID,
		Timestamp:      time.Now().UnixMilli(),
		Payload: ports.StreamChunkPayload{
			BlockIndex: blockIndex,
			Content:    content,
			IsDone:     false,
		},
	})
}

// emitStreamError publishes a streaming error event.
func (o *Orchestrator) emitStreamError(conversationID, messageID string, err error) {

	payload := ports.StreamChunkPayload{
		BlockIndex: 0,
		Content:    "",
		IsDone:     true,
		Error:      err.Error(),
		StatusCode: ports.StatusCodeFromErr(err),
	}
	o.emitChatEvent(ports.ChatEvent{
		Type:           "chat.stream.error",
		ConversationID: conversationID,
		MessageID:      messageID,
		Timestamp:      time.Now().UnixMilli(),
		Payload:        payload,
	})
}

// emitStreamComplete publishes a stream completion event.
func (o *Orchestrator) emitStreamComplete(conversationID, messageID string, metadata *MessageMetadata) {

	o.emitChatEvent(ports.ChatEvent{
		Type:           "chat.stream.complete",
		ConversationID: conversationID,
		MessageID:      messageID,
		Timestamp:      time.Now().UnixMilli(),
		Payload: ports.StreamChunkPayload{
			BlockIndex: 0,
			Content:    "",
			IsDone:     true,
			Metadata:   metadata,
		},
	})
}

// ensureProviderConfigured returns a configured provider or an error.
func (o *Orchestrator) ensureProviderConfigured(name string) (ports.Provider, error) {

	if o.registry == nil {
		return nil, fmt.Errorf("provider registry not configured")
	}
	if o.secrets == nil {
		return nil, fmt.Errorf("secret store not configured")
	}
	prov := o.registry.Get(name)
	if prov == nil {
		return nil, fmt.Errorf("provider not found: %s", name)
	}
	key, err := o.secrets.GetProviderKey(name)
	if err != nil || key == "" {
		return nil, fmt.Errorf("no API key configured for %s", name)
	}
	_ = prov.Configure(ports.ProviderConfig{APIKey: key})
	return prov, nil
}

// buildProviderMessages builds the provider-facing message list.
func (o *Orchestrator) buildProviderMessages(conv *Conversation, streamingMessageID string) []ports.ProviderMessage {

	conv.Lock()
	defer conv.Unlock()

	messages := make([]ports.ProviderMessage, 0, len(conv.Messages)+1)
	systemPrompt := strings.TrimSpace(conv.Settings.SystemPrompt)
	if systemPrompt != "" {
		messages = append(messages, ports.ProviderMessage{
			Role:    ports.RoleSystem,
			Content: systemPrompt,
		})
	}

	for _, msg := range conv.Messages {
		if msg.ID == streamingMessageID && len(msg.Blocks) == 0 {
			continue
		}
		content := textFromBlocks(msg.Blocks)
		if strings.TrimSpace(content) == "" {
			continue
		}
		messages = append(messages, ports.ProviderMessage{
			Role:    ports.Role(msg.Role),
			Content: content,
		})
	}
	return messages
}

// textFromBlocks builds a text-only content string from message blocks.
func textFromBlocks(blocks []Block) string {

	if len(blocks) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, block := range blocks {
		if block.Type == BlockTypeText {
			builder.WriteString(block.Content)
		}
	}
	return builder.String()
}

// consumeStream handles incoming provider chunks and emits events.
func (o *Orchestrator) consumeStream(conversationID, messageID, providerName, fallbackModel string, chunks <-chan ports.Chunk) {

	defer o.stream.clear(conversationID, messageID)

	start := time.Now()
	var (
		finishReason string
		usage        *ports.UsageStats
		model        string
	)

	for chunk := range chunks {
		if chunk.Error != nil {
			if isContextCanceled(chunk.Error) {
				metadata := o.buildMetadata(providerName, chooseModel(model, fallbackModel), "cancelled", usage, start, nil)
				o.service.FinalizeMessage(conversationID, messageID, metadata)
				o.emitStreamComplete(conversationID, messageID, metadata)
			} else {
				o.emitStreamError(conversationID, messageID, chunk.Error)
				metadata := o.buildMetadata(providerName, chooseModel(model, fallbackModel), "error", usage, start, chunk.Error)
				o.service.FinalizeMessage(conversationID, messageID, metadata)
			}
			return
		}

		if chunk.Content != "" {
			o.service.AppendToMessage(conversationID, messageID, 0, chunk.Content)
			o.emitStreamChunk(conversationID, messageID, 0, chunk.Content)
		}
		if chunk.Model != "" {
			model = chunk.Model
		}
		if chunk.Usage != nil {
			usage = chunk.Usage
		}
		if chunk.FinishReason != "" {
			finishReason = chunk.FinishReason
		}
	}

	if finishReason == "" && o.stream.wasCancelled(conversationID, messageID) {
		finishReason = "cancelled"
	}

	metadata := o.buildMetadata(providerName, chooseModel(model, fallbackModel), finishReason, usage, start, nil)
	o.service.FinalizeMessage(conversationID, messageID, metadata)
	o.emitStreamComplete(conversationID, messageID, metadata)
}

// buildMetadata builds message metadata from provider results.
func (o *Orchestrator) buildMetadata(
	providerName,
	model,
	finishReason string,
	usage *ports.UsageStats,
	start time.Time,
	err error,
) *MessageMetadata {

	meta := &MessageMetadata{
		Provider:     providerName,
		Model:        model,
		FinishReason: finishReason,
		LatencyMs:    time.Since(start).Milliseconds(),
	}
	if meta.FinishReason == "" {
		meta.FinishReason = "stop"
	}
	if usage != nil {
		meta.TokensIn = usage.PromptTokens
		meta.TokensOut = usage.CompletionTokens
		meta.TokensTotal = usage.TotalTokens
		if meta.TokensTotal == 0 {
			meta.TokensTotal = usage.PromptTokens + usage.CompletionTokens
		}
	}
	if err != nil {
		meta.StatusCode = ports.StatusCodeFromErr(err)
		meta.ErrorMessage = err.Error()
	}
	return meta
}

// maybeAutoTitle updates the conversation title on the first user message.
func (o *Orchestrator) maybeAutoTitle(conversationID string, message *Message) {

	if message.Role != RoleUser {
		return
	}
	conv := o.service.GetConversation(conversationID)
	if conv == nil {
		return
	}
	conv.Lock()
	messageCount := len(conv.Messages)
	conv.Unlock()
	if messageCount != 1 {
		return
	}
	if len(message.Blocks) == 0 {
		return
	}
	content := message.Blocks[0].Content
	if content == "" {
		return
	}
	if len(content) > 50 {
		content = content[:50] + "..."
	}
	if o.service.SetConversationTitle(conversationID, content) {
		o.emitChatEvent(ports.ChatEvent{
			Type:           "chat.conversation.title",
			ConversationID: conversationID,
			Timestamp:      time.Now().UnixMilli(),
			Payload: ports.ConversationTitlePayload{
				Title: content,
			},
		})
	}
}

// chooseModel selects the best available model name.
func chooseModel(primary, fallback string) string {

	if strings.TrimSpace(primary) != "" {
		return primary
	}
	return fallback
}

// isContextCanceled checks for context cancellation errors.
func isContextCanceled(err error) bool {

	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}
