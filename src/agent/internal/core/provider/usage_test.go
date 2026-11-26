package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"plan/agent/internal/core/messages"
)

// TestAnvändningFångasIVanligtSvar kontrollerar att tokenanvändningen
// läses ur svaret - utan den går det inte att räkna ut kostnaden.
func TestAnvändningFångasIVanligtSvar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Användningen efterfrågas - då ska svaret inte strömmas.
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("kunde inte läsa förfrågan: %v", err)
		}
		if streamed, ok := body["stream"].(bool); ok && streamed {
			t.Error("svaret ska inte strömmas när användningen efterfrågas")
		}

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"choices": [{"index": 0, "finish_reason": "stop", "message": {"content": "hej"}}],
			"usage": {"prompt_tokens": 120, "completion_tokens": 30, "total_tokens": 150}
		}`))
	}))
	defer server.Close()

	provider := NewMinimaxProvider()
	provider.endpoint = server.URL

	model := Model{Provider: "minimax", Name: "MiniMax-M2.7", APIKey: "test"}
	msgs := []messages.Message{
		&messages.UserMessage{
			Role:    messages.RoleUser,
			Content: []messages.ContentBlock{&messages.TextBlock{Text: "hej"}},
		},
	}

	stream, err := provider.Stream(context.Background(), model, msgs, nil, StreamingOptions{Usage: true})
	if err != nil {
		t.Fatalf("Stream: %v", err)
	}

	var usage *messages.Usage
	for event := range stream {
		switch e := event.(type) {
		case *DoneEvent:
			usage = e.Message.Usage
		case *ErrorEvent:
			t.Fatalf("fel: %v", e.Error)
		}
	}

	if usage == nil {
		t.Fatal("användningen saknas")
	}
	if usage.InputTokens != 120 || usage.OutputTokens != 30 {
		t.Errorf("användning = %+v", usage)
	}
}
