package provider

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"plan-agent/internal/core/messages"
)

// MinimaxProvider implements LLMProvider for MiniMax's OpenAI-compatible API
type MinimaxProvider struct {
	httpClient *http.Client
	endpoint   string
}

// NewMinimaxProvider creates a new MiniMax provider
func NewMinimaxProvider() *MinimaxProvider {
	return &MinimaxProvider{
		httpClient: &http.Client{Timeout: 120 * time.Second},
		endpoint:   "https://api.minimax.io/v1",
	}
}

func (p *MinimaxProvider) Name() string {
	return "minimax"
}

func (p *MinimaxProvider) Stream(
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
	if len(toolDefs) > 0 {
		go p.handleNonStreamingResponse(resp.Body, events)
	} else {
		go p.streamResponse(resp.Body, events)
	}
	return events, nil
}

func (p *MinimaxProvider) buildRequest(
	model Model,
	msgs []messages.Message,
	toolDefs []ToolDefinition,
	opts StreamingOptions,
) ([]byte, string, error) {
	stream := len(toolDefs) == 0

	reqBody := map[string]any{
		"model":    model.Name,
		"stream":   stream,
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

	if os.Getenv("MINIMAX_DEBUG") != "" {
		log.Printf("[MINIMAX] request: %s", string(body))
	}

	endpoint := p.endpoint + "/chat/completions"
	if model.Endpoint != "" {
		endpoint = model.Endpoint
	}

	return body, endpoint, nil
}

func (p *MinimaxProvider) doRequest(ctx context.Context, endpoint, apiKey string, body []byte) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(string(body)))
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

