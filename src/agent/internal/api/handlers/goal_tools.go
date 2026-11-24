package handlers

import (
	"context"
	"fmt"

	"plan-agent/internal/core/messages"
	"plan-agent/internal/core/provider"
	"plan-agent/internal/core/tools"
	"plan-agent/internal/planclient"
)

// createSubgoalTool lets an interactive goal chat persist a proposed child
// goal without granting the chat shell or arbitrary file-system access.
func createSubgoalTool(client *planclient.Client, parentID int64, recorder *runRecorder) *tools.ToolDefinition {
	return &tools.ToolDefinition{
		Name:        "create_subgoal",
		Description: "Create one concrete sub-goal under the current Plan goal. Use only when the user explicitly asks to add or create sub-goals.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "Short, actionable title for the sub-goal.",
				},
				"description": map[string]any{
					"type":        "string",
					"description": "Optional description and acceptance criteria.",
				},
			},
			"required": []any{"title"},
		},
		Execute: func(_ string, args map[string]any, _ <-chan struct{}, _ func(*provider.ToolResult)) (*provider.ToolResult, error) {
			title, ok := args["title"].(string)
			if !ok || title == "" {
				return toolError("title is required"), nil
			}
			description, _ := args["description"].(string)

			goal, err := client.CreateSubgoal(context.Background(), parentID, title, description)
			if err != nil {
				recorder.Record(context.Background(), planclient.EventError, "", map[string]any{
					"tool":  "create_subgoal",
					"error": err.Error(),
				})
				return toolError(err.Error()), nil
			}

			recorder.Record(context.Background(), planclient.EventSubgoalCreated, "", map[string]any{
				"goal_id":   goal.ID,
				"parent_id": parentID,
				"title":     goal.Title,
			})

			return &provider.ToolResult{Content: []messages.ContentBlock{
				&messages.TextBlock{Text: fmt.Sprintf("Created sub-goal #%d: %s", goal.ID, goal.Title)},
			}}, nil
		},
	}
}

func toolError(message string) *provider.ToolResult {
	return &provider.ToolResult{
		IsError: true,
		Content: []messages.ContentBlock{&messages.TextBlock{Text: "Error: " + message}},
	}
}
