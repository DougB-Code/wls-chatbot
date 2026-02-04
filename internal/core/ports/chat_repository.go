// define chat conversation persistence contracts.
// internal/core/ports/chat_repository.go
package ports

import "github.com/MadeByDoug/wls-chatbot/internal/core/domain/chat"

// ChatRepository defines storage operations for conversations.
type ChatRepository interface {
	Create(conv *chat.Conversation) error
	Get(id string) (*chat.Conversation, error)
	List() ([]*chat.Conversation, error)
	Update(conv *chat.Conversation) error
	Delete(id string) error
}
