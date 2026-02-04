// define chat conversation persistence contracts.
// internal/features/chat/ports/chat_repository.go
package ports

import chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/chat/domain"

// ChatRepository defines storage operations for conversations.
type ChatRepository interface {
	Create(conv *chatdomain.Conversation) error
	Get(id string) (*chatdomain.Conversation, error)
	List() ([]*chatdomain.Conversation, error)
	Update(conv *chatdomain.Conversation) error
	Delete(id string) error
}
