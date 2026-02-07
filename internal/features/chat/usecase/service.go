// manage conversation lifecycle and message state mutations.
// internal/features/chat/usecase/service.go
package chat

import (
	"sync"
	"time"
)

// Service manages conversations and messages.
type Service struct {
	repo Repository

	mu           sync.RWMutex
	activeConvID string
}

// NewService creates a new chat service with the provided repository.
func NewService(repo Repository) *Service {

	return &Service{
		repo: repo,
	}
}

// CreateConversation creates a new conversation with the given settings.
func (s *Service) CreateConversation(settings ConversationSettings) (*Conversation, error) {

	conv := NewConversation(settings)
	if err := s.repo.Create(conv); err != nil {
		return nil, err
	}
	s.SetActiveConversation(conv.ID)
	return conv, nil
}

// SetActiveConversation marks a conversation as active.
func (s *Service) SetActiveConversation(id string) {

	s.mu.Lock()
	s.activeConvID = id
	s.mu.Unlock()
}

// ActiveConversationID returns the active conversation ID.
func (s *Service) ActiveConversationID() string {

	s.mu.RLock()
	id := s.activeConvID
	s.mu.RUnlock()
	return id
}

// GetConversation retrieves a conversation by ID.
func (s *Service) GetConversation(id string) *Conversation {

	conv, err := s.repo.Get(id)
	if err != nil {
		return nil
	}
	if conv == nil {
		return nil
	}
	return conv.Snapshot()
}

// ListConversations returns summaries of all conversations, sorted by updatedAt.
func (s *Service) ListConversations() []ConversationSummary {

	return s.listConversationsByArchivedStatus(false)
}

// ListDeletedConversations returns summaries for archived conversations.
func (s *Service) ListDeletedConversations() []ConversationSummary {

	return s.listConversationsByArchivedStatus(true)
}

// listConversationsByArchivedStatus returns summaries filtered by archive status.
func (s *Service) listConversationsByArchivedStatus(archived bool) []ConversationSummary {

	convs, err := s.repo.List()
	if err != nil {
		return nil
	}
	summaries := make([]ConversationSummary, 0, len(convs))
	for _, conv := range convs {
		if conv.CheckIsArchived() == archived {
			summaries = append(summaries, conv.GetSummary())
		}
	}

	// Sort by updatedAt descending
	for i := 0; i < len(summaries)-1; i++ {
		for j := i + 1; j < len(summaries); j++ {
			if summaries[j].UpdatedAt > summaries[i].UpdatedAt {
				summaries[i], summaries[j] = summaries[j], summaries[i]
			}
		}
	}

	return summaries
}

// AddMessage adds a message to a conversation.
func (s *Service) AddMessage(conversationID string, role Role, content string) *Message {

	conv, err := s.repo.Get(conversationID)
	if err != nil {
		return nil
	}
	if conv == nil || conv.CheckIsArchived() {
		return nil
	}

	msg := NewMessage(conversationID, role, content)
	conv.AddMessage(msg)
	if err := s.repo.Update(conv); err != nil {
		return nil
	}
	return msg
}

// SetConversationTitle updates the title for a conversation.
func (s *Service) SetConversationTitle(conversationID, title string) bool {

	conv, err := s.repo.Get(conversationID)
	if err != nil {
		return false
	}
	if conv == nil {
		return false
	}

	conv.Lock()
	conv.Title = title
	conv.UpdatedAt = time.Now().UnixMilli()
	conv.Unlock()

	return s.repo.Update(conv) == nil
}

// CreateStreamingMessage creates a new streaming message placeholder.
func (s *Service) CreateStreamingMessage(conversationID string, role Role) *Message {

	conv, err := s.repo.Get(conversationID)
	if err != nil {
		return nil
	}
	if conv == nil || conv.CheckIsArchived() {
		return nil
	}

	msg := NewStreamingMessage(conversationID, role)
	conv.AddMessage(msg)
	if err := s.repo.Update(conv); err != nil {
		return nil
	}
	return msg
}

// AppendToMessage appends content to a streaming message.
func (s *Service) AppendToMessage(conversationID, messageID string, blockIndex int, content string) bool {

	conv, err := s.repo.Get(conversationID)
	if err != nil {
		return false
	}
	if conv == nil {
		return false
	}

	conv.Lock()
	defer conv.Unlock()

	updated := false
	for _, msg := range conv.Messages {
		if msg.ID == messageID {
			// Extend blocks if needed
			for len(msg.Blocks) <= blockIndex {
				msg.Blocks = append(msg.Blocks, Block{Type: BlockTypeText})
			}
			msg.Blocks[blockIndex].Content += content
			updated = true
			break
		}
	}

	if updated {
		if err := s.repo.Update(conv); err != nil {
			return false
		}
	}

	return updated
}

// FinalizeMessage marks a streaming message as complete.
func (s *Service) FinalizeMessage(conversationID, messageID string, metadata *MessageMetadata) bool {

	conv, err := s.repo.Get(conversationID)
	if err != nil {
		return false
	}
	if conv == nil {
		return false
	}

	conv.Lock()
	defer conv.Unlock()

	updated := false
	for _, msg := range conv.Messages {
		if msg.ID == messageID {
			msg.IsStreaming = false
			msg.Metadata = metadata
			updated = true
			break
		}
	}

	if updated {
		if err := s.repo.Update(conv); err != nil {
			return false
		}
	}

	return updated
}

// DeleteConversation moves a conversation into the recycle bin.
func (s *Service) DeleteConversation(id string) bool {

	conv, err := s.repo.Get(id)
	if err != nil {
		return false
	}
	if conv == nil || conv.CheckIsArchived() {
		return false
	}

	conv.Lock()
	conv.IsArchived = true
	conv.UpdatedAt = time.Now().UnixMilli()
	conv.Unlock()

	if s.ActiveConversationID() == id {
		s.SetActiveConversation("")
	}

	return s.repo.Update(conv) == nil
}

// RestoreConversation restores a conversation from the recycle bin.
func (s *Service) RestoreConversation(id string) bool {

	conv, err := s.repo.Get(id)
	if err != nil {
		return false
	}
	if conv == nil || !conv.CheckIsArchived() {
		return false
	}

	conv.Lock()
	conv.IsArchived = false
	conv.UpdatedAt = time.Now().UnixMilli()
	conv.Unlock()

	return s.repo.Update(conv) == nil
}

// PurgeConversation permanently deletes a conversation.
func (s *Service) PurgeConversation(id string) bool {

	if s.ActiveConversationID() == id {
		s.SetActiveConversation("")
	}

	err := s.repo.Delete(id)
	return err == nil
}

// UpdateConversationModel updates the model for a conversation.
func (s *Service) UpdateConversationModel(id, model string) bool {

	if model == "" {
		return false
	}
	conv, err := s.repo.Get(id)
	if err != nil {
		return false
	}
	if conv != nil && !conv.CheckIsArchived() {
		conv.Lock()
		defer conv.Unlock()
		conv.Settings.Model = model
		conv.UpdatedAt = time.Now().UnixMilli()
		return s.repo.Update(conv) == nil
	}
	return false
}

// UpdateConversationProvider updates the provider for a conversation.
func (s *Service) UpdateConversationProvider(id, provider string) bool {

	if provider == "" {
		return false
	}
	conv, err := s.repo.Get(id)
	if err != nil {
		return false
	}
	if conv != nil && !conv.CheckIsArchived() {
		conv.Lock()
		defer conv.Unlock()
		conv.Settings.Provider = provider
		conv.UpdatedAt = time.Now().UnixMilli()
		return s.repo.Update(conv) == nil
	}
	return false
}
