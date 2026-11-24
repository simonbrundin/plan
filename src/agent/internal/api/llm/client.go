package llm

import (
	"context"
	"os"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
	"github.com/anthropics/anthropic-sdk-go/option"
)

// Client is a wrapper for Anthropic-compatible APIs
type Client struct {
	anthropic anthropic.Client
	baseURL   string
}

// Config for the LLM client
type Config struct {
	APIKey  string
	BaseURL string
}

// NewClient creates a new LLM client
func NewClient(cfg Config) *Client {
	opts := []option.RequestOption{
		option.WithAPIKey(cfg.APIKey),
		option.WithoutEnvironmentDefaults(),
	}

	if cfg.BaseURL != "" {
		opts = append(opts, option.WithBaseURL(cfg.BaseURL))
	}

	client := anthropic.NewClient(opts...)

	return &Client{
		anthropic: client,
		baseURL:   cfg.BaseURL,
	}
}

// DefaultClient returns a client configured from environment variables
func DefaultClient() *Client {
	apiKey := os.Getenv("MINIMAX_API_KEY")
	baseURL := "https://api.minimax.io/anthropic"

	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
		baseURL = "https://api.anthropic.com"
	}

	return NewClient(Config{
		APIKey:  apiKey,
		BaseURL: baseURL,
	})
}

// ProviderName returns the name of the current provider
func (c *Client) ProviderName() string {
	switch c.baseURL {
	case "https://api.minimax.io/anthropic":
		return "minimax"
	case "https://api.anthropic.com":
		return "anthropic"
	default:
		return "custom"
	}
}

// ChatRequest for chat
type ChatRequest struct {
	Model       string
	Messages    []Message
	Temperature float64
}

// Message represents a chat message
type Message struct {
	Role    string
	Content string
}

// ChatResponse from non-streaming chat
type ChatResponse struct {
	Content      string
	Model        string
	Provider     string
	TokensUsed   int
	LatencyMs    int64
	FinishReason string
	Error        string
}

// Chat performs a non-streaming chat request
func (c *Client) Chat(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	start := time.Now()

	// Find last user message
	var userContent string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userContent = req.Messages[i].Content
			break
		}
	}

	model := req.Model
	if model == "" {
		model = "MiniMax-M2.7"
	}

	maxTokens := int64(8192)
	params := anthropic.MessageNewParams{
		MaxTokens: maxTokens,
		Model:     model,
		Messages: []anthropic.MessageParam{
			anthropic.NewUserMessage(anthropic.NewTextBlock(userContent)),
		},
	}

	if req.Temperature > 0 {
		params.Temperature = anthropic.Float(req.Temperature)
	}

	resp, err := c.anthropic.Messages.New(ctx, params)
	if err != nil {
		return &ChatResponse{
			Provider:  c.ProviderName(),
			Model:     model,
			Error:     err.Error(),
			LatencyMs: time.Since(start).Milliseconds(),
		}, err
	}

	content := ""
	// Handle ContentBlockUnion - check Type field
	for _, block := range resp.Content {
		if block.Type == "text" {
			content = block.Text
			break
		}
	}

	finishReason := ""
	if resp.StopReason != "" {
		finishReason = string(resp.StopReason)
	}

	return &ChatResponse{
		Content:      content,
		Model:        model,
		Provider:     c.ProviderName(),
		TokensUsed:   int(resp.Usage.InputTokens + resp.Usage.OutputTokens),
		LatencyMs:    time.Since(start).Milliseconds(),
		FinishReason: finishReason,
	}, nil
}