// streamResponse reads MiniMax's SSE streaming format ("data: {...}" per line)
func (p *MinimaxProvider) streamResponse(body io.Reader, events chan<- StreamEvent) {
	defer close(events)
	defer func() { recover() }()

	scanner := bufio.NewScanner(body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	pendingToolCalls := make(map[string]*messages.ToolCallBlock)
	partialMsg := newAssistantMessage()
	stopReasonSeen := ""
	extraChunks := 0

	for scanner.Scan() {
		line := scanner.Text()
		if line == "" || !strings.HasPrefix(line, "data: ") {
			continue
		}

		dataStr := strings.TrimPrefix(line, "data: ")
		if dataStr == "[DONE]" {
			if len(partialMsg.Content) > 0 {
				partialMsg.StopReason = "end_turn"
				events <- &DoneEvent{Message: partialMsg}
			}
			return
		}

		event, stopReason := p.parseChunk([]byte(dataStr), partialMsg, pendingToolCalls)
		if event != nil {
			events <- event
		}
		if stopReason != "" {
			p.finalizeToolCalls(partialMsg, pendingToolCalls)
			partialMsg.StopReason = normalizeStopReason(stopReason)

			// Användningen kommer ibland i ett eget stycke efter det
			// sista innehållet. Därför läses streamen vidare tills
			// användningen kommer - annars blir kostnaden noll.
			if hasUsage(partialMsg.Usage) {
				events <- &DoneEvent{Message: partialMsg}
				return
			}

			stopReasonSeen = stopReason
			extraChunks++
			if extraChunks > maxChunksAfterStop {
				events <- &DoneEvent{Message: partialMsg}
				return
			}
			continue
		}
	}

	if err := scanner.Err(); err != nil {
		events <- &ErrorEvent{Error: err}
		return
	}

	if len(partialMsg.Content) > 0 {
		if stopReasonSeen != "" {
			partialMsg.StopReason = normalizeStopReason(stopReasonSeen)
		} else {
			partialMsg.StopReason = "end_turn"
		}
		events <- &DoneEvent{Message: partialMsg}
	}
}

// maxChunksAfterStop är hur många extra stycken som läses efter att
// modellen angett att den är klar. Det ska räcka för att användningen
// ska hinna komma.
const maxChunksAfterStop = 5

// hasUsage berättar om användningen innehåller några tokens.
func hasUsage(usage *messages.Usage) bool {
	return usage != nil && (usage.TotalTokens > 0 || usage.InputTokens > 0 || usage.OutputTokens > 0)
}

func (p *MinimaxProvider) parseChunk(data []byte, partialMsg *messages.AssistantMessage, _ map[string]*messages.ToolCallBlock) (StreamEvent, string) {
	var parsed struct {
		Choices []struct {
			Index        int            `json:"index"`
			Delta        map[string]any `json:"delta"`
			FinishReason string         `json:"finish_reason"`
		} `json:"choices"`
		Usage *struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
			CachedTokens     int `json:"cached_tokens"`
		} `json:"usage"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, ""
	}

	// Användningen sparas på meddelandet, så att den kan rapporteras
	// vidare. Utan den går det inte att räkna ut vad körningen kostade.
	if parsed.Usage != nil {
		usage := &messages.Usage{
			InputTokens:  parsed.Usage.PromptTokens,
			OutputTokens: parsed.Usage.CompletionTokens,
			TotalTokens:  parsed.Usage.TotalTokens,
			CacheTokens:  parsed.Usage.CachedTokens,
		}
		// En tom användning ska inte skriva över en riktig.
		if hasUsage(usage) || partialMsg.Usage == nil {
			partialMsg.Usage = usage
		}
	}

	if parsed.Error != nil {
		return &ErrorEvent{Error: errors.New(parsed.Error.Message)}, ""
	}

	var lastEvent StreamEvent

	for _, choice := range parsed.Choices {
		delta := choice.Delta

		if content, ok := delta["content"]; ok {
			if contentStr, ok := content.(string); ok {
				partialMsg.Content = append(partialMsg.Content, &messages.TextBlock{Text: contentStr})
				lastEvent = &TextDeltaEvent{Delta: contentStr, Partial: partialMsg}
				continue
			}
		}

		if tc, ok := delta["tool_calls"]; ok {
			if tcArray, ok := tc.([]any); ok {
				for _, tcItem := range tcArray {
					if tcMap, ok := tcItem.(map[string]any); ok {
						if tcBlock := handleToolCallDelta(tcMap, partialMsg); tcBlock != nil {
							lastEvent = &ToolCallDeltaEvent{
								ID:        tcBlock.ID,
								Name:      tcBlock.Name,
								Arguments: fmt.Sprint(tcMap["arguments"]),
								Partial:   partialMsg,
							}
						}
					}
				}
			}
		}

		if choice.FinishReason != "" {
			return nil, choice.FinishReason
		}
	}

	return lastEvent, ""
}

func (p *MinimaxProvider) finalizeToolCalls(partialMsg *messages.AssistantMessage, pending map[string]*messages.ToolCallBlock) {
	for _, tc := range pending {
		args := tc.Arguments["_raw_args"].(string)
		parsed := ParseJSONArgs(args)
		if _, hasRaw := parsed["_raw"]; !hasRaw {
			tc.Arguments = parsed
			partialMsg.Content = append(partialMsg.Content, tc)
		} else {
			tc.Arguments = map[string]any{"error": "incomplete arguments: " + args}
			partialMsg.Content = append(partialMsg.Content, tc)
		}
	}
}

func normalizeStopReason(reason string) string {
	switch reason {
	case "stop", "eos":
		return "end_turn"
	default:
		return reason
	}
}

// handleNonStreamingResponse handles non-streaming JSON responses
func (p *MinimaxProvider) handleNonStreamingResponse(body io.Reader, events chan<- StreamEvent) {
	defer close(events)
	defer func() { recover() }()

	data, err := io.ReadAll(body)
	if err != nil {
		events <- &ErrorEvent{Error: err}
		return
	}

	var response struct {
		Choices []struct {
			Index        int         `json:"index"`
			Message      interface{} `json:"message"`
			FinishReason string      `json:"finish_reason"`
		} `json:"choices"`
		Error *struct {
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(data, &response); err != nil {
		events <- &ErrorEvent{Error: err}
		return
	}

	if response.Error != nil {
		events <- &ErrorEvent{Error: errors.New(response.Error.Message)}
		return
	}

	if len(response.Choices) == 0 {
		return
	}

	choice := response.Choices[0]
	msg := newAssistantMessage()
	msg.StopReason = normalizeStopReason(choice.FinishReason)

	if msgMap, ok := choice.Message.(map[string]any); ok {
		if content, ok := msgMap["content"].(string); ok && content != "" {
			msg.Content = append(msg.Content, &messages.TextBlock{Text: content})
			events <- &TextDeltaEvent{Delta: content, Partial: msg}
		}
		if tc, ok := msgMap["tool_calls"].([]any); ok {
			log.Printf("[MINIMAX] Non-streaming: found %d tool calls in message", len(tc))
			for _, tcItem := range tc {
				if tcMap, ok := tcItem.(map[string]any); ok {
					if tcBlock := handleToolCallFromNonStreaming(tcMap); tcBlock != nil {
						log.Printf("[MINIMAX] Non-streaming tool call: name=%s id=%s", tcBlock.Name, tcBlock.ID)
						msg.Content = append(msg.Content, tcBlock)
					}
				}
			}
		}
	}

	events <- &DoneEvent{Message: msg}
}

func handleToolCallFromNonStreaming(tcMap map[string]any) *messages.ToolCallBlock {
	index, _ := tcMap["index"].(float64)
	id, _ := tcMap["id"].(string)

	funcObj, ok := tcMap["function"].(map[string]any)
	if !ok {
		return nil
	}
	name, _ := funcObj["name"].(string)
	args, _ := funcObj["arguments"].(string)

	if name == "" {
		return nil
	}

	if id == "" {
		id = fmt.Sprintf("call_%d_%s", int(index), name)
	}

	parsedArgs := ParseJSONArgs(args)

	return &messages.ToolCallBlock{
		ID:        id,
		Name:      name,
		Arguments: parsedArgs,
	}
}
