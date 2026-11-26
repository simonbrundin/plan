package goals

import (
	"context"
	"fmt"
	"strings"
	"time"

	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
	"plan/agent/internal/core/tools"
)

// ToolName constants
const (
	ToolNameListGoals    = "list_goals"
	ToolNameGetGoal      = "get_goal"
	ToolNameCreateGoal   = "create_goal"
	ToolNameUpdateGoal   = "update_goal"
	ToolNameDeleteGoal   = "delete_goal"
	ToolNameListStatuses = "list_goal_statuses"
)

// DefaultTools returns the goal tools using the default singleton store
func DefaultTools(ctx context.Context) ([]*tools.ToolDefinition, error) {
	store, err := DefaultStore(ctx)
	if err != nil {
		return nil, err
	}
	return Tools(store), nil
}

// Tools returns the goal-related tools for the agent
func Tools(store Store) []*tools.ToolDefinition {
	return []*tools.ToolDefinition{
		{
			Name:        ToolNameListGoals,
			Description: "List all goals in the system. Use this to see what goals exist and their status.",
			InputSchema: listGoalsInputSchema(),
			Execute:     listGoalsExecute(store),
		},
		{
			Name:        ToolNameGetGoal,
			Description: "Get detailed information about a specific goal including its children.",
			InputSchema: getGoalInputSchema(),
			Execute:     getGoalExecute(store),
		},
		{
			Name:        ToolNameCreateGoal,
			Description: "Create a new goal. Use this when a user wants to add a new task or sub-goal.",
			InputSchema: createGoalInputSchema(),
			Execute:     createGoalExecute(store),
		},
		{
			Name:        ToolNameUpdateGoal,
			Description: "Update an existing goal's title, status, or other fields.",
			InputSchema: updateGoalInputSchema(),
			Execute:     updateGoalExecute(store),
		},
		{
			Name:        ToolNameDeleteGoal,
			Description: "Delete a goal by ID.",
			InputSchema: deleteGoalInputSchema(),
			Execute:     deleteGoalExecute(store),
		},
		{
			Name:        ToolNameListStatuses,
			Description: "List all available goal statuses.",
			InputSchema: listStatusesInputSchema(),
			Execute:     listStatusesExecute(store),
		},
	}
}

// ToolFunc is the function signature for tool implementations
type ToolFunc = tools.ToolFunc

func listGoalsInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"parent_id": map[string]any{
				"type":        "string",
				"description": "Filter goals by parent ID. Omit to get all goals.",
			},
		},
	}
}

func getGoalInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "The goal ID to retrieve.",
			},
		},
		"required": []any{"id"},
	}
}

func createGoalInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "Unique ID for the goal (kebab-case recommended).",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "The goal title.",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "Optional description of the goal.",
			},
			"parent_id": map[string]any{
				"type":        "string",
				"description": "Parent goal ID if this is a sub-goal.",
			},
			"status": map[string]any{
				"type":        "string",
				"description": "Goal status: define, plan, execute, verify, completed.",
				"default":     "define",
			},
			"icon": map[string]any{
				"type":        "string",
				"description": "Emoji or icon for the goal.",
				"default":     "📌",
			},
		},
		"required": []any{"id", "title"},
	}
}

func updateGoalInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "The goal ID to update.",
			},
			"title": map[string]any{
				"type":        "string",
				"description": "New title for the goal.",
			},
			"description": map[string]any{
				"type":        "string",
				"description": "New description for the goal.",
			},
			"status": map[string]any{
				"type":        "string",
				"description": "New status for the goal.",
			},
			"icon": map[string]any{
				"type":        "string",
				"description": "New icon for the goal.",
			},
		},
		"required": []any{"id"},
	}
}

func listStatusesInputSchema() map[string]any {
	return map[string]any{
		"type":       "object",
		"properties": map[string]any{},
	}
}

func deleteGoalInputSchema() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"id": map[string]any{
				"type":        "string",
				"description": "The goal ID to delete.",
			},
		},
		"required": []any{"id"},
	}
}

// Execute functions

func listGoalsExecute(store Store) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		ctx := context.Background()

		var goals []*Goal
		var err error

		if parentID, ok := args["parent_id"].(string); ok && parentID != "" {
			goals, err = store.GetChildren(ctx, parentID)
		} else {
			goals, err = store.GetAll(ctx)
		}

		if err != nil {
			return errorResult(err.Error()), nil
		}

		if len(goals) == 0 {
			return &provider.ToolResult{
				Content: []messages.ContentBlock{&messages.TextBlock{Text: "No goals found."}},
			}, nil
		}

		var sb strings.Builder
		sb.WriteString("## Goals\n\n")
		for _, g := range goals {
			_, _ = fmt.Fprintf(&sb, "- [%s] %s (%s)", statusEmoji(g.Status), g.Title, g.Status)
			if g.Description != "" {
				_, _ = fmt.Fprintf(&sb, "\n  %s", g.Description)
			}
			sb.WriteString("\n")
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: sb.String()}},
		}, nil
	}
}

