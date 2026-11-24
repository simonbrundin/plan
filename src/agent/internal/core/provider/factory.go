package provider

import (
	"fmt"
	"net/http"
	"os"
	"time"
)

// Factory creates LLM providers based on environment configuration
type Factory struct{}

// NewFactory creates a new provider factory
func NewFactory() *Factory {
	return &Factory{}
}

// CreateFromEnv creates a provider based on environment variables
// Looks for MINIMAX_API_KEY, OPENAI_API_KEY, etc.
func (f *Factory) CreateFromEnv() (LLMProvider, Model, error) {
	// Check for MiniMax first
	if apiKey := os.Getenv("MINIMAX_API_KEY"); apiKey != "" {
		modelName := os.Getenv("MINIMAX_MODEL")
		if modelName == "" {
			modelName = "MiniMax-M2.7"
		}
		endpoint := os.Getenv("MINIMAX_ENDPOINT")
		if endpoint == "" {
			endpoint = "https://api.minimax.io/v1"
		}
		p := &MinimaxProvider{
			httpClient: &http.Client{Timeout: 120 * time.Second},
			endpoint:   endpoint,
		}
		return p, Model{
			Provider: "minimax",
			Name:     modelName,
			APIKey:   apiKey,
		}, nil
	}

	// Check for OpenAI
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		modelName := os.Getenv("OPENAI_MODEL")
		if modelName == "" {
			modelName = "gpt-4o"
		}
		endpoint := os.Getenv("OPENAI_BASE_URL")
		if endpoint == "" {
			endpoint = "https://api.openai.com/v1"
		}
		p := &OpenAIProvider{
			httpClient: &http.Client{Timeout: 120 * time.Second},
			endpoint:   endpoint,
		}
		return p, Model{
			Provider: "openai",
			Name:     modelName,
			APIKey:   apiKey,
		}, nil
	}

	return nil, Model{}, fmt.Errorf("no LLM provider configured: set MINIMAX_API_KEY or OPENAI_API_KEY")
}

// CreateProviderManager creates a ProviderManager with all configured providers
func (f *Factory) CreateProviderManager() (*ProviderManager, error) {
	pm := NewProviderManager()

	// Add MiniMax if configured
	if apiKey := os.Getenv("MINIMAX_API_KEY"); apiKey != "" {
		endpoint := os.Getenv("MINIMAX_ENDPOINT")
		if endpoint == "" {
			endpoint = "https://api.minimax.io/v1"
		}
		p := &MinimaxProvider{
			httpClient: &http.Client{Timeout: 120 * time.Second},
			endpoint:   endpoint,
		}
		pm.Add("minimax", p)
	}

	// Add OpenAI if configured
	if apiKey := os.Getenv("OPENAI_API_KEY"); apiKey != "" {
		endpoint := os.Getenv("OPENAI_BASE_URL")
		if endpoint == "" {
			endpoint = "https://api.openai.com/v1"
		}
		p := &OpenAIProvider{
			httpClient: &http.Client{Timeout: 120 * time.Second},
			endpoint:   endpoint,
		}
		pm.Add("openai", p)
	}

	if len(pm.providers) == 0 {
		return nil, fmt.Errorf("no LLM providers configured")
	}

	return pm, nil
}
