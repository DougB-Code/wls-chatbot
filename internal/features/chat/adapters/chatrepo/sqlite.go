// sqlite.go persists chat conversations with a normalized SQLite schema.
// internal/features/chat/adapters/chatrepo/sqlite.go
package chatrepo

import (
	"database/sql"
	"errors"
	"fmt"

	chatdomain "github.com/MadeByDoug/wls-chatbot/internal/features/chat/domain"
	chatports "github.com/MadeByDoug/wls-chatbot/internal/features/chat/ports"
)

const chatSchema = `
CREATE TABLE IF NOT EXISTS chat_conversations (
	id TEXT PRIMARY KEY,
	title TEXT NOT NULL,
	provider TEXT NOT NULL,
	model TEXT NOT NULL,
	temperature REAL NOT NULL,
	max_tokens INTEGER NOT NULL,
	system_prompt TEXT NOT NULL,
	created_at INTEGER NOT NULL,
	updated_at INTEGER NOT NULL,
	is_archived INTEGER NOT NULL CHECK (is_archived IN (0, 1))
);

CREATE TABLE IF NOT EXISTS chat_messages (
	id TEXT PRIMARY KEY,
	conversation_id TEXT NOT NULL,
	role TEXT NOT NULL,
	timestamp INTEGER NOT NULL,
	is_streaming INTEGER NOT NULL CHECK (is_streaming IN (0, 1)),
	provider TEXT,
	model TEXT,
	tokens_in INTEGER,
	tokens_out INTEGER,
	tokens_total INTEGER,
	latency_ms INTEGER,
	finish_reason TEXT,
	status_code INTEGER,
	error_message TEXT,
	FOREIGN KEY (conversation_id) REFERENCES chat_conversations(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_chat_messages_conversation_ts
ON chat_messages (conversation_id, timestamp DESC);

CREATE TABLE IF NOT EXISTS chat_message_blocks (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	message_id TEXT NOT NULL,
	block_index INTEGER NOT NULL,
	block_type TEXT NOT NULL,
	content TEXT NOT NULL,
	language TEXT,
	is_collapsed INTEGER NOT NULL CHECK (is_collapsed IN (0, 1)),
	artifact_id TEXT,
	artifact_name TEXT,
	artifact_type TEXT,
	artifact_content TEXT,
	artifact_language TEXT,
	artifact_version INTEGER,
	artifact_created_at INTEGER,
	artifact_updated_at INTEGER,
	action_id TEXT,
	action_tool_name TEXT,
	action_description TEXT,
	action_status TEXT,
	action_result TEXT,
	action_started_at INTEGER,
	action_completed_at INTEGER,
	FOREIGN KEY (message_id) REFERENCES chat_messages(id) ON DELETE CASCADE
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_chat_message_blocks_order
ON chat_message_blocks (message_id, block_index);
`

// Repository stores conversations in SQLite.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a SQLite-backed chat repository.
func NewRepository(db *sql.DB) (*Repository, error) {

	if db == nil {
		return nil, fmt.Errorf("chat repo: db required")
	}
	if _, err := db.Exec(chatSchema); err != nil {
		return nil, fmt.Errorf("chat repo: ensure schema: %w", err)
	}

	return &Repository{db: db}, nil
}

var _ chatports.ChatRepository = (*Repository)(nil)

// Create saves a new conversation.
func (r *Repository) Create(conv *chatdomain.Conversation) error {

	if err := r.validateConversation(conv); err != nil {
		return err
	}

	return withTx(r.db, func(tx *sql.Tx) error {
		if err := insertConversation(tx, conv); err != nil {
			return err
		}
		if err := replaceMessages(tx, conv.ID, conv.Messages); err != nil {
			return err
		}
		return nil
	})
}

