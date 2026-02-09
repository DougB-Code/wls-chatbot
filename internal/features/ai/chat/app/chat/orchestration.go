// orchestration.go orchestrates conversation state, persistence, and streaming events.
// internal/features/ai/chat/app/chat/orchestration.go
package chat

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	coreevents "github.com/MadeByDoug/wls-chatbot/internal/core/events"
	chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/domain"
	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/ports"
)

// Orchestrator coordinates chat workflows and event emission.
type Orchestrator struct {
	service *Service
	chat    chatports.ChatInterface
	emitter coreevents.Bus
	stream  *streamManager
}

// NewOrchestrator creates a chat orchestrator with required dependencies.
func NewOrchestrator(chatService *Service, completionService chatports.ChatInterface, emitter coreevents.Bus) *Orchestrator {

	return &Orchestrator{
		service: chatService,
		chat:    completionService,
		emitter: emitter,
		stream:  newStreamManager(),
	}
}

// CreateConversation creates a new conversation with the given settings.
func (o *Orchestrator) CreateConversation(providerName, model string) (*chatdomain.Conversation, error) {

	return o.service.CreateConversation(chatdomain.ConversationSettings{
		Provider: providerName,
		Model:    model,
	})
}

// SetActiveConversation sets the active conversation by ID.
func (o *Orchestrator) SetActiveConversation(id string) {

	o.service.SetActiveConversation(id)
}

// GetActiveConversation returns the currently active conversation.
func (o *Orchestrator) GetActiveConversation() *chatdomain.Conversation {

	id := o.service.ActiveConversationID()
	if id == "" {
		return nil
	}
	return o.service.GetConversation(id)
}

// GetConversation returns a conversation by ID.
func (o *Orchestrator) GetConversation(id string) *chatdomain.Conversation {

	return o.service.GetConversation(id)
}

// ListConversations returns summaries of all conversations.
func (o *Orchestrator) ListConversations() []chatdomain.ConversationSummary {

	return o.service.ListConversations()
}

// ListDeletedConversations returns summaries of archived conversations.
func (o *Orchestrator) ListDeletedConversations() []chatdomain.ConversationSummary {

	return o.service.ListDeletedConversations()
}

// UpdateConversationModel updates the model for a conversation.
func (o *Orchestrator) UpdateConversationModel(conversationID, model string) bool {

	return o.service.UpdateConversationModel(conversationID, model)
}

