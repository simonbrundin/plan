package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Conversation represents a chat conversation
type Conversation struct {
	ID        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Title     string    `json:"title"`
	Model     string    `json:"model"`
	Provider  string    `json:"provider"`
}

// Message represents a single message in a conversation
type Message struct {
	ID             uuid.UUID `json:"id"`
	ConversationID uuid.UUID `json:"conversation_id"`
	Role           string    `json:"role"`
	Content        string    `json:"content"`
	ToolName       string    `json:"tool_name,omitempty"`
	ToolArgs       string    `json:"tool_args,omitempty"`
	ToolResult     string    `json:"tool_result,omitempty"`
	IsError        bool      `json:"is_error,omitempty"`
	Thinking       string    `json:"thinking,omitempty"`
	TokensUsed     int       `json:"tokens_used"`
	LatencyMs      int       `json:"latency_ms"`
	CreatedAt      time.Time `json:"created_at"`
}

// CreateConversation creates a new conversation
func CreateConversation(ctx context.Context, model, provider string) (*Conversation, error) {
	var conv Conversation
	err := pool.QueryRow(ctx,
		`INSERT INTO conversations (model, provider) 
		 VALUES ($1, $2) 
		 RETURNING id, created_at, updated_at, title, model, provider`,
		model, provider,
	).Scan(&conv.ID, &conv.CreatedAt, &conv.UpdatedAt, &conv.Title, &conv.Model, &conv.Provider)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// GetConversation retrieves a conversation by ID
func GetConversation(ctx context.Context, id uuid.UUID) (*Conversation, error) {
	var conv Conversation
	err := pool.QueryRow(ctx,
		`SELECT id, created_at, updated_at, title, model, provider 
		 FROM conversations WHERE id = $1`,
		id,
	).Scan(&conv.ID, &conv.CreatedAt, &conv.UpdatedAt, &conv.Title, &conv.Model, &conv.Provider)
	if err != nil {
		return nil, err
	}
	return &conv, nil
}

// ListConversations returns all conversations, newest first
func ListConversations(ctx context.Context, limit, offset int) ([]*Conversation, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, created_at, updated_at, title, model, provider 
		 FROM conversations 
		 ORDER BY created_at DESC 
		 LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var conversations []*Conversation
	for rows.Next() {
		var conv Conversation
		if err := rows.Scan(&conv.ID, &conv.CreatedAt, &conv.UpdatedAt, &conv.Title, &conv.Model, &conv.Provider); err != nil {
			return nil, err
		}
		conversations = append(conversations, &conv)
	}
	return conversations, nil
}

// CreateMessage creates a new message in a conversation
func CreateMessage(ctx context.Context, convID uuid.UUID, role, content string, tokensUsed, latencyMs int) (*Message, error) {
	var msg Message
	err := pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, role, content, tokens_used, latency_ms) 
		 VALUES ($1, $2, $3, $4, $5) 
		 RETURNING id, conversation_id, role, content, tool_name, tool_args, tool_result, is_error, thinking, tokens_used, latency_ms, created_at`,
		convID, role, content, tokensUsed, latencyMs,
	).Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.ToolName, &msg.ToolArgs, &msg.ToolResult, &msg.IsError, &msg.Thinking, &msg.TokensUsed, &msg.LatencyMs, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	pool.Exec(ctx, `UPDATE conversations SET updated_at = NOW() WHERE id = $1`, convID)

	return &msg, nil
}

// CreateToolMessage creates a tool call or tool result message
func CreateToolMessage(ctx context.Context, convID uuid.UUID, role, toolName, toolArgs, toolResult string, isError bool) (*Message, error) {
	var msg Message
	err := pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, role, content, tool_name, tool_args, tool_result, is_error) 
		 VALUES ($1, $2, $3, $4, $5, $6, $7) 
		 RETURNING id, conversation_id, role, content, tool_name, tool_args, tool_result, is_error, thinking, tokens_used, latency_ms, created_at`,
		convID, role, " ", toolName, toolArgs, toolResult, isError,
	).Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.ToolName, &msg.ToolArgs, &msg.ToolResult, &msg.IsError, &msg.Thinking, &msg.TokensUsed, &msg.LatencyMs, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	pool.Exec(ctx, `UPDATE conversations SET updated_at = NOW() WHERE id = $1`, convID)

	return &msg, nil
}

