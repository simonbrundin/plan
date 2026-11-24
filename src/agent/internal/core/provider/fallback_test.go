package provider

import (
	"context"
	"errors"
	"testing"

	msgpkg "plan-agent/internal/core/messages"
)

// mockProvider is a test provider that can be configured to fail or succeed
type mockProvider struct {
	name       string
	shouldErr  bool
	errMsg     string
	eventCount int
}

func (m *mockProvider) Name() string {
	return m.name
}

func (m *mockProvider) Stream(
	_ context.Context,
	_ Model,
	_ []msgpkg.Message,
	_ []ToolDefinition,
	_ StreamingOptions,
) (<-chan StreamEvent, error) {
	m.eventCount++
	if m.shouldErr {
		return nil, errors.New(m.errMsg)
	}

	events := make(chan StreamEvent, 2)
	// Non-blocking send
	select {
	case events <- &StartEvent{Partial: &msgpkg.AssistantMessage{
		Role:    "assistant",
		Content: []msgpkg.ContentBlock{&msgpkg.TextBlock{Text: "test"}},
	}}:
	default:
	}
	close(events)
	return events, nil
}

func TestProviderManagerSuccess(t *testing.T) {
	mgr := NewProviderManager()
	mgr.Add("mock1", &mockProvider{name: "mock1", shouldErr: false})
	mgr.Add("mock2", &mockProvider{name: "mock2", shouldErr: false})

	events, name, err := mgr.Stream(context.Background(), Model{Name: "test"}, nil, nil, StreamingOptions{})
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}

	if name != "mock1" {
		t.Errorf("Expected first provider 'mock1', got: %s", name)
	}

	// Drain events
	for event := range events {
		_ = event
	}
}

func TestProviderManagerFallsBackOnError(t *testing.T) {
	mgr := NewProviderManager()
	mgr.Add("fail1", &mockProvider{name: "fail1", shouldErr: true, errMsg: "connection refused"})
	mgr.Add("success", &mockProvider{name: "success", shouldErr: false})

	events, name, err := mgr.Stream(context.Background(), Model{Name: "test"}, nil, nil, StreamingOptions{})
	if err != nil {
		t.Fatalf("Expected no error (fallback succeeded), got: %v", err)
	}

	if name != "success" {
		t.Errorf("Expected fallback provider 'success', got: %s", name)
	}

	// Drain events
	for event := range events {
		_ = event
	}
}

func TestProviderManagerAllFail(t *testing.T) {
	mgr := NewProviderManager()
	mgr.Add("fail1", &mockProvider{name: "fail1", shouldErr: true, errMsg: "error1"})
	mgr.Add("fail2", &mockProvider{name: "fail2", shouldErr: true, errMsg: "error2"})

	events, name, err := mgr.Stream(context.Background(), Model{Name: "test"}, nil, nil, StreamingOptions{})
	if events != nil {
		t.Error("Expected nil events channel on all-fail")
	}
	if name != "" {
		t.Errorf("Expected empty provider name on all-fail, got: %s", name)
	}

	var allErr *AllProvidersFailedError
	if errors.As(err, &allErr) {
		if len(allErr.Providers) != 2 {
			t.Errorf("Expected 2 providers in error, got: %d", len(allErr.Providers))
		}
	} else {
		t.Errorf("Expected AllProvidersFailedError, got: %T", err)
	}
}

func TestProviderManagerNoProviders(t *testing.T) {
	mgr := NewProviderManager()

	events, name, err := mgr.Stream(context.Background(), Model{Name: "test"}, nil, nil, StreamingOptions{})
	if events != nil {
		t.Error("Expected nil events channel with no providers")
	}
	if name != "" {
		t.Errorf("Expected empty provider name, got: %s", name)
	}

	var noProvidersErr *NoProvidersConfiguredError
	if !errors.As(err, &noProvidersErr) {
		t.Errorf("Expected NoProvidersConfiguredError, got: %T", err)
	}
}

func TestAllProvidersFailedError(t *testing.T) {
	err := &AllProvidersFailedError{
		Providers: []string{"p1", "p2"},
		LastError: errors.New("last"),
	}

	expected := "all providers failed: p1, p2 — last error: last"
	if err.Error() != expected {
		t.Errorf("Error string mismatch:\ngot:  %s\nwant: %s", err.Error(), expected)
	}
}

func TestNoProvidersConfiguredError(t *testing.T) {
	err := &NoProvidersConfiguredError{}
	expected := "no LLM providers configured"
	if err.Error() != expected {
		t.Errorf("Error string mismatch: got %s, want %s", err.Error(), expected)
	}
}
