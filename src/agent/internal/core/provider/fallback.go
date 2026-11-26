package provider

import (
	"context"
	"log"
	"strings"

	"plan/agent/internal/core/messages"
)

// ProviderManager manages multiple providers with automatic fallback
type ProviderManager struct {
	providers []LLMProvider
	names     []string
}

// NewProviderManager creates a manager with providers in priority order
func NewProviderManager() *ProviderManager {
	return &ProviderManager{
		providers: make([]LLMProvider, 0),
		names:     make([]string, 0),
	}
}

// Add adds a provider to the manager
func (m *ProviderManager) Add(name string, p LLMProvider) {
	m.providers = append(m.providers, p)
	m.names = append(m.names, name)
}

// Stream attempts to stream from providers in order, falling back on errors
func (m *ProviderManager) Stream(
	ctx context.Context,
	model Model,
	messages []messages.Message,
	tools []ToolDefinition,
	opts StreamingOptions,
) (<-chan StreamEvent, string, error) {

	var lastErr error

	for i, p := range m.providers {
		name := m.names[i]

		events, err := p.Stream(ctx, model, messages, tools, opts)
		if err != nil {
			log.Printf("Provider %s failed: %v (trying next provider)", name, err)
			lastErr = err
			continue
		}

		log.Printf("Using provider: %s", name)
		return events, name, nil
	}

	if lastErr != nil {
		return nil, "", &AllProvidersFailedError{
			Providers: m.names,
			LastError: lastErr,
		}
	}

	return nil, "", &NoProvidersConfiguredError{}
}

// Complete returns a single response by collecting the stream
func (m *ProviderManager) Complete(
	ctx context.Context,
	model Model,
	messages []messages.Message,
) (string, error) {
	events, _, err := m.Stream(ctx, model, messages, nil, StreamingOptions{})
	if err != nil {
		return "", err
	}

	var response string
	for event := range events {
		switch e := event.(type) {
		case *TextDeltaEvent:
			response += e.Delta
		case *DoneEvent:
			return response, nil
		case *ErrorEvent:
			return response, e.Error
		}
	}

	return response, nil
}

// AllProvidersFailedError indicates all configured providers failed
type AllProvidersFailedError struct {
	Providers []string
	LastError error
}

func (e *AllProvidersFailedError) Error() string {
	return "all providers failed: " + strings.Join(e.Providers, ", ") + " — last error: " + e.LastError.Error()
}

// Unwrap returns the last error
func (e *AllProvidersFailedError) Unwrap() error {
	return e.LastError
}

// NoProvidersConfiguredError indicates no providers are configured
type NoProvidersConfiguredError struct{}

func (e *NoProvidersConfiguredError) Error() string {
	return "no LLM providers configured"
}
