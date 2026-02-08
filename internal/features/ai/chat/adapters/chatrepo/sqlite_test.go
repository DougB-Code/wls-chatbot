// sqlite_test.go verifies chat repository persistence behavior.
// internal/features/chat/adapters/chatrepo/sqlite_test.go
package chatrepo

import (
	"path/filepath"
	"testing"

	"github.com/MadeByDoug/wls-chatbot/internal/core/datastore"
	chatcore "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/domain"
)

// TestRepositoryCreateAndGetRoundTrip verifies normalized persistence and retrieval.
func TestRepositoryCreateAndGetRoundTrip(t *testing.T) {

	repo := newTestRepository(t)
	conv := &chatcore.Conversation{
		ID:    "conv-1",
		Title: "Launch Plan",
		Settings: chatcore.ConversationSettings{
			Provider:     "openai",
			Model:        "gpt-4o",
			Temperature:  0.2,
			MaxTokens:    2048,
			SystemPrompt: "Be concise",
		},
		CreatedAt:  1,
		UpdatedAt:  2,
		IsArchived: false,
		Messages: []*chatcore.Message{
			{
				ID:             "msg-user",
				ConversationID: "conv-1",
				Role:           chatcore.RoleUser,
				Blocks: []chatcore.Block{
					{Type: chatcore.BlockTypeText, Content: "Hello"},
				},
				Timestamp: 10,
			},
			{
				ID:             "msg-assistant",
				ConversationID: "conv-1",
				Role:           chatcore.RoleAssistant,
				Blocks: []chatcore.Block{
					{Type: chatcore.BlockTypeText, Content: "Hi there"},
				},
				Timestamp: 11,
				Metadata: &chatcore.MessageMetadata{
					Provider:     "openai",
					Model:        "gpt-4o",
					TokensIn:     12,
					TokensOut:    34,
					TokensTotal:  46,
					LatencyMs:    123,
					FinishReason: "stop",
				},
			},
		},
	}

	if err := repo.Create(conv); err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	loaded, err := repo.Get("conv-1")
	if err != nil {
		t.Fatalf("get conversation: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected conversation, got nil")
	}
	if loaded.Title != conv.Title {
		t.Fatalf("expected title %q, got %q", conv.Title, loaded.Title)
	}
	if loaded.Settings.Provider != "openai" || loaded.Settings.Model != "gpt-4o" {
		t.Fatalf("unexpected settings: %+v", loaded.Settings)
	}
	if len(loaded.Messages) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(loaded.Messages))
	}
	if loaded.Messages[0].Blocks[0].Content != "Hello" {
		t.Fatalf("expected first message content to round-trip")
	}
	metadata := loaded.Messages[1].Metadata
	if metadata == nil {
		t.Fatalf("expected metadata for assistant message")
	}
	if metadata.Provider != "openai" || metadata.Model != "gpt-4o" {
		t.Fatalf("unexpected metadata provider/model: %+v", metadata)
	}
	if metadata.TokensTotal != 46 {
		t.Fatalf("expected total tokens 46, got %d", metadata.TokensTotal)
	}
	if metadata.LatencyMs != 123 {
		t.Fatalf("expected latency 123, got %d", metadata.LatencyMs)
	}
}

// TestRepositoryUpdateReplacesMessages verifies update rewrites message rows for a conversation.
func TestRepositoryUpdateReplacesMessages(t *testing.T) {

	repo := newTestRepository(t)
	conv := &chatcore.Conversation{
		ID:    "conv-2",
		Title: "Draft",
		Settings: chatcore.ConversationSettings{
			Provider: "openai",
			Model:    "gpt-4o",
		},
		CreatedAt: 1,
		UpdatedAt: 2,
		Messages: []*chatcore.Message{
			{
				ID:             "msg-old",
				ConversationID: "conv-2",
				Role:           chatcore.RoleUser,
				Blocks: []chatcore.Block{
					{Type: chatcore.BlockTypeText, Content: "Old"},
				},
				Timestamp: 1,
			},
		},
	}
	if err := repo.Create(conv); err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	conv.Title = "Updated"
	conv.UpdatedAt = 3
	conv.Messages = []*chatcore.Message{
		{
			ID:             "msg-new",
			ConversationID: "conv-2",
			Role:           chatcore.RoleAssistant,
			Blocks: []chatcore.Block{
				{Type: chatcore.BlockTypeError, Content: "provider unavailable"},
			},
			Timestamp: 3,
			Metadata: &chatcore.MessageMetadata{
				Provider:     "openai",
				Model:        "gpt-4o-mini",
				FinishReason: "error",
				StatusCode:   503,
				ErrorMessage: "provider unavailable",
			},
		},
	}
	if err := repo.Update(conv); err != nil {
		t.Fatalf("update conversation: %v", err)
	}

	loaded, err := repo.Get("conv-2")
	if err != nil {
		t.Fatalf("get updated conversation: %v", err)
	}
	if loaded == nil {
		t.Fatalf("expected conversation after update")
	}
	if loaded.Title != "Updated" {
		t.Fatalf("expected updated title, got %q", loaded.Title)
	}
	if len(loaded.Messages) != 1 {
		t.Fatalf("expected one replaced message, got %d", len(loaded.Messages))
	}
	if loaded.Messages[0].ID != "msg-new" {
		t.Fatalf("expected message replacement, got %q", loaded.Messages[0].ID)
	}
	if loaded.Messages[0].Metadata == nil || loaded.Messages[0].Metadata.StatusCode != 503 {
		t.Fatalf("expected persisted status code metadata, got %+v", loaded.Messages[0].Metadata)
	}
}

// TestRepositoryDeleteRemovesConversation verifies hard deletion behavior.
func TestRepositoryDeleteRemovesConversation(t *testing.T) {

	repo := newTestRepository(t)
	conv := &chatcore.Conversation{
		ID:    "conv-3",
		Title: "To delete",
		Settings: chatcore.ConversationSettings{
			Provider: "openai",
			Model:    "gpt-4o",
		},
		CreatedAt: 1,
		UpdatedAt: 1,
	}
	if err := repo.Create(conv); err != nil {
		t.Fatalf("create conversation: %v", err)
	}

	if err := repo.Delete("conv-3"); err != nil {
		t.Fatalf("delete conversation: %v", err)
	}

	loaded, err := repo.Get("conv-3")
	if err != nil {
		t.Fatalf("get deleted conversation: %v", err)
	}
	if loaded != nil {
		t.Fatalf("expected nil after delete, got %+v", loaded)
	}
}

// newTestRepository creates an isolated SQLite-backed repository.
func newTestRepository(t *testing.T) *Repository {

	t.Helper()
	path := filepath.Join(t.TempDir(), "chatrepo.db")
	db, err := datastore.OpenSQLite(path)
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	repo, err := NewRepository(db)
	if err != nil {
		t.Fatalf("new repository: %v", err)
	}

	return repo
}