// CreateThinkingMessage creates a thinking/reasoning message
func CreateThinkingMessage(ctx context.Context, convID uuid.UUID, thinking string) (*Message, error) {
	var msg Message
	err := pool.QueryRow(ctx,
		`INSERT INTO messages (conversation_id, role, content, thinking) 
		 VALUES ($1, $2, $3, $4) 
		 RETURNING id, conversation_id, role, content, tool_name, tool_args, tool_result, is_error, thinking, tokens_used, latency_ms, created_at`,
		convID, "thinking", "", thinking,
	).Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.ToolName, &msg.ToolArgs, &msg.ToolResult, &msg.IsError, &msg.Thinking, &msg.TokensUsed, &msg.LatencyMs, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	pool.Exec(ctx, `UPDATE conversations SET updated_at = NOW() WHERE id = $1`, convID)

	return &msg, nil
}

// GetMessages retrieves all messages for a conversation
func GetMessages(ctx context.Context, convID uuid.UUID) ([]*Message, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, conversation_id, role, content, tool_name, tool_args, tool_result, is_error, thinking, tokens_used, latency_ms, created_at 
		 FROM messages WHERE conversation_id = $1 
		 ORDER BY created_at ASC`,
		convID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.ToolName, &msg.ToolArgs, &msg.ToolResult, &msg.IsError, &msg.Thinking, &msg.TokensUsed, &msg.LatencyMs, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

// UpdateConversationTitle updates the conversation title
func UpdateConversationTitle(ctx context.Context, convID uuid.UUID, title string) error {
	_, err := pool.Exec(ctx,
		`UPDATE conversations SET title = $1, updated_at = NOW() WHERE id = $2`,
		title, convID,
	)
	return err
}

// APIKey represents an API key for a provider
type APIKey struct {
	ID        int       `json:"id"`
	Provider  string    `json:"provider"`
	APIKey    string    `json:"api_key"`
	Model     string    `json:"model"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// GetAPIKeys returns all configured API keys
func GetAPIKeys(ctx context.Context) ([]*APIKey, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, provider, api_key, model, created_at, updated_at FROM api_keys`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var keys []*APIKey
	for rows.Next() {
		var k APIKey
		if err := rows.Scan(&k.ID, &k.Provider, &k.APIKey, &k.Model, &k.CreatedAt, &k.UpdatedAt); err != nil {
			return nil, err
		}
		keys = append(keys, &k)
	}
	return keys, nil
}

// GetAPIKey returns the API key for a specific provider
func GetAPIKey(ctx context.Context, provider string) (*APIKey, error) {
	var k APIKey
	err := pool.QueryRow(ctx,
		`SELECT id, provider, api_key, model, created_at, updated_at FROM api_keys WHERE provider = $1`,
		provider,
	).Scan(&k.ID, &k.Provider, &k.APIKey, &k.Model, &k.CreatedAt, &k.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &k, nil
}

// SetAPIKey stores or updates an API key
func SetAPIKey(ctx context.Context, provider, apiKey, model string) error {
	_, err := pool.Exec(ctx,
		`INSERT INTO api_keys (provider, api_key, model) VALUES ($1, $2, $3)
		 ON CONFLICT (provider) DO UPDATE SET api_key = $2, model = $3, updated_at = NOW()`,
		provider, apiKey, model,
	)
	return err
}

// DeleteAPIKey removes an API key
func DeleteAPIKey(ctx context.Context, provider string) error {
	_, err := pool.Exec(ctx, `DELETE FROM api_keys WHERE provider = $1`, provider)
	return err
}

// GetMessagesSince retrieves messages after a certain time (for SSE polling)
func GetMessagesSince(ctx context.Context, convID uuid.UUID, since time.Time) ([]*Message, error) {
	rows, err := pool.Query(ctx,
		`SELECT id, conversation_id, role, content, tokens_used, latency_ms, created_at 
		 FROM messages WHERE conversation_id = $1 AND created_at > $2 
		 ORDER BY created_at ASC`,
		convID, since,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*Message
	for rows.Next() {
		var msg Message
		if err := rows.Scan(&msg.ID, &msg.ConversationID, &msg.Role, &msg.Content, &msg.TokensUsed, &msg.LatencyMs, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}
