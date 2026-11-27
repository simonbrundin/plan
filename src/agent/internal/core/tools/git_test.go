package tools

import (
	"context"
	"testing"
)

func TestGitToolIsGitRepo(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	// We're in a git repo during tests, so this should be true
	isRepo := git.IsGitRepo(ctx)
	if !isRepo {
		t.Log("Not in a git repo (expected in some test environments)")
	}
}

func TestGitToolCurrentBranch(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		t.Skip("Not in a git repo")
	}

	branch, err := git.CurrentBranch(ctx)
	if err != nil {
		t.Fatalf("CurrentBranch failed: %v", err)
	}

	// I ett utcheckat merge-commit - som CI gör - finns ingen branch.
	// Då är det inget att testa.
	if branch == "" {
		t.Skip("Detached HEAD - ingen branch")
	}

	t.Logf("Current branch: %s", branch)
}

func TestGitToolStatus(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		t.Skip("Not in a git repo")
	}

	status, err := git.Status(ctx)
	if err != nil {
		t.Fatalf("Status failed: %v", err)
	}

	t.Logf("Git status:\n%s", status)
}

func TestGitToolListBranches(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		t.Skip("Not in a git repo")
	}

	branches, err := git.ListBranches(ctx)
	if err != nil {
		t.Fatalf("ListBranches failed: %v", err)
	}

	if len(branches) == 0 {
		t.Error("Should have at least one branch")
	}

	t.Logf("Found %d branches", len(branches))
}

func TestGitToolLog(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		t.Skip("Not in a git repo")
	}

	log, err := git.Log(ctx, 5)
	if err != nil {
		t.Fatalf("Log failed: %v", err)
	}

	if log == "" {
		t.Error("Log should not be empty")
	}

	t.Logf("Recent commits:\n%s", log)
}

func TestGitToolHasChanges(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		t.Skip("Not in a git repo")
	}

	hasChanges, err := git.HasChanges(ctx)
	if err != nil {
		t.Fatalf("HasChanges failed: %v", err)
	}

	t.Logf("Has changes: %v", hasChanges)
}

func TestGitToolDiff(t *testing.T) {
	git := NewGitTool()
	ctx := context.Background()

	if !git.IsGitRepo(ctx) {
		t.Skip("Not in a git repo")
	}

	diff, err := git.Diff(ctx)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	t.Logf("Diff:\n%s", diff)
}

func TestCaptureCommand(t *testing.T) {
	ctx := context.Background()

	// Test successful command
	out, err := captureCommand(ctx, "echo", "hello")
	if err != nil {
		t.Fatalf("Command failed: %v", err)
	}
	if out.Stdout != "hello\n" {
		t.Errorf("Expected 'hello\n', got '%s'", out.Stdout)
	}
	if out.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", out.ExitCode)
	}
}

func TestGitStatusTool(t *testing.T) {
	tool := GitStatusTool()
	if tool.Name != "git_status" {
		t.Errorf("Expected name 'git_status', got '%s'", tool.Name)
	}
	if tool.InputSchema != nil {
		t.Error("git_status should have nil schema")
	}
}

func TestGitBranchTool(t *testing.T) {
	tool := GitBranchTool()
	if tool.Name != "git_branch" {
		t.Errorf("Expected name 'git_branch', got '%s'", tool.Name)
	}
	if tool.InputSchema == nil {
		t.Fatal("Expected non-nil schema")
	}
	props := tool.InputSchema["properties"].(map[string]any)
	if _, ok := props["name"]; !ok {
		t.Error("Expected 'name' property")
	}
}

func TestGitCommitTool(t *testing.T) {
	tool := GitCommitTool()
	if tool.Name != "git_commit" {
		t.Errorf("Expected name 'git_commit', got '%s'", tool.Name)
	}
}

func TestGitPushTool(t *testing.T) {
	tool := GitPushTool()
	if tool.Name != "git_push" {
		t.Errorf("Expected name 'git_push', got '%s'", tool.Name)
	}
}

func TestGitPRTool(t *testing.T) {
	tool := GitPRTool()
	if tool.Name != "git_pr" {
		t.Errorf("Expected name 'git_pr', got '%s'", tool.Name)
	}
}

func TestGitLogTool(t *testing.T) {
	tool := GitLogTool()
	if tool.Name != "git_log" {
		t.Errorf("Expected name 'git_log', got '%s'", tool.Name)
	}
}