// UpdateConversationProvider updates the provider for a conversation.
func (o *Orchestrator) UpdateConversationProvider(conversationID, provider string) bool {

	return o.service.UpdateConversationProvider(conversationID, provider)
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
func (o *Orchestrator) SendMessage(ctx context.Context, conversationID, content string) (*chatdomain.Message, error) {

	conversationID = strings.TrimSpace(conversationID)
	content = strings.TrimSpace(content)
	if conversationID == "" {
		return nil, errors.New("conversation ID required")
	}
	if content == "" {
		return nil, errors.New("message content required")
	}

	conversation := o.service.GetConversation(conversationID)
	if conversation == nil {
		return nil, fmt.Errorf("conversation not found: %s", conversationID)
	}
	if conversation.IsArchived {
		return nil, fmt.Errorf("conversation archived: %s", conversationID)
	}

	userMsg := o.service.AddMessage(conversationID, chatdomain.RoleUser, content)
	if userMsg == nil {
		return nil, fmt.Errorf("failed to persist user message for conversation: %s", conversationID)
	}

	o.maybeAutoTitle(conversationID, userMsg)

	coreevents.Emit(o.emitter, SignalMessageCreated, MessageEventPayload{
		ConversationID: conversationID,
		MessageID:      userMsg.ID,
		Timestamp:      time.Now().UnixMilli(),
		Message:        userMsg,
	})

	conv := o.service.GetConversation(conversationID)
	if conv == nil {
		return userMsg, nil
	}

	providerName := strings.TrimSpace(conv.Settings.Provider)
	if providerName == "" {
		return userMsg, nil
	}

	streamMsg := o.service.CreateStreamingMessage(conversationID, chatdomain.RoleAssistant)
	if streamMsg == nil {
		return userMsg, nil
	}

	coreevents.Emit(o.emitter, SignalStreamStarted, MessageEventPayload{
		ConversationID: conversationID,
		MessageID:      streamMsg.ID,
		Timestamp:      time.Now().UnixMilli(),
		Message:        streamMsg,
	})

	if o.chat == nil {
		err := fmt.Errorf("chat service not configured")
		o.emitStreamError(conversationID, streamMsg.ID, err)
		metadata := o.buildMetadata(providerName, conv.Settings.Model, "error", nil, time.Now(), err)
		_ = o.service.FinalizeMessage(conversationID, streamMsg.ID, metadata)
		return userMsg, nil
	}

	chatRequest := chatports.ChatRequest{
		ProviderName: providerName,
		ModelName:    conv.Settings.Model,
		Messages:     o.buildChatMessages(conv, streamMsg.ID),
		Options: chatports.ChatOptions{
			Temperature: conv.Settings.Temperature,
			MaxTokens:   conv.Settings.MaxTokens,
			Stream:      true,
		},
	}

	ctx, cancel := context.WithCancel(ctx)
	o.stream.start(conversationID, streamMsg.ID, cancel)

	chunks, err := o.chat.Chat(ctx, chatRequest)
	if err != nil {
		o.stream.clear(conversationID, streamMsg.ID)
		o.emitStreamError(conversationID, streamMsg.ID, err)
		metadata := o.buildMetadata(providerName, conv.Settings.Model, "error", nil, time.Now(), err)
		_ = o.service.FinalizeMessage(conversationID, streamMsg.ID, metadata)
		return userMsg, nil
	}

	go o.consumeStream(conversationID, streamMsg.ID, providerName, conv.Settings.Model, chunks)

	return userMsg, nil
}

// StopStream cancels the currently running stream.
func (o *Orchestrator) StopStream() {

	o.stream.stop()
}

// emitStreamChunk publishes a streaming chunk event.
func (o *Orchestrator) emitStreamChunk(conversationID, messageID string, blockIndex int, content string) {

	coreevents.Emit(o.emitter, SignalStreamChunk, StreamChunkEventPayload{
		ConversationID: conversationID,
		MessageID:      messageID,
		Timestamp:      time.Now().UnixMilli(),
		BlockIndex:     blockIndex,
		Content:        content,
		IsDone:         false,
	})
}

// emitStreamError publishes a streaming error event.
func (o *Orchestrator) emitStreamError(conversationID, messageID string, err error) {

	coreevents.Emit(o.emitter, SignalStreamError, StreamChunkEventPayload{
		ConversationID: conversationID,
		MessageID:      messageID,
		Timestamp:      time.Now().UnixMilli(),
		BlockIndex:     0,
		Content:        "",
		IsDone:         true,
		Error:          err.Error(),
		StatusCode:     chatdomain.StatusCodeFromErr(err),
	})
}

// emitStreamComplete publishes a stream completion event.
func (o *Orchestrator) emitStreamComplete(conversationID, messageID string, metadata *chatdomain.MessageMetadata) {

	coreevents.Emit(o.emitter, SignalStreamCompleted, StreamChunkEventPayload{
		ConversationID: conversationID,
		MessageID:      messageID,
		Timestamp:      time.Now().UnixMilli(),
		BlockIndex:     0,
		Content:        "",
		IsDone:         true,
		Metadata:       metadata,
	})
}

// buildChatMessages builds the chat request message list.
func (o *Orchestrator) buildChatMessages(conv *chatdomain.Conversation, streamingMessageID string) []chatports.ChatMessage {

	conv.Lock()
	defer conv.Unlock()

	messages := make([]chatports.ChatMessage, 0, len(conv.Messages)+1)
	systemPrompt := strings.TrimSpace(conv.Settings.SystemPrompt)
	if systemPrompt != "" {
		messages = append(messages, chatports.ChatMessage{
			Role:    chatports.ChatRoleSystem,
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
		messages = append(messages, chatports.ChatMessage{
			Role:    chatports.ChatRole(msg.Role),
			Content: content,
		})
	}
	return messages
}

// textFromBlocks builds a text-only content string from message blocks.
func textFromBlocks(blocks []chatdomain.Block) string {

	if len(blocks) == 0 {
		return ""
	}

	var builder strings.Builder
	for _, block := range blocks {
		if block.Type == chatdomain.BlockTypeText {
			builder.WriteString(block.Content)
		}
	}
	return builder.String()
}

// consumeStream handles incoming chat chunks and emits events.
func (o *Orchestrator) consumeStream(conversationID, messageID, providerName, fallbackModel string, chunks <-chan chatports.ChatChunk) {

	defer o.stream.clear(conversationID, messageID)

	start := time.Now()
	var (
		finishReason string
		usage        *chatports.ChatUsage
		model        string
	)

	for chunk := range chunks {
		if chunk.Error != "" {
			chunkErr := errors.New(chunk.Error)
			if isContextCanceledMessage(chunk.Error) {
				metadata := o.buildMetadata(providerName, chooseModel(model, fallbackModel), "cancelled", usage, start, nil)
				_ = o.service.FinalizeMessage(conversationID, messageID, metadata)
				o.emitStreamComplete(conversationID, messageID, metadata)
			} else {
				o.emitStreamError(conversationID, messageID, chunkErr)
				metadata := o.buildMetadata(providerName, chooseModel(model, fallbackModel), "error", usage, start, chunkErr)
				_ = o.service.FinalizeMessage(conversationID, messageID, metadata)
			}
			return
		}

		if chunk.Content != "" {
			if !o.service.AppendToMessage(conversationID, messageID, 0, chunk.Content) {
				err := fmt.Errorf("failed to persist stream chunk")
				o.emitStreamError(conversationID, messageID, err)
				metadata := o.buildMetadata(providerName, chooseModel(model, fallbackModel), "error", usage, start, err)
				_ = o.service.FinalizeMessage(conversationID, messageID, metadata)
				return
			}
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
	if !o.service.FinalizeMessage(conversationID, messageID, metadata) {
		err := fmt.Errorf("failed to persist stream completion")
		o.emitStreamError(conversationID, messageID, err)
		return
	}
	o.emitStreamComplete(conversationID, messageID, metadata)
}

// buildMetadata builds message metadata from provider results.
func (o *Orchestrator) buildMetadata(
	providerName,
	model,
	finishReason string,
	usage *chatports.ChatUsage,
	start time.Time,
	err error,
) *chatdomain.MessageMetadata {

	meta := &chatdomain.MessageMetadata{
		Provider:     providerName,
		Model:        model,
		FinishReason: finishReason,
		LatencyMs:    time.Since(start).Milliseconds(),
	}
	if meta.FinishReason == "" {
		meta.FinishReason = "stop"
	}
	if usage != nil {
		meta.TokensIn = usage.InputTokens
		meta.TokensOut = usage.OutputTokens
		meta.TokensTotal = usage.TotalTokens
		if meta.TokensTotal == 0 {
			meta.TokensTotal = usage.InputTokens + usage.OutputTokens
		}
	}
	if err != nil {
		meta.StatusCode = chatdomain.StatusCodeFromErr(err)
		meta.ErrorMessage = err.Error()
	}
	return meta
}

// maybeAutoTitle updates the conversation title on the first user message.
func (o *Orchestrator) maybeAutoTitle(conversationID string, message *chatdomain.Message) {

	if message.Role != chatdomain.RoleUser {
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
		coreevents.Emit(o.emitter, SignalConversationTitle, ConversationTitleEventPayload{
			ConversationID: conversationID,
			Timestamp:      time.Now().UnixMilli(),
			Title:          content,
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

// isContextCanceledMessage checks whether an error string represents cancellation.
func isContextCanceledMessage(message string) bool {

	lower := strings.ToLower(strings.TrimSpace(message))
	if lower == "" {
		return false
	}
	return strings.Contains(lower, context.Canceled.Error()) ||
		strings.Contains(lower, context.DeadlineExceeded.Error())
}