// Get returns a conversation by ID.
func (r *Repository) Get(id string) (*chatdomain.Conversation, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("chat repo: db required")
	}
	if id == "" {
		return nil, nil
	}

	var conv chatdomain.Conversation
	var isArchived int
	err := r.db.QueryRow(
		`SELECT id, title, provider, model, temperature, max_tokens, system_prompt, created_at, updated_at, is_archived
		 FROM chat_conversations
		 WHERE id = ?`,
		id,
	).Scan(
		&conv.ID,
		&conv.Title,
		&conv.Settings.Provider,
		&conv.Settings.Model,
		&conv.Settings.Temperature,
		&conv.Settings.MaxTokens,
		&conv.Settings.SystemPrompt,
		&conv.CreatedAt,
		&conv.UpdatedAt,
		&isArchived,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("chat repo: get conversation: %w", err)
	}
	conv.IsArchived = isArchived == 1

	messages, err := loadMessages(r.db, conv.ID)
	if err != nil {
		return nil, err
	}
	conv.Messages = messages

	return &conv, nil
}

// List returns all conversations.
func (r *Repository) List() ([]*chatdomain.Conversation, error) {

	if r == nil || r.db == nil {
		return nil, fmt.Errorf("chat repo: db required")
	}

	rows, err := r.db.Query(
		`SELECT id, title, provider, model, temperature, max_tokens, system_prompt, created_at, updated_at, is_archived
		 FROM chat_conversations
		 ORDER BY updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("chat repo: list conversations: %w", err)
	}

	conversations := make([]*chatdomain.Conversation, 0)
	for rows.Next() {
		conv := &chatdomain.Conversation{}
		var isArchived int
		if err := rows.Scan(
			&conv.ID,
			&conv.Title,
			&conv.Settings.Provider,
			&conv.Settings.Model,
			&conv.Settings.Temperature,
			&conv.Settings.MaxTokens,
			&conv.Settings.SystemPrompt,
			&conv.CreatedAt,
			&conv.UpdatedAt,
			&isArchived,
		); err != nil {
			return nil, fmt.Errorf("chat repo: list scan conversation: %w", err)
		}
		conv.IsArchived = isArchived == 1
		conversations = append(conversations, conv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat repo: list rows: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("chat repo: close rows: %w", err)
	}

	for _, conv := range conversations {
		messages, err := loadMessages(r.db, conv.ID)
		if err != nil {
			return nil, err
		}
		conv.Messages = messages
	}

	return conversations, nil
}

// Update persists a conversation update.
func (r *Repository) Update(conv *chatdomain.Conversation) error {

	if err := r.validateConversation(conv); err != nil {
		return err
	}

	return withTx(r.db, func(tx *sql.Tx) error {
		if err := upsertConversation(tx, conv); err != nil {
			return err
		}
		if err := replaceMessages(tx, conv.ID, conv.Messages); err != nil {
			return err
		}
		return nil
	})
}

// Delete removes a conversation by ID.
func (r *Repository) Delete(id string) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("chat repo: db required")
	}
	if id == "" {
		return nil
	}

	if _, err := r.db.Exec("DELETE FROM chat_conversations WHERE id = ?", id); err != nil {
		return fmt.Errorf("chat repo: delete: %w", err)
	}
	return nil
}

// validateConversation validates repository and conversation inputs.
func (r *Repository) validateConversation(conv *chatdomain.Conversation) error {

	if r == nil || r.db == nil {
		return fmt.Errorf("chat repo: db required")
	}
	if conv == nil || conv.ID == "" {
		return fmt.Errorf("chat repo: conversation required")
	}
	return nil
}

// withTx executes a function in a transaction.
func withTx(db *sql.DB, fn func(tx *sql.Tx) error) (err error) {

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("chat repo: begin tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("chat repo: commit tx: %w", err)
	}

	return nil
}

// insertConversation inserts a new conversation row.
func insertConversation(tx *sql.Tx, conv *chatdomain.Conversation) error {

	_, err := tx.Exec(
		`INSERT INTO chat_conversations (id, title, provider, model, temperature, max_tokens, system_prompt, created_at, updated_at, is_archived)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		conv.ID,
		conv.Title,
		conv.Settings.Provider,
		conv.Settings.Model,
		conv.Settings.Temperature,
		conv.Settings.MaxTokens,
		conv.Settings.SystemPrompt,
		conv.CreatedAt,
		conv.UpdatedAt,
		boolToInt(conv.IsArchived),
	)
	if err != nil {
		return fmt.Errorf("chat repo: insert conversation: %w", err)
	}
	return nil
}

