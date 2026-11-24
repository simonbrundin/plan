package workspace

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestNewKopierarRepot kontrollerar att arbetsytan är en kopia och att
// originalet är orört.
func TestNewKopierarRepot(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")
	writeFile(t, filepath.Join(source, "README.md"), "# hej\n")

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	content, err := ws.Read("main.go")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if content != "package main\n" {
		t.Errorf("filinnehåll = %q, vill ha originalet", content)
	}

	// Originalet ska vara orört - ändringen får inte läcka tillbaka.
	if err := ws.Write("main.go", "package main // ändrad\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	original, err := os.ReadFile(filepath.Join(source, "main.go"))
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(original) != "package main\n" {
		t.Errorf("originalet ändrades: %q", string(original))
	}
}

// TestWriteSkaparKataloger kontrollerar att nya filer kan placeras i nya mappar.
func TestWriteSkaparKataloger(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	if err := ws.Write("pkg/sub/fil.go", "package sub\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if !ws.Exists("pkg/sub/fil.go") {
		t.Error("filen skapades inte")
	}
	content, err := ws.Read("pkg/sub/fil.go")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if content != "package sub\n" {
		t.Errorf("innehåll = %q", content)
	}
}

// TestWriteBlockerarUtväg kontrollerar att agenten inte kan skriva utanför
// arbetsytan.
func TestWriteBlockerarUtväg(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	for _, path := range []string{"../ut.txt", "/tmp/ut.txt", "a/../../ut.txt"} {
		if err := ws.Write(path, "x"); err == nil {
			t.Errorf("sökvägen %q borde ha avvisats", path)
		}
	}
}

// TestChangedFilesHittarÄndringar kontrollerar att diffen visar vad som ändrats.
func TestChangedFilesHittarÄndringar(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")
	writeFile(t, filepath.Join(source, "stable.go"), "package main\n")

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	if err := ws.Write("main.go", "package main // ändrad\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}
	if err := ws.Write("new.go", "package main\n"); err != nil {
		t.Fatalf("Write: %v", err)
	}

	changed, err := ws.ChangedFiles()
	if err != nil {
		t.Fatalf("ChangedFiles: %v", err)
	}

	found := map[string]bool{}
	for _, file := range changed {
		found[file] = true
	}

	if !found["main.go"] {
		t.Error("main.go saknas i ändringslistan")
	}
	if !found["new.go"] {
		t.Error("new.go saknas i ändringslistan")
	}
	if found["stable.go"] {
		t.Error("stable.go ska inte vara med - den är oförändrad")
	}
}

// TestRunKommando kontrollerar att kommandon kör i arbetsytan.
func TestRunKommando(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "fil.txt"), "innehåll\n")

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	output, err := ws.Run("cat", "fil.txt")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if output != "innehåll\n" {
		t.Errorf("utdata = %q", output)
	}
}

// TestRunFelkommando kontrollerar att felkommandon ger ett fel med utdata.
func TestRunFelkommando(t *testing.T) {
	source := t.TempDir()

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	if _, err := ws.Run("cat", "saknas.txt"); err == nil {
		t.Error("ett felande kommando ska ge ett fel")
	}
}

// run kör ett git-kommando i en katalog.
func run(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("%v: %v: %s", args, err, string(output))
	}
}

// initRepo skapar ett git-repo med en identitet och en första commit.
//
// Identiteten sätts lokalt i repot - testerna ska inte kräva en global
// git-identitet, annars faller de på en ren CI-runner.
func initRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	run(t, dir, "git", "init", "--initial-branch=main")
	run(t, dir, "git", "config", "user.email", "test@example.com")
	run(t, dir, "git", "config", "user.name", "Test")
	run(t, dir, "git", "commit", "--allow-empty", "-m", "start")
	return dir
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestParseGitStatusBehållerSökvägar kontrollerar att filnamnen inte
// förkortas när git status läses.
func TestParseGitStatusBehållerSökvägar(t *testing.T) {
	output := " M calculator.go\n M calculator_test.go\n?? nyfil.go\n"

	files := parseGitStatus(output)

	want := []string{"calculator.go", "calculator_test.go", "nyfil.go"}
	if len(files) != len(want) {
		t.Fatalf("antal filer = %d, vill ha %d: %v", len(files), len(want), files)
	}
	for i := range want {
		if files[i] != want[i] {
			t.Errorf("fil %d = %q, vill ha %q", i, files[i], want[i])
		}
	}
}

// TestWorkspaceÅterställerRemote kontrollerar att arbetsytan pekar mot
// källans remote.
//
// Arbetsytan klonas lokalt - utan återställd remote når pushen aldrig
// det riktiga repot.
func TestWorkspaceÅterställerRemote(t *testing.T) {
	source := initRepo(t)
	run(t, source, "git", "remote", "add", "origin", "https://github.com/simonbrundin/plan.git")

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	output, err := ws.Run("git", "remote", "get-url", "origin")
	if err != nil {
		t.Fatalf("read remote: %v", err)
	}
	if !strings.Contains(output, "github.com/simonbrundin/plan.git") {
		t.Errorf("remote = %q, vill ha källans remote", output)
	}
}

// TestWorkspaceSätterIdentitet kontrollerar att agentens commits kan
// attribueras.
func TestWorkspaceSätterIdentitet(t *testing.T) {
	source := initRepo(t)

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	email, err := ws.Run("git", "config", "user.email")
	if err != nil {
		t.Fatalf("read user.email: %v", err)
	}
	if strings.TrimSpace(email) != "agent@plan.local" {
		t.Errorf("user.email = %q", email)
	}
}

// TestPushBranchAnvänderRättFlaggor kontrollerar att auth-flaggan sätts
// före subkommandot.
//
// "git push -c ..." är inte giltigt - flaggan hör till git, inte till push.
func TestPushBranchAnvänderRättFlaggor(t *testing.T) {
	source := initRepo(t)

	ws, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = ws.Cleanup() }()

	t.Setenv("GIT_AUTH_HEADER", "Authorization: Basic secret")

	err = ws.pushBranch("agent/goal-1")
	if err == nil {
		t.Fatal("pushen ska misslyckas - det finns ingen remote att skicka till")
	}
	if strings.Contains(err.Error(), "usage:") {
		t.Errorf("git-flaggorna är fel: %v", err)
	}
}
