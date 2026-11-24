package provider

import (
	"net/http"
	"time"
)

// AgentGatewayProvider implements LLMProvider for AgentGateway
// AgentGateway is an OpenAI-compatible gateway that routes to multiple providers
type AgentGatewayProvider struct {
	*OpenAIProvider // Embed OpenAIProvider for base functionality
}

// NewAgentGatewayProvider creates a new AgentGateway provider
func NewAgentGatewayProvider(gatewayURL string) *AgentGatewayProvider {
	return &AgentGatewayProvider{
		OpenAIProvider: &OpenAIProvider{
			httpClient: &http.Client{Timeout: 120 * time.Second},
			endpoint:   gatewayURL,
		},
	}
}

func (p *AgentGatewayProvider) Name() string {
	return "agentgateway"
}
