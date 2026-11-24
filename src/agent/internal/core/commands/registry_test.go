package commands

import (
	"testing"
)

func TestLoad(t *testing.T) {
	r, err := Load("/home/simon/repos/agents/src/commands")
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	cmds := r.List()
	if len(cmds) == 0 {
		t.Fatal("expected at least one command")
	}

	// Check a known command
	push := r.Get("push")
	if push == nil {
		t.Fatal("expected push command")
	}
	if push.Description == "" {
		t.Error("push command should have a description")
	}

	// Check ai-prime
	prime := r.Get("ai-prime")
	if prime == nil {
		t.Fatal("expected ai-prime command")
	}
	if prime.Description == "" {
		t.Error("ai-prime should have a description")
	}

	t.Logf("Loaded %d commands:", len(cmds))
	for _, c := range cmds {
		t.Logf("  • %s: %s", c.Name, c.Description)
	}
}

func TestLoadMissingDir(t *testing.T) {
	r, err := Load("/nonexistent/path")
	if err != nil {
		t.Fatal("should not error on missing dir")
	}
	if len(r.List()) != 0 {
		t.Error("should return empty registry for missing dir")
	}
}
