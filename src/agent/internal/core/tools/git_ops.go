package tools

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

// GitTool provides git operations for the agent
type GitTool struct{}

// NewGitTool creates a new git tool
func NewGitTool() *GitTool {
	return &GitTool{}
}

// CreateBranch creates a new git branch
func (t *GitTool) CreateBranch(ctx context.Context, name string) error {
	cmd := exec.CommandContext(ctx, "git", "checkout", "-b", name)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create branch %s: %w\n%s", name, err, out)
	}
	return nil
}

// CurrentBranch returns the current git branch
func (t *GitTool) CurrentBranch(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "--show-current")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get current branch: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

// Status returns the git status
func (t *GitTool) Status(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "status", "--short")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get status: %w", err)
	}
	return string(out), nil
}

// HasChanges returns true if there are uncommitted changes
func (t *GitTool) HasChanges(ctx context.Context) (bool, error) {
	status, err := t.Status(ctx)
	if err != nil {
		return false, err
	}
	return len(status) > 0, nil
}

// Commit commits changes with a message
func (t *GitTool) Commit(ctx context.Context, message string) error {
	cmd := exec.CommandContext(ctx, "git", "add", "-A")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to stage changes: %w\n%s", err, out)
	}

	cmd = exec.CommandContext(ctx, "git", "commit", "-m", message)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to commit: %w\n%s", err, out)
	}
	return nil
}

// Push pushes the current branch
func (t *GitTool) Push(ctx context.Context, remote string) error {
	cmd := exec.CommandContext(ctx, "git", "push", "-u", remote)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to push: %w\n%s", err, out)
	}
	return nil
}

// CreatePR creates a pull request using gh CLI
func (t *GitTool) CreatePR(ctx context.Context, title, body, base string) (string, error) {
	cmd := exec.CommandContext(ctx, "gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--base", base,
		"--fill")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("failed to create PR: %w\n%s", err, out)
	}

	return strings.TrimSpace(string(out)), nil
}

// ListBranches lists all local and remote branches
func (t *GitTool) ListBranches(ctx context.Context) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "branch", "-a", "--format=%(refname:short)")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	return strings.Split(strings.TrimSpace(string(out)), "\n"), nil
}

// Diff returns the diff of changes
func (t *GitTool) Diff(ctx context.Context) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "diff", "--stat")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get diff: %w", err)
	}
	return string(out), nil
}

// Log returns the git log
func (t *GitTool) Log(ctx context.Context, n int) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "log", fmt.Sprintf("-%d", n), "--oneline")
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get log: %w", err)
	}
	return string(out), nil
}

// IsGitRepo returns true if we're in a git repository
func (t *GitTool) IsGitRepo(ctx context.Context) bool {
	cmd := exec.CommandContext(ctx, "git", "rev-parse", "--is-inside-work-tree")
	return cmd.Run() == nil
}