// upsertConversation inserts or updates a conversation row.
func upsertConversation(tx *sql.Tx, conv *chatdomain.Conversation) error {

	_, err := tx.Exec(
		`INSERT INTO chat_conversations (id, title, provider, model, temperature, max_tokens, system_prompt, created_at, updated_at, is_archived)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		 ON CONFLICT(id) DO UPDATE SET
		  title = excluded.title,
		  provider = excluded.provider,
		  model = excluded.model,
		  temperature = excluded.temperature,
		  max_tokens = excluded.max_tokens,
		  system_prompt = excluded.system_prompt,
		  created_at = excluded.created_at,
		  updated_at = excluded.updated_at,
		  is_archived = excluded.is_archived`,
		conv.ID,
		conv.Title,
		conv.Settings.Provider,
		conv.Settings.Model,
		conv.Settings.Temperature,
		conv.Settings.MaxTokens,
		conv.Settings.SystemPrompt,
		conv.CreatedAt,
		conv.UpdatedAt,
		boolToInt(conv.IsArchived),
	)
	if err != nil {
		return fmt.Errorf("chat repo: upsert conversation: %w", err)
	}
	return nil
}

// replaceMessages rewrites all messages and blocks for a conversation.
func replaceMessages(tx *sql.Tx, conversationID string, messages []*chatdomain.Message) error {

	if _, err := tx.Exec("DELETE FROM chat_messages WHERE conversation_id = ?", conversationID); err != nil {
		return fmt.Errorf("chat repo: delete messages: %w", err)
	}

	for _, message := range messages {
		if message == nil || message.ID == "" {
			continue
		}
		messageConversationID := message.ConversationID
		if messageConversationID == "" {
			messageConversationID = conversationID
		}

		provider := sql.NullString{}
		model := sql.NullString{}
		tokensIn := sql.NullInt64{}
		tokensOut := sql.NullInt64{}
		tokensTotal := sql.NullInt64{}
		latencyMs := sql.NullInt64{}
		finishReason := sql.NullString{}
		statusCode := sql.NullInt64{}
		errorMessage := sql.NullString{}

		if message.Metadata != nil {
			provider = newNullString(message.Metadata.Provider)
			model = newNullString(message.Metadata.Model)
			tokensIn = sql.NullInt64{Int64: int64(message.Metadata.TokensIn), Valid: true}
			tokensOut = sql.NullInt64{Int64: int64(message.Metadata.TokensOut), Valid: true}
			tokensTotal = sql.NullInt64{Int64: int64(message.Metadata.TokensTotal), Valid: true}
			if !tokensTotal.Valid || tokensTotal.Int64 == 0 {
				tokensTotal = sql.NullInt64{Int64: int64(message.Metadata.TokensIn + message.Metadata.TokensOut), Valid: true}
			}
			latencyMs = sql.NullInt64{Int64: message.Metadata.LatencyMs, Valid: true}
			finishReason = newNullString(message.Metadata.FinishReason)
			statusCode = sql.NullInt64{Int64: int64(message.Metadata.StatusCode), Valid: true}
			errorMessage = newNullString(message.Metadata.ErrorMessage)
		}

		if errorText := findFirstErrorContent(message.Blocks); errorText != "" {
			errorMessage = newNullString(errorText)
		}

		if _, err := tx.Exec(
			`INSERT INTO chat_messages
			 (id, conversation_id, role, timestamp, is_streaming, provider, model, tokens_in, tokens_out, tokens_total, latency_ms, finish_reason, status_code, error_message)
			 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			message.ID,
			messageConversationID,
			string(message.Role),
			message.Timestamp,
			boolToInt(message.IsStreaming),
			provider,
			model,
			tokensIn,
			tokensOut,
			tokensTotal,
			latencyMs,
			finishReason,
			statusCode,
			errorMessage,
		); err != nil {
			return fmt.Errorf("chat repo: insert message: %w", err)
		}

		for blockIndex, block := range message.Blocks {
			if err := insertMessageBlock(tx, message.ID, blockIndex, block); err != nil {
				return err
			}
		}
	}

	return nil
}

// insertMessageBlock inserts one block row for a message.
func insertMessageBlock(tx *sql.Tx, messageID string, blockIndex int, block chatdomain.Block) error {

	artifactID := sql.NullString{}
	artifactName := sql.NullString{}
	artifactType := sql.NullString{}
	artifactContent := sql.NullString{}
	artifactLanguage := sql.NullString{}
	artifactVersion := sql.NullInt64{}
	artifactCreatedAt := sql.NullInt64{}
	artifactUpdatedAt := sql.NullInt64{}

	if block.Artifact != nil {
		artifactID = newNullString(block.Artifact.ID)
		artifactName = newNullString(block.Artifact.Name)
		artifactType = newNullString(block.Artifact.Type)
		artifactContent = newNullString(block.Artifact.Content)
		artifactLanguage = newNullString(block.Artifact.Language)
		artifactVersion = sql.NullInt64{Int64: int64(block.Artifact.Version), Valid: true}
		artifactCreatedAt = sql.NullInt64{Int64: block.Artifact.CreatedAt, Valid: true}
		artifactUpdatedAt = sql.NullInt64{Int64: block.Artifact.UpdatedAt, Valid: true}
	}

	actionID := sql.NullString{}
	actionToolName := sql.NullString{}
	actionDescription := sql.NullString{}
	actionStatus := sql.NullString{}
	actionResult := sql.NullString{}
	actionStartedAt := sql.NullInt64{}
	actionCompletedAt := sql.NullInt64{}

	if block.Action != nil {
		actionID = newNullString(block.Action.ID)
		actionToolName = newNullString(block.Action.ToolName)
		actionDescription = newNullString(block.Action.Description)
		actionStatus = newNullString(string(block.Action.Status))
		actionResult = newNullString(block.Action.Result)
		actionStartedAt = sql.NullInt64{Int64: block.Action.StartedAt, Valid: true}
		actionCompletedAt = sql.NullInt64{Int64: block.Action.CompletedAt, Valid: true}
	}

	_, err := tx.Exec(
		`INSERT INTO chat_message_blocks
		 (message_id, block_index, block_type, content, language, is_collapsed,
		  artifact_id, artifact_name, artifact_type, artifact_content, artifact_language, artifact_version, artifact_created_at, artifact_updated_at,
		  action_id, action_tool_name, action_description, action_status, action_result, action_started_at, action_completed_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		messageID,
		blockIndex,
		string(block.Type),
		block.Content,
		nullableString(block.Language),
		boolToInt(block.IsCollapsed),
		artifactID,
		artifactName,
		artifactType,
		artifactContent,
		artifactLanguage,
		artifactVersion,
		artifactCreatedAt,
		artifactUpdatedAt,
		actionID,
		actionToolName,
		actionDescription,
		actionStatus,
		actionResult,
		actionStartedAt,
		actionCompletedAt,
	)
	if err != nil {
		return fmt.Errorf("chat repo: insert block: %w", err)
	}

	return nil
}

// loadMessages fetches messages and blocks for one conversation.
func loadMessages(db *sql.DB, conversationID string) ([]*chatdomain.Message, error) {

	rows, err := db.Query(
		`SELECT id, role, timestamp, is_streaming, provider, model, tokens_in, tokens_out, tokens_total, latency_ms, finish_reason, status_code, error_message
		 FROM chat_messages
		 WHERE conversation_id = ?
		 ORDER BY timestamp ASC, id ASC`,
		conversationID,
	)
	if err != nil {
		return nil, fmt.Errorf("chat repo: list messages: %w", err)
	}

	messages := make([]*chatdomain.Message, 0)
	for rows.Next() {
		msg := &chatdomain.Message{ConversationID: conversationID}
		var (
			role        string
			isStreaming int
			provider    sql.NullString
			model       sql.NullString
			tokensIn    sql.NullInt64
			tokensOut   sql.NullInt64
			tokensTotal sql.NullInt64
			latencyMs   sql.NullInt64
			finish      sql.NullString
			statusCode  sql.NullInt64
			errorText   sql.NullString
		)
		if err := rows.Scan(
			&msg.ID,
			&role,
			&msg.Timestamp,
			&isStreaming,
			&provider,
			&model,
			&tokensIn,
			&tokensOut,
			&tokensTotal,
			&latencyMs,
			&finish,
			&statusCode,
			&errorText,
		); err != nil {
			return nil, fmt.Errorf("chat repo: scan message: %w", err)
		}
		msg.Role = chatdomain.Role(role)
		msg.IsStreaming = isStreaming == 1

		meta := &chatdomain.MessageMetadata{}
		if provider.Valid {
			meta.Provider = provider.String
		}
		if model.Valid {
			meta.Model = model.String
		}
		if tokensIn.Valid {
			meta.TokensIn = int(tokensIn.Int64)
		}
		if tokensOut.Valid {
			meta.TokensOut = int(tokensOut.Int64)
		}
		if tokensTotal.Valid {
			meta.TokensTotal = int(tokensTotal.Int64)
		}
		if latencyMs.Valid {
			meta.LatencyMs = latencyMs.Int64
		}
		if finish.Valid {
			meta.FinishReason = finish.String
		}
		if statusCode.Valid {
			meta.StatusCode = int(statusCode.Int64)
		}
		if errorText.Valid {
			meta.ErrorMessage = errorText.String
		}
		if hasMetadata(meta) {
			msg.Metadata = meta
		}

		messages = append(messages, msg)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat repo: message rows: %w", err)
	}
	if err := rows.Close(); err != nil {
		return nil, fmt.Errorf("chat repo: close message rows: %w", err)
	}

	for _, msg := range messages {
		blocks, err := loadBlocks(db, msg.ID)
		if err != nil {
			return nil, err
		}
		msg.Blocks = blocks
	}

	return messages, nil
}

// loadBlocks fetches all blocks for one message.
func loadBlocks(db *sql.DB, messageID string) ([]chatdomain.Block, error) {

	rows, err := db.Query(
		`SELECT block_type, content, language, is_collapsed,
		        artifact_id, artifact_name, artifact_type, artifact_content, artifact_language, artifact_version, artifact_created_at, artifact_updated_at,
		        action_id, action_tool_name, action_description, action_status, action_result, action_started_at, action_completed_at
		 FROM chat_message_blocks
		 WHERE message_id = ?
		 ORDER BY block_index ASC`,
		messageID,
	)
	if err != nil {
		return nil, fmt.Errorf("chat repo: list blocks: %w", err)
	}
	defer rows.Close()

	blocks := make([]chatdomain.Block, 0)
	for rows.Next() {
		var (
			blockType        string
			content          string
			language         sql.NullString
			isCollapsed      int
			artifactID       sql.NullString
			artifactName     sql.NullString
			artifactType     sql.NullString
			artifactContent  sql.NullString
			artifactLanguage sql.NullString
			artifactVersion  sql.NullInt64
			artifactCreated  sql.NullInt64
			artifactUpdated  sql.NullInt64
			actionID         sql.NullString
			actionToolName   sql.NullString
			actionDesc       sql.NullString
			actionStatus     sql.NullString
			actionResult     sql.NullString
			actionStarted    sql.NullInt64
			actionCompleted  sql.NullInt64
		)
		if err := rows.Scan(
			&blockType,
			&content,
			&language,
			&isCollapsed,
			&artifactID,
			&artifactName,
			&artifactType,
			&artifactContent,
			&artifactLanguage,
			&artifactVersion,
			&artifactCreated,
			&artifactUpdated,
			&actionID,
			&actionToolName,
			&actionDesc,
			&actionStatus,
			&actionResult,
			&actionStarted,
			&actionCompleted,
		); err != nil {
			return nil, fmt.Errorf("chat repo: scan block: %w", err)
		}

		block := chatdomain.Block{
			Type:        chatdomain.BlockType(blockType),
			Content:     content,
			Language:    nullableValue(language),
			IsCollapsed: isCollapsed == 1,
		}
		if artifactID.Valid || artifactName.Valid || artifactType.Valid || artifactContent.Valid {
			block.Artifact = &chatdomain.Artifact{
				ID:        nullableValue(artifactID),
				Name:      nullableValue(artifactName),
				Type:      nullableValue(artifactType),
				Content:   nullableValue(artifactContent),
				Language:  nullableValue(artifactLanguage),
				Version:   int(nullableInt64Value(artifactVersion)),
				CreatedAt: nullableInt64Value(artifactCreated),
				UpdatedAt: nullableInt64Value(artifactUpdated),
			}
		}
		if actionID.Valid || actionToolName.Valid || actionDesc.Valid || actionStatus.Valid || actionResult.Valid {
			block.Action = &chatdomain.ActionExecution{
				ID:          nullableValue(actionID),
				ToolName:    nullableValue(actionToolName),
				Description: nullableValue(actionDesc),
				Status:      chatdomain.ActionStatus(nullableValue(actionStatus)),
				Result:      nullableValue(actionResult),
				StartedAt:   nullableInt64Value(actionStarted),
				CompletedAt: nullableInt64Value(actionCompleted),
				Args:        nil,
			}
		}

		blocks = append(blocks, block)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("chat repo: block rows: %w", err)
	}

	return blocks, nil
}

// hasMetadata reports whether metadata carries meaningful values.
func hasMetadata(meta *chatdomain.MessageMetadata) bool {

	if meta == nil {
		return false
	}
	return meta.Provider != "" ||
		meta.Model != "" ||
		meta.TokensIn != 0 ||
		meta.TokensOut != 0 ||
		meta.TokensTotal != 0 ||
		meta.LatencyMs != 0 ||
		meta.FinishReason != "" ||
		meta.StatusCode != 0 ||
		meta.ErrorMessage != ""
}

// findFirstErrorContent extracts the first error block text in a message.
func findFirstErrorContent(blocks []chatdomain.Block) string {

	for _, block := range blocks {
		if block.Type == chatdomain.BlockTypeError && block.Content != "" {
			return block.Content
		}
	}
	return ""
}

// newNullString maps empty strings to NULL.
func newNullString(value string) sql.NullString {

	if value == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: value, Valid: true}
}

// nullableString maps empty strings to NULL for writes.
func nullableString(value string) sql.NullString {

	return newNullString(value)
}

// nullableValue converts a nullable string to a plain value.
func nullableValue(value sql.NullString) string {

	if !value.Valid {
		return ""
	}
	return value.String
}

// nullableInt64Value converts a nullable int64 to a plain value.
func nullableInt64Value(value sql.NullInt64) int64 {

	if !value.Valid {
		return 0
	}
	return value.Int64
}

// boolToInt maps booleans into SQLite-friendly integers.
func boolToInt(value bool) int {

	if value {
		return 1
	}
	return 0
}
