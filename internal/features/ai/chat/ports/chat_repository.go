// chat_repository.go defines chat persistence ports.
// internal/features/ai/chat/ports/chat_repository.go
package ports

import chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/domain"

// ChatRepository defines storage operations for conversations.
type ChatRepository interface {
	Create(conv *chatdomain.Conversation) error
	Get(id string) (*chatdomain.Conversation, error)
	List() ([]*chatdomain.Conversation, error)
	Update(conv *chatdomain.Conversation) error
	Delete(id string) error
}
