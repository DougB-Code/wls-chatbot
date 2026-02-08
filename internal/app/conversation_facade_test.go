// conversation_facade_test.go validates conversation facade contract mapping behavior.
// internal/app/conversation_facade_test.go
package app

import (
	"testing"

	chatfeature "github.com/MadeByDoug/wls-chatbot/internal/features/ai/chat/app/chat"
)

// TestMapConversationCopiesConversationAndNestedFields validates conversation and nested field mapping.
func TestMapConversationCopiesConversationAndNestedFields(t *testing.T) {

	source := &chatfeature.Conversation{
		ID:    "conv-1",
		Title: "Conversation",
		Messages: []*chatfeature.Message{
			{
				ID:             "msg-1",
				ConversationID: "conv-1",
				Role:           chatfeature.RoleAssistant,
				Blocks: []chatfeature.Block{
					{
						Type:    chatfeature.BlockTypeText,
						Content: "hello",
					},
					{
						Type: chatfeature.BlockTypeAction,
						Action: &chatfeature.ActionExecution{
							ID:          "action-1",
							ToolName:    "tool",
							Description: "desc",
							Args: map[string]interface{}{
								"x": "y",
							},
							Status: chatfeature.ActionStatusCompleted,
						},
					},
				},
				Metadata: &chatfeature.MessageMetadata{
					Provider: "openai",
					Model:    "gpt-4.1",
				},
			},
		},
		Settings: chatfeature.ConversationSettings{
			Provider:     "openai",
			Model:        "gpt-4.1",
			Temperature:  0.7,
			MaxTokens:    123,
			SystemPrompt: "be concise",
		},
		CreatedAt:  1,
		UpdatedAt:  2,
		IsArchived: false,
	}

	mapped := mapConversation(source)
	if mapped == nil {
		t.Fatalf("expected mapped conversation")
	}

	if mapped.ID != source.ID || mapped.Title != source.Title {
		t.Fatalf("expected top-level fields to match")
	}
	if mapped.Settings.Provider != source.Settings.Provider || mapped.Settings.Model != source.Settings.Model {
		t.Fatalf("expected settings to match")
	}
	if len(mapped.Messages) != 1 {
		t.Fatalf("expected one message, got %d", len(mapped.Messages))
	}

	if mapped.Messages[0].Metadata == nil || mapped.Messages[0].Metadata.Provider != "openai" {
		t.Fatalf("expected mapped metadata")
	}

	action := mapped.Messages[0].Blocks[1].Action
	if action == nil || action.Args["x"] != "y" {
		t.Fatalf("expected mapped action args")
	}

	source.Messages[0].Blocks[1].Action.Args["x"] = "changed"
	if action.Args["x"] != "y" {
		t.Fatalf("expected mapped args to be deep-copied")
	}
	source.Messages[0].Metadata.Provider = "changed"
	if mapped.Messages[0].Metadata.Provider != "openai" {
		t.Fatalf("expected mapped metadata to be copied")
	}
}

// TestMapConversationSummariesHandlesEmpty validates empty summary mapping behavior.
func TestMapConversationSummariesHandlesEmpty(t *testing.T) {

	if mapped := mapConversationSummaries(nil); mapped != nil {
		t.Fatalf("expected nil mapping for nil input")
	}
	if mapped := mapConversationSummaries([]chatfeature.ConversationSummary{}); mapped != nil {
		t.Fatalf("expected nil mapping for empty input")
	}
}
