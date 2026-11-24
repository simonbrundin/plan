package handlers

import (
	"os"
	"path/filepath"

	"plan-agent/internal/api/db"
	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/tools"
)

// providerConfig holds configuration for an LLM provider
type providerConfig struct {
	name         string
	keyEnvVar    string
	modelEnvVar  string
	defaultModel string
	urlEnvVar    string
	defaultURL   string
}

var providerPriority = []providerConfig{
	{"agentgateway", "AGENTGATEWAY_API_KEY", "AGENTGATEWAY_MODEL", "gpt-4o-mini", "AGENTGATEWAY_URL", "https://api.agentgateway.dev/v1"},
	{"minimax", "MINIMAX_API_KEY", "MINIMAX_MODEL", "MiniMax-M2.7", "", ""},
	{"openai", "OPENAI_API_KEY", "OPENAI_MODEL", "gpt-4o-mini", "", ""},
	{"anthropic", "ANTHROPIC_API_KEY", "ANTHROPIC_MODEL", "claude-sonnet-4-7-25", "", ""},
}

// selectedProvider holds the resolved provider configuration
type selectedProvider struct {
	name   string
	apiKey string
	model  string
	url    string
}

func selectProvider(keyMap, modelMap map[string]string, getenv func(string) string) *selectedProvider {
	for _, cfg := range providerPriority {
		if key, ok := keyMap[cfg.name]; ok && key != "" {
			return buildSelected(cfg, key, modelMap[cfg.name], getenv)
		}
		if key := getenv(cfg.keyEnvVar); key != "" {
			return buildSelected(cfg, key, getenv(cfg.modelEnvVar), getenv)
		}
	}
	return nil
}

func buildSelected(cfg providerConfig, apiKey, modelHint string, getenv func(string) string) *selectedProvider {
	model := modelHint
	if model == "" {
		model = cfg.defaultModel
	}
	url := getenv(cfg.urlEnvVar)
	if url == "" {
		url = cfg.defaultURL
	}
	return &selectedProvider{name: cfg.name, apiKey: apiKey, model: model, url: url}
}

// getDefaultTools returns no shell or file tools for interactive chat.
// Chat is used for goal clarification and planning. Code changes happen in a
// separate, isolated AgentRun workspace instead of inside the API container.
func getDefaultTools() []*tools.ToolDefinition {
	return nil
}

// NoAPIKeyError indicates no API key is configured
type NoAPIKeyError struct{}

func (e *NoAPIKeyError) Error() string {
	return "no API key configured: add one in settings or set MINIMAX_API_KEY / OPENAI_API_KEY environment variable"
}

func createProvider(name, url string) provider.LLMProvider {
	switch name {
	case "agentgateway":
		if url == "" {
			url = "https://api.agentgateway.dev/v1"
		}
		return provider.NewAgentGatewayProvider(url)
	case "minimax":
		return provider.NewMinimaxProvider()
	default:
		return provider.NewOpenAIProvider()
	}
}

// defaultSystemPrompt is the default system prompt for the agent
const defaultSystemPrompt = `You are a helpful AI assistant for planning work.

Be concise and helpful. Respond in the same language as the user.

Interactive chat is for understanding and planning a goal. You may create sub-goals
only when the user explicitly asks you to add them; use the create_subgoal tool for
that instead of merely describing them. Do not invent repository state, claim that
files were changed, or pretend that tests were run. Ask focused questions when
important information is missing. A separate isolated AgentRun is responsible for
repository inspection and code changes after the plan is approved.`

// loadProjectSystemPrompt loads project-specific instructions from AGENTS.md
// if present in the current working directory or any parent directory.
// Returns "" if no file is found.
func loadProjectSystemPrompt() string {
	dir, err := os.Getwd()
	if err != nil {
		return ""
	}

	// Walk up from current directory to find AGENTS.md
	for {
		path := filepath.Join(dir, "AGENTS.md")
		if data, err := os.ReadFile(path); err == nil && len(data) > 0 {
			return "\n\n--- Project Instructions (AGENTS.md) ---\n" + string(data)
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached root
		}
		dir = parent
	}
	return ""
}

// buildSystemPrompt returns the default system prompt augmented with project-specific
// instructions from AGENTS.md if available.
func buildSystemPrompt() string {
	prompt := defaultSystemPrompt
	if projectContext := loadProjectSystemPrompt(); projectContext != "" {
		prompt += projectContext
		prompt += "\n\nIMPORTANT: The project instructions above take precedence over " +
			"the general instructions when they conflict."
	}
	return prompt
}

// dbMessagesToAgentMessages converts database message rows to agent message types
func dbMessagesToAgentMessages(dbMsgs []*db.Message) []messages.Message {
	result := make([]messages.Message, 0, len(dbMsgs))
	for _, m := range dbMsgs {
		// Skip tool messages - they're already executed, don't re-send to LLM
		if m.Role == "tool_request" || m.Role == "tool_result" || m.Role == "thinking" {
			continue
		}
		switch m.Role {
		case "user":
			result = append(result, &messages.UserMessage{
				Role:      messages.RoleUser,
				Content:   []messages.ContentBlock{&messages.TextBlock{Text: m.Content}},
				Timestamp: m.CreatedAt.UnixMilli(),
			})
		case "assistant":
			result = append(result, &messages.AssistantMessage{
				Role:      messages.RoleAssistant,
				Content:   []messages.ContentBlock{&messages.TextBlock{Text: m.Content}},
				Timestamp: m.CreatedAt.UnixMilli(),
			})
		case "system":
			result = append(result, &messages.SystemMessage{
				Role:      messages.RoleSystem,
				Content:   []messages.ContentBlock{&messages.TextBlock{Text: m.Content}},
				Timestamp: m.CreatedAt.UnixMilli(),
			})
		}
	}
	return result
}
