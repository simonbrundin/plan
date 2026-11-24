package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"plan-agent/internal/core/messages"
)

// OpenAIProvider implements LLMProvider for OpenAI-compatible APIs
type OpenAIProvider struct {
	httpClient *http.Client
	endpoint   string
}

// NewOpenAIProvider creates a new OpenAI provider
func NewOpenAIProvider() *OpenAIProvider {
	return &OpenAIProvider{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		endpoint:   "https://api.openai.com/v1",
	}
}

func (p *OpenAIProvider) Name() string {
	return "openai"
}

func (p *OpenAIProvider) Stream(
	ctx context.Context,
	model Model,
	msgs []messages.Message,
	toolDefs []ToolDefinition,
	opts StreamingOptions,
) (<-chan StreamEvent, error) {
	body, endpoint, err := p.buildRequest(model, msgs, toolDefs, opts)
	if err != nil {
		return nil, err
	}

	resp, err := p.doRequest(ctx, endpoint, model.APIKey, body)
	if err != nil {
		return nil, err
	}

	events := make(chan StreamEvent, 100)
	go p.streamResponse(resp.Body, events)
	return events, nil
}

func (p *OpenAIProvider) buildRequest(
	model Model,
	msgs []messages.Message,
	toolDefs []ToolDefinition,
	opts StreamingOptions,
) ([]byte, string, error) {
	reqBody := map[string]any{
		"model":    model.Name,
		"stream":   true,
		"messages": convertMessages(msgs),
	}

	if opts.MaxTokens != nil {
		reqBody["max_tokens"] = *opts.MaxTokens
	}
	if opts.Temperature != nil {
		reqBody["temperature"] = *opts.Temperature
	}

	if len(toolDefs) > 0 {
		reqBody["tools"] = toolDefsToOpenAIFormat(toolDefs)
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, "", fmt.Errorf("failed to marshal request: %w", err)
	}

	endpoint := p.endpoint + "/chat/completions"
	if model.Endpoint != "" {
		endpoint = model.Endpoint
	}

	return body, endpoint, nil
}

func (p *OpenAIProvider) doRequest(ctx context.Context, endpoint, apiKey string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API error: %s - %s", resp.Status, string(bodyBytes))
	}

	return resp, nil
}

// streamResponse reads OpenAI's NDJSON streaming format (one JSON object per line)
func (p *OpenAIProvider) streamResponse(body io.Reader, events chan<- StreamEvent) {
	defer close(events)
	defer func() { recover() }() // Handle panic on close

	decoder := json.NewDecoder(body)
	partialMsg := newAssistantMessage()

	for decoder.More() {
		var raw json.RawMessage
		if err := decoder.Decode(&raw); err != nil {
			events <- &ErrorEvent{Error: err}
			return
		}

		event, stopReason := p.parseChunk(raw, partialMsg)
		if event != nil {
			events <- event
		}
		if stopReason != "" {
			partialMsg.StopReason = stopReason
			events <- &DoneEvent{Message: partialMsg}
			return
		}
	}

	// If we get here without finish_reason, send what we have
	if len(partialMsg.Content) > 0 {
		partialMsg.StopReason = "end_turn"
		events <- &DoneEvent{Message: partialMsg}
	}
}

func (p *OpenAIProvider) parseChunk(raw json.RawMessage, partialMsg *messages.AssistantMessage) (StreamEvent, string) {
	var data struct {
		Choices []struct {
			Index        int            `json:"index"`
			Delta        map[string]any `json:"delta"`
			FinishReason string         `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, ""
	}

	if data.Error != nil {
		return &ErrorEvent{Error: errors.New(data.Error.Message)}, ""
	}

	for _, choice := range data.Choices {
		delta := choice.Delta

		if content, ok := delta["content"]; ok {
			if contentStr, ok := content.(string); ok {
				partialMsg.Content = append(partialMsg.Content, &messages.TextBlock{Text: contentStr})
				return &TextDeltaEvent{Delta: contentStr, Partial: partialMsg}, ""
			}
		}

		if tc, ok := delta["tool_calls"]; ok {
			if tcArray, ok := tc.([]any); ok {
				for _, tcItem := range tcArray {
					if tcMap, ok := tcItem.(map[string]any); ok {
						handleToolCallDelta(tcMap, partialMsg)
					}
				}
			}
		}

		if choice.FinishReason != "" {
			return nil, choice.FinishReason
		}
	}

	return nil, ""
}
