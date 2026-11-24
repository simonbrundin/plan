package definer

import (
	"context"
	"fmt"
	"strings"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
)

// Definition represents a goal definition
type Definition struct {
	Title       string
	Description string
}

// Definer uses an LLM to define goals
type Definer struct {
	provider provider.LLMProvider
	model    provider.Model
}

// New creates a new goal definer
func New(provider provider.LLMProvider, model provider.Model) *Definer {
	return &Definer{
		provider: provider,
		model:    model,
	}
}

// Define creates a goal definition from an initial goal title
func (d *Definer) Define(ctx context.Context, goalTitle, existingDescription, feedback string) (*Definition, error) {
	systemPrompt := `You are a software requirements analyst. Your job is to take a loose goal or idea and turn it into a well-defined goal with:
1. A clear, concise title (max 50 characters)
2. A detailed description that covers:
   - What the goal accomplishes
   - Key requirements
   - Success criteria
   - Potential scope (if not obvious)

Be specific and actionable. Avoid vague language like "improve" or "optimize" without specifics.`

	userPrompt := buildUserPrompt(goalTitle, existingDescription, feedback)

	msgs := []messages.Message{
		&messages.SystemMessage{
			Role:      messages.RoleSystem,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: systemPrompt}},
			Timestamp: 0,
		},
		&messages.UserMessage{
			Role:      messages.RoleUser,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: userPrompt}},
			Timestamp: 0,
		},
	}

	stream, err := d.provider.Stream(ctx, d.model, msgs, nil, provider.StreamingOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to stream response: %w", err)
	}

	var response strings.Builder
	for event := range stream {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response.WriteString(e.Delta)
		case *provider.DoneEvent:
			return parseDefinition(response.String())
		case *provider.ErrorEvent:
			return nil, e.Error
		}
	}

	return nil, fmt.Errorf("stream ended without completion")
}

func buildUserPrompt(title, description, feedback string) string {
	var sb strings.Builder

	sb.WriteString("## GOAL TO DEFINE\n\n")
	sb.WriteString(title)
	sb.WriteString("\n\n")

	if description != "" {
		sb.WriteString("## EXISTING DESCRIPTION\n\n")
		sb.WriteString(description)
		sb.WriteString("\n\n")
	}

	if feedback != "" {
		sb.WriteString("## PREVIOUS FEEDBACK (please address this)\n\n")
		sb.WriteString(feedback)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## YOUR TASK\n\n")
	sb.WriteString("Provide a refined title and detailed description. Format your response as:\n\n")
	sb.WriteString("TITLE: [your refined title]\n\n")
	sb.WriteString("DESCRIPTION: [your detailed description]\n\n")
	sb.WriteString("Be specific and actionable. If the goal is unclear, make reasonable assumptions and state them.")

	return sb.String()
}

func parseDefinition(response string) (*Definition, error) {
	// Try to extract title and description from structured response
	lines := strings.Split(response, "\n")

	var title, description string
	inDescription := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "TITLE:") {
			title = strings.TrimPrefix(trimmed, "TITLE:")
			title = strings.TrimSpace(title)
			continue
		}

		if strings.HasPrefix(trimmed, "DESCRIPTION:") {
			description = strings.TrimPrefix(trimmed, "DESCRIPTION:")
			description = strings.TrimSpace(description)
			inDescription = true
			continue
		}

		if inDescription {
			// Check for another section header
			if strings.HasPrefix(trimmed, "TITLE:") || strings.HasPrefix(trimmed, "ANALYSIS:") {
				break
			}
			description += " " + trimmed
		}
	}

	// Fallback if structured parsing failed
	if title == "" && description == "" {
		// Use entire response as description
		parts := strings.SplitN(response, "\n", 2)
		title = parts[0]
		if len(parts) > 1 {
			description = parts[1]
		}
	}

	if title == "" {
		return nil, fmt.Errorf("failed to parse title from LLM response")
	}

	// Truncate title if too long
	if len(title) > 50 {
		title = title[:47] + "..."
	}

	return &Definition{
		Title:       title,
		Description: strings.TrimSpace(description),
	}, nil
}
