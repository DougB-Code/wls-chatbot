// define conversation state and behavior for chat sessions.
// internal/features/chat/domain/conversation.go
package chat

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"sync"
	"time"
)

// ConversationSettings holds the configuration for a conversation.
type ConversationSettings struct {
	Provider     string  `json:"provider"`
	Model        string  `json:"model"`
	Temperature  float64 `json:"temperature,omitempty"`
	MaxTokens    int     `json:"maxTokens,omitempty"`
	SystemPrompt string  `json:"systemPrompt,omitempty"`
}

// Conversation represents a chat conversation.
type Conversation struct {
	mu         sync.RWMutex
	ID         string               `json:"id"`
	Title      string               `json:"title"`
	Messages   []*Message           `json:"messages"`
	Settings   ConversationSettings `json:"settings"`
	CreatedAt  int64                `json:"createdAt"`
	UpdatedAt  int64                `json:"updatedAt"`
	IsArchived bool                 `json:"isArchived"`
}

// ConversationSummary is a lightweight representation for listing.
type ConversationSummary struct {
	ID           string `json:"id"`
	Title        string `json:"title"`
	LastMessage  string `json:"lastMessage,omitempty"`
	MessageCount int    `json:"messageCount"`
	UpdatedAt    int64  `json:"updatedAt"`
}

// NewConversation creates a new conversation with the given settings.
func NewConversation(settings ConversationSettings) *Conversation {

	now := time.Now().UnixMilli()
	return &Conversation{
		ID:        generateID(),
		Title:     "New conversation",
		Messages:  []*Message{},
		Settings:  settings,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

// Lock locks the conversation for writes.
func (c *Conversation) Lock() {

	c.mu.Lock()
}

// Unlock releases the conversation write lock.
func (c *Conversation) Unlock() {

	c.mu.Unlock()
}

// CheckIsArchived returns the archived status safely.
func (c *Conversation) CheckIsArchived() bool {

	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.IsArchived
}

// AddMessage adds a message to the conversation.
func (c *Conversation) AddMessage(msg *Message) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.Messages = append(c.Messages, msg)
	c.UpdatedAt = time.Now().UnixMilli()
}

// GetSummary returns a lightweight summary of the conversation.
func (c *Conversation) GetSummary() ConversationSummary {

	c.mu.RLock()
	defer c.mu.RUnlock()

	summary := ConversationSummary{
		ID:           c.ID,
		Title:        c.Title,
		MessageCount: len(c.Messages),
		UpdatedAt:    c.UpdatedAt,
	}

	if len(c.Messages) > 0 {
		lastMsg := c.Messages[len(c.Messages)-1]
		if len(lastMsg.Blocks) > 0 {
			content := lastMsg.Blocks[0].Content
			if len(content) > 100 {
				summary.LastMessage = content[:100]
			} else {
				summary.LastMessage = content
			}
		}
	}

	return summary
}

// Snapshot returns a deep copy of the conversation for safe reads.
func (c *Conversation) Snapshot() *Conversation {

	if c == nil {
		return nil
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	return &Conversation{
		ID:         c.ID,
		Title:      c.Title,
		Messages:   cloneMessages(c.Messages),
		Settings:   c.Settings,
		CreatedAt:  c.CreatedAt,
		UpdatedAt:  c.UpdatedAt,
		IsArchived: c.IsArchived,
	}
}

// generateID creates a random ID.
func generateID() string {

	b := make([]byte, 16)
	read, err := rand.Read(b)
	if err != nil || read != len(b) {
		now := time.Now().UnixNano()
		binary.LittleEndian.PutUint64(b[:8], uint64(now))
		binary.LittleEndian.PutUint64(b[8:], uint64(now>>1))
	}
	return hex.EncodeToString(b)
}

// cloneMessages deep copies message pointers for snapshots.
func cloneMessages(messages []*Message) []*Message {

	if messages == nil {
		return nil
	}

	cloned := make([]*Message, len(messages))
	for i, message := range messages {
		cloned[i] = cloneMessage(message)
	}
	return cloned
}

// cloneMessage deep copies a message for snapshots.
func cloneMessage(message *Message) *Message {

	if message == nil {
		return nil
	}

	return &Message{
		ID:             message.ID,
		ConversationID: message.ConversationID,
		Role:           message.Role,
		Blocks:         cloneBlocks(message.Blocks),
		Timestamp:      message.Timestamp,
		IsStreaming:    message.IsStreaming,
		Metadata:       cloneMetadata(message.Metadata),
	}
}

// cloneBlocks deep copies blocks for snapshots.
func cloneBlocks(blocks []Block) []Block {

	if blocks == nil {
		return nil
	}

	cloned := make([]Block, len(blocks))
	for i, block := range blocks {
		cloned[i] = Block{
			Type:        block.Type,
			Content:     block.Content,
			Language:    block.Language,
			Artifact:    cloneArtifact(block.Artifact),
			Action:      cloneAction(block.Action),
			IsCollapsed: block.IsCollapsed,
		}
	}
	return cloned
}

// cloneArtifact deep copies an artifact for snapshots.
func cloneArtifact(artifact *Artifact) *Artifact {

	if artifact == nil {
		return nil
	}

	clone := *artifact
	return &clone
}

// cloneAction deep copies an action execution for snapshots.
func cloneAction(action *ActionExecution) *ActionExecution {

	if action == nil {
		return nil
	}

	clone := *action
	if action.Args != nil {
		clone.Args = make(map[string]interface{}, len(action.Args))
		for key, value := range action.Args {
			clone.Args[key] = value
		}
	}
	return &clone
}

// cloneMetadata deep copies message metadata for snapshots.
func cloneMetadata(metadata *MessageMetadata) *MessageMetadata {

	if metadata == nil {
		return nil
	}

	clone := *metadata
	return &clone
}