func getGoalExecute(store Store) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		ctx := context.Background()

		id, ok := args["id"].(string)
		if !ok || id == "" {
			return errorResult("id is required"), nil
		}

		goal, err := store.GetByID(ctx, id)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		children, _ := store.GetChildren(ctx, id)

		var sb strings.Builder
		_, _ = fmt.Fprintf(&sb, "## %s %s\n\n", goal.Icon, goal.Title)
		_, _ = fmt.Fprintf(&sb, "**ID:** `%s`\n", goal.ID)
		_, _ = fmt.Fprintf(&sb, "**Status:** %s\n", goal.Status)
		if goal.Description != "" {
			_, _ = fmt.Fprintf(&sb, "**Description:** %s\n", goal.Description)
		}

		if len(children) > 0 {
			sb.WriteString("\n### Sub-goals\n")
			for _, c := range children {
				_, _ = fmt.Fprintf(&sb, "- [%s] %s (%s)\n", statusEmoji(c.Status), c.Title, c.Status)
			}
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: sb.String()}},
		}, nil
	}
}

func createGoalExecute(store Store) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		ctx := context.Background()

		id, _ := args["id"].(string)
		title, _ := args["title"].(string)
		description, _ := args["description"].(string)
		parentID, _ := args["parent_id"].(string)
		status, _ := args["status"].(string)
		icon, _ := args["icon"].(string)

		if id == "" || title == "" {
			return errorResult("id and title are required"), nil
		}

		if status == "" {
			status = "define"
		}
		if icon == "" {
			icon = "📌"
		}

		goal := &Goal{
			ID:          id,
			Title:       title,
			Description: description,
			ParentID:    parentID,
			Status:      status,
			Icon:        icon,
			Created:     contextTime(ctx),
		}

		if err := store.Create(ctx, goal); err != nil {
			return errorResult(err.Error()), nil
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Created goal '%s' with ID '%s'", title, id)}},
		}, nil
	}
}

func updateGoalExecute(store Store) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		ctx := context.Background()

		id, _ := args["id"].(string)
		if id == "" {
			return errorResult("id is required"), nil
		}

		goal, err := store.GetByID(ctx, id)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		if title, ok := args["title"].(string); ok && title != "" {
			goal.Title = title
		}
		if description, ok := args["description"].(string); ok {
			goal.Description = description
		}
		if status, ok := args["status"].(string); ok && status != "" {
			goal.Status = status
		}
		if icon, ok := args["icon"].(string); ok && icon != "" {
			goal.Icon = icon
		}

		if err := store.Update(ctx, id, goal); err != nil {
			return errorResult(err.Error()), nil
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Updated goal '%s'", id)}},
		}, nil
	}
}

func listStatusesExecute(store Store) tools.ToolFunc {
	return func(_ string, _ map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		ctx := context.Background()

		statuses, err := store.GetStatuses(ctx)
		if err != nil {
			return errorResult(err.Error()), nil
		}

		var sb strings.Builder
		sb.WriteString("## Available Goal Statuses\n\n")
		for _, s := range statuses {
			_, _ = fmt.Fprintf(&sb, "- `%s` - %s\n", s.Name, s.Label)
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: sb.String()}},
		}, nil
	}
}

func deleteGoalExecute(store Store) tools.ToolFunc {
	return func(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
		ctx := context.Background()

		id, ok := args["id"].(string)
		if !ok || id == "" {
			return errorResult("id is required"), nil
		}

		if err := store.Delete(ctx, id); err != nil {
			return errorResult(err.Error()), nil
		}

		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Deleted goal '%s'", id)}},
		}, nil
	}
}

func statusEmoji(status string) string {
	switch status {
	case "define":
		return "📝"
	case "plan":
		return "📋"
	case "execute":
		return "⚡"
	case "verify":
		return "🔍"
	case "completed":
		return "✅"
	default:
		return "⬜"
	}
}

func errorResult(msg string) *provider.ToolResult {
	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: msg}},
		IsError: true,
	}
}

func contextTime(ctx context.Context) time.Time {
	return time.Now()
}
