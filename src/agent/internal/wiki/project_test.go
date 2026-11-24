package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectTag(t *testing.T) {
	if got := ProjectTag("typetrainer"); got != "projekt:typetrainer" {
		t.Errorf("ProjectTag = %q, want projekt:typetrainer", got)
	}
	if got := ProjectTag(""); got != "" {
		t.Errorf("ProjectTag(\"\") = %q, want empty", got)
	}
}

func TestNoteHasProject(t *testing.T) {
	note := &Note{ID: "a", Tags: []string{"bug", "projekt:typetrainer"}}

	if note.ProjectOf() != "typetrainer" {
		t.Errorf("ProjectOf = %q, want typetrainer", note.ProjectOf())
	}
	if !note.HasProject("typetrainer") {
		t.Error("HasProject(typetrainer) = false, want true")
	}
	if note.HasProject("dashboard") {
		t.Error("HasProject(dashboard) = true, want false")
	}
	if !note.HasProject("") {
		t.Error("HasProject(\"\") = false, want true (ingen filtrering)")
	}
}

func TestDetectProjectFindsRepoName(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	project := DetectProject(dir)
	if project == "" {
		t.Skip("git kunde inte köras i testmiljön")
	}
	if project != filepath.Base(dir) {
		t.Errorf("DetectProject = %q, want %q", project, filepath.Base(dir))
	}
}

func TestEnsureAgentsBlockCreatesFile(t *testing.T) {
	dir := t.TempDir()

	changed, err := EnsureAgentsBlock(dir)
	if err != nil {
		t.Fatalf("EnsureAgentsBlock: %v", err)
	}
	if !changed {
		t.Error("changed = false, want true")
	}

	content, err := os.ReadFile(filepath.Join(dir, AgentsFileName))
	if err != nil {
		t.Fatalf("AGENTS.md skapades inte: %v", err)
	}
	if !strings.Contains(string(content), AgentsBlockStart) {
		t.Errorf("AGENTS.md saknar startmarkör:\n%s", content)
	}
	if !strings.Contains(string(content), "wikictl note") {
		t.Errorf("AGENTS.md saknar kommandona:\n%s", content)
	}
	if strings.HasPrefix(string(content), "# <!--") {
		t.Errorf("AGENTS.md har trasig rubrik:\n%s", content)
	}
}

func TestEnsureAgentsBlockIsIdempotent(t *testing.T) {
	dir := t.TempDir()

	if _, err := EnsureAgentsBlock(dir); err != nil {
		t.Fatal(err)
	}
	first, err := os.ReadFile(filepath.Join(dir, AgentsFileName))
	if err != nil {
		t.Fatal(err)
	}

	changed, err := EnsureAgentsBlock(dir)
	if err != nil {
		t.Fatal(err)
	}
	if changed {
		t.Error("changed = true, want false (inget att ändra)")
	}

	second, err := os.ReadFile(filepath.Join(dir, AgentsFileName))
	if err != nil {
		t.Fatal(err)
	}
	if string(first) != string(second) {
		t.Error("filen ändrades vid andra körningen")
	}
}

func TestEnsureAgentsBlockKeepsExistingContent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, AgentsFileName)
	if err := os.WriteFile(path, []byte("# Eget projekt\n\nMin egen text.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureAgentsBlock(dir); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "Min egen text.") {
		t.Errorf("befintligt innehåll försvann:\n%s", content)
	}
	if !strings.Contains(string(content), AgentsBlockStart) {
		t.Errorf("wikiblocket saknas:\n%s", content)
	}
}

func TestEnsureAgentsBlockReplacesOldBlock(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, AgentsFileName)
	old := AgentsBlockStart + "\n GAMMAL TEXT\n" + AgentsBlockEnd
	if err := os.WriteFile(path, []byte("# Projekt\n\n"+old+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := EnsureAgentsBlock(dir); err != nil {
		t.Fatal(err)
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(content), "GAMMAL TEXT") {
		t.Errorf("gammalt block ersattes inte:\n%s", content)
	}
	if !strings.Contains(string(content), AgentsBlockStart) {
		t.Errorf("nytt block saknas:\n%s", content)
	}
	if strings.Count(string(content), AgentsBlockStart) != 1 {
		t.Errorf("startmarkören finns fler än en gång:\n%s", content)
	}
}
