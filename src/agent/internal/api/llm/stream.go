package llm

import (
	"context"
	"errors"
	"time"

	"github.com/anthropics/anthropic-sdk-go"
)

// StreamChunk sent to the client
type StreamChunk struct {
	Type     string `json:"type"`
	Content  string `json:"content,omitempty"`
	Tokens   int    `json:"tokens,omitempty"`
	Latency  int64  `json:"latency,omitempty"`
	Reason   string `json:"reason,omitempty"`
	ErrorMsg string `json:"error,omitempty"`
	Thinking string `json:"thinking,omitempty"`
}

// StreamChat performs a streaming chat request
func (c *Client) StreamChat(ctx context.Context, req ChatRequest, writer func(StreamChunk)) error {
	start := time.Now()

	// Find last user message
	var userContent string
	for i := len(req.Messages) - 1; i >= 0; i-- {
		if req.Messages[i].Role == "user" {
			userContent = req.Messages[i].Content
			break
		}
	}

	if userContent == "" {
		writer(StreamChunk{Type: "error", ErrorMsg: "No user message found"})
		return errors.New("no user message found")
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

	// Use streaming
	stream := c.anthropic.Messages.NewStreaming(ctx, params)

	for stream.Next() {
		event := stream.Current()

		// Handle different event types using AsAny()
		switch e := event.AsAny().(type) {
		case anthropic.ContentBlockDeltaEvent:
			// Content delta - check if it's a text delta or thinking delta
			if e.Delta.Type == "text_delta" {
				writer(StreamChunk{
					Type:    "content",
					Content: e.Delta.Text,
				})
			} else if e.Delta.Type == "thinking_delta" {
				// Send thinking block separately for display
				writer(StreamChunk{
					Type:     "thinking",
					Thinking: e.Delta.Text,
				})
			}

		case anthropic.MessageDeltaEvent:
			// Message done
			if e.Delta.StopReason != "" {
				writer(StreamChunk{
					Type:    "done",
					Reason:  string(e.Delta.StopReason),
					Latency: time.Since(start).Milliseconds(),
				})
			}
		}
	}

	if err := stream.Err(); err != nil {
		writer(StreamChunk{
			Type:     "error",
			ErrorMsg: err.Error(),
			Latency:  time.Since(start).Milliseconds(),
		})
		return err
	}

	return nil
}
