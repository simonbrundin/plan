package orchestrator

import (
	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/tools"
)

// AddLearningTool returns a tool for adding learnings
func AddLearningTool(o *Orchestrator) *tools.ToolDefinition {
	return &tools.ToolDefinition{
		Name:        "add_learning",
		Description: "Record a learning for future reference. Use when you discover something important, make a mistake, or find a better approach.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"category": map[string]any{
					"type":        "string",
					"description": "Category: test, bug, pattern, optimization, security, refactor, or general",
					"enum":        []any{"test", "bug", "pattern", "optimization", "security", "refactor", "general"},
				},
				"title": map[string]any{
					"type":        "string",
					"description": "Brief title for this learning",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "What happened or what you discovered",
				},
				"lesson": map[string]any{
					"type":        "string",
					"description": "The key takeaway or best practice to follow",
				},
			},
			"required": []any{"category", "title", "lesson"},
		},
		Execute: addLearningExecute(o),
	}
}

// GetLearningsTool returns a tool for retrieving learnings
func GetLearningsTool(o *Orchestrator) *tools.ToolDefinition {
	return &tools.ToolDefinition{
		Name:        "get_learnings",
		Description: "Retrieve all recorded learnings. Use this to learn from past mistakes and follow best practices.",
		InputSchema: nil,
		Execute:     getLearningsExecute(o),
	}
}

func addLearningExecute(o *Orchestrator) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		category := "general"
		if c, ok := args["category"].(string); ok {
			category = c
		}
		title := ""
		if t, ok := args["title"].(string); ok {
			title = t
		}
		description := ""
		if d, ok := args["description"].(string); ok {
			description = d
		}
		lesson := ""
		if l, ok := args["lesson"].(string); ok {
			lesson = l
		}

		if title == "" {
			return &provider.ToolResult{
				Content: []messages.ContentBlock{&messages.TextBlock{Text: "title is required"}},
				IsError: true,
			}, nil
		}

		if lesson == "" {
			return &provider.ToolResult{
				Content: []messages.ContentBlock{&messages.TextBlock{Text: "lesson is required"}},
				IsError: true,
			}, nil
		}

		err := o.AddLearning(category, title, description, lesson)
		if err != nil {
			return &provider.ToolResult{
				Content: []messages.ContentBlock{&messages.TextBlock{Text: "Failed to save learning: " + err.Error()}},
				IsError: true,
			}, nil
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: "Learning recorded: " + title}},
		}, nil
	}
}

func getLearningsExecute(o *Orchestrator) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		// Generate prompt context (summarized learnings)
		context := o.GenerateLearningsContext()
		if context == "" {
			return &provider.ToolResult{
				Content: []messages.ContentBlock{&messages.TextBlock{Text: "No learnings recorded yet."}},
			}, nil
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: context}},
		}, nil
	}
}
