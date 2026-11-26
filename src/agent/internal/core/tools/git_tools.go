package tools

import (
	"context"
	"fmt"

	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
)

// Public aliases for backward compatibility with existing tests
var (
	GitStatusTool = gitStatusTool
	GitBranchTool = gitBranchTool
	GitCommitTool = gitCommitTool
	GitPushTool   = gitPushTool
	GitPRTool     = gitPRTool
	GitLogTool    = gitLogTool
)

// GitToolDefinition returns a tool definition by name
func GitToolDefinition(name string) *ToolDefinition {
	switch name {
	case "git_status":
		return gitStatusTool()
	case "git_branch":
		return gitBranchTool()
	case "git_commit":
		return gitCommitTool()
	case "git_push":
		return gitPushTool()
	case "git_pr":
		return gitPRTool()
	case "git_log":
		return gitLogTool()
	default:
		return nil
	}
}

// AllGitTools returns all git tool definitions
func AllGitTools() []*ToolDefinition {
	return []*ToolDefinition{
		gitStatusTool(),
		gitBranchTool(),
		gitCommitTool(),
		gitPushTool(),
		gitPRTool(),
		gitLogTool(),
	}
}

// Tool definition functions

func gitStatusTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_status",
		Description: "Show git working tree status. Returns list of changed files.",
		InputSchema: nil,
		Execute:     gitStatusExecute,
	}
}

func gitBranchTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_branch",
		Description: "Create a new git branch. Returns the branch name.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"name": map[string]any{
					"type":        "string",
					"description": "Branch name to create",
				},
			},
			"required": []any{"name"},
		},
		Execute: gitBranchExecute,
	}
}

func gitCommitTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_commit",
		Description: "Commit staged changes with a message.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"message": map[string]any{
					"type":        "string",
					"description": "Commit message",
				},
			},
			"required": []any{"message"},
		},
		Execute: gitCommitExecute,
	}
}

func gitPushTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_push",
		Description: "Push current branch to remote repository.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"remote": map[string]any{
					"type":        "string",
					"description": "Remote name (default: origin)",
					"default":     "origin",
				},
			},
		},
		Execute: gitPushExecute,
	}
}

func gitPRTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_pr",
		Description: "Create a pull request using gh CLI. Returns PR URL.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"title": map[string]any{
					"type":        "string",
					"description": "PR title",
				},
				"body": map[string]any{
					"type":        "string",
					"description": "PR description",
				},
				"base": map[string]any{
					"type":        "string",
					"description": "Base branch (default: main)",
					"default":     "main",
				},
			},
			"required": []any{"title", "body"},
		},
		Execute: gitPRExecute,
	}
}

func gitLogTool() *ToolDefinition {
	return &ToolDefinition{
		Name:        "git_log",
		Description: "Show recent git commits.",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"count": map[string]any{
					"type":        "number",
					"description": "Number of commits to show (default: 10)",
					"default":     10,
				},
			},
		},
		Execute: gitLogExecute,
	}
}

// Executor functions

func gitStatusExecute(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		return errorResult("Not in a git repository"), nil
	}

	status, err := git.Status(ctx)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	if status == "" {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: "Working tree clean"}},
		}, nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: "Changed files:\n" + status}},
	}, nil
}

func gitBranchExecute(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	name, ok := args["name"].(string)
	if !ok || name == "" {
		return errorResult("branch name is required"), nil
	}

	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		return errorResult("Not in a git repository"), nil
	}

	current, _ := git.CurrentBranch(ctx)
	if current == name {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Already on branch '%s'", name)}},
		}, nil
	}

	if err := git.CreateBranch(ctx, name); err != nil {
		return errorResult(err.Error()), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Created and switched to branch '%s'", name)}},
	}, nil
}

func gitCommitExecute(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	message, ok := args["message"].(string)
	if !ok || message == "" {
		return errorResult("commit message is required"), nil
	}

	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		return errorResult("Not in a git repository"), nil
	}

	hasChanges, err := git.HasChanges(ctx)
	if err != nil {
		return errorResult(err.Error()), nil
	}
	if !hasChanges {
		return &provider.ToolResult{
			Content: []messages.ContentBlock{&messages.TextBlock{Text: "Nothing to commit"}},
		}, nil
	}

	if err := git.Commit(ctx, message); err != nil {
		return errorResult(err.Error()), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Committed: %s", message)}},
	}, nil
}

func gitPushExecute(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	remote := "origin"
	if r, ok := args["remote"].(string); ok && r != "" {
		remote = r
	}

	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		return errorResult("Not in a git repository"), nil
	}

	branch, err := git.CurrentBranch(ctx)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	if err := git.Push(ctx, remote); err != nil {
		return errorResult(err.Error()), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Pushed '%s' to %s", branch, remote)}},
	}, nil
}

func gitPRExecute(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	title, ok := args["title"].(string)
	if !ok || title == "" {
		return errorResult("PR title is required"), nil
	}
	body := ""
	if b, ok := args["body"].(string); ok {
		body = b
	}
	base := "main"
	if b, ok := args["base"].(string); ok && b != "" {
		base = b
	}

	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		return errorResult("Not in a git repository"), nil
	}

	url, err := git.CreatePR(ctx, title, body, base)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: fmt.Sprintf("Created PR: %s", url)}},
	}, nil
}

func gitLogExecute(_ string, args map[string]any, _ <-chan struct{}, _ func(partial *provider.ToolResult)) (*provider.ToolResult, error) {
	count := 10
	if c, ok := args["count"].(float64); ok {
		count = int(c)
	}

	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		return errorResult("Not in a git repository"), nil
	}

	log, err := git.Log(ctx, count)
	if err != nil {
		return errorResult(err.Error()), nil
	}

	return &provider.ToolResult{
		Content: []messages.ContentBlock{&messages.TextBlock{Text: "Recent commits:\n" + log}},
	}, nil
}
