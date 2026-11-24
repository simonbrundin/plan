package executor

import (
	"context"
	"os"
	"path/filepath"
	"plan-agent/internal/runusage"
	"strings"
	"testing"
)

// TestExecuteApplicerarÄndring kontrollerar att filer skrivs och att diffen
// visar ändringen.
func TestExecuteApplicerarÄndring(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")

	exec, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = exec.Cleanup() }()

	result, err := exec.Execute(context.Background(), Request{
		Changes: []Change{{
			Path:    "main.go",
			Content: "package main // ändrad\n",
			Reason:  "test",
		}},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if len(result.ChangedFiles) == 0 {
		t.Error("inga ändrade filer rapporterades")
	}
	if !strings.Contains(result.Diff, "main.go") {
		t.Errorf("diffen nämner inte filen:\n%s", result.Diff)
	}

	// Originalet ska vara orört.
	original, err := os.ReadFile(filepath.Join(source, "main.go"))
	if err != nil {
		t.Fatalf("read original: %v", err)
	}
	if string(original) != "package main\n" {
		t.Errorf("originalet ändrades: %q", string(original))
	}
}

// TestExecuteKörTester kontrollerar att testkommandot körs och att resultatet
// speglar utfallet.
func TestExecuteKörTester(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "fil.txt"), "hej\n")

	exec, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = exec.Cleanup() }()

	result, err := exec.Execute(context.Background(), Request{
		Changes:     []Change{{Path: "fil.txt", Content: "hej igen\n"}},
		TestCommand: "cat fil.txt",
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if !result.TestsPassed {
		t.Errorf("testet ska ha gått igenom: %s", result.TestError)
	}
	if !strings.Contains(result.TestOutput, "hej igen") {
		t.Errorf("testutdata = %q", result.TestOutput)
	}
}

// TestExecuteRapportarMisslyckatTest kontrollerar att ett misslyckat test
// beskrivs i resultatet i stället för att kasta bort allt.
func TestExecuteRapportarMisslyckatTest(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "fil.txt"), "hej\n")

	exec, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = exec.Cleanup() }()

	result, err := exec.Execute(context.Background(), Request{
		Changes:     []Change{{Path: "fil.txt", Content: "hej\n"}},
		TestCommand: "cat fil-saknas.txt",
	})
	if err != nil {
		t.Fatalf("Execute ska inte kasta: %v", err)
	}

	if result.TestsPassed {
		t.Error("testet ska ha misslyckats")
	}
	if result.TestError == "" {
		t.Error("testfelet ska beskrivas")
	}
}

// TestExecuteUtanTestKommando kontrollerar att ändringar kan göras utan
// verifiering.
func TestExecuteUtanTestKommando(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")

	exec, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = exec.Cleanup() }()

	result, err := exec.Execute(context.Background(), Request{
		Changes: []Change{{Path: "main.go", Content: "package main // x\n"}},
	})
	if err != nil {
		t.Fatalf("Execute: %v", err)
	}

	if result.TestOutput != "" {
		t.Errorf("testutdata ska vara tom: %q", result.TestOutput)
	}
}

// TestExecuteBlockerarAvbrutenKörning kontrollerar att en avbruten körning
// inte applicerar ändringar.
func TestExecuteBlockerarAvbrutenKörning(t *testing.T) {
	source := t.TempDir()
	writeFile(t, filepath.Join(source, "main.go"), "package main\n")

	exec, err := New(source)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer func() { _ = exec.Cleanup() }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := exec.Execute(ctx, Request{
		Changes: []Change{{Path: "main.go", Content: "package main // x\n"}},
	}); err == nil {
		t.Error("en avbruten körning ska ge ett fel")
	}

	content, err := exec.Workspace().Read("main.go")
	if err != nil {
		t.Fatalf("Read: %v", err)
	}
	if content != "package main\n" {
		t.Errorf("ändringen applicerades trots avbruten körning: %q", content)
	}
}

// TestSummary kontrollerar att resultatet kan beskrivas i en händelse.
func TestSummary(t *testing.T) {
	result := &Result{
		ChangedFiles: []string{"main.go"},
		Diff:         "diff --git a/main.go b/main.go",
		TestError:    "fel",
	}

	summary := result.Summary()
	if summary["tests_passed"] != false {
		t.Error("tests_passed ska vara false")
	}
	if summary["test_error"] != "fel" {
		t.Errorf("test_error = %v", summary["test_error"])
	}
	if summary["changed_files"] == nil {
		t.Error("changed_files saknas")
	}
}

// TestDiffPreview kontrollerar att en lång diff kortas av.
func TestDiffPreview(t *testing.T) {
	long := strings.Repeat("rad\n", 100)

	preview := (&Result{Diff: long}).DiffPreview(10)
	if strings.Count(preview, "\n") > 10 {
		t.Errorf("förhandsvisningen har för många rader: %d", strings.Count(preview, "\n"))
	}
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

// TestVerktygsanropRegistreras kontrollerar att exekveraren rapporterar
// vilka verktyg som användes - det ska gå att se i kostnadsrapporten.
func TestVerktygsanropRegistreras(t *testing.T) {
	recorder := runusage.New("minimax", "MiniMax-M2.7")

	executor := &Executor{usage: recorder}
	executor.recordTool("write_file")
	executor.recordTool("run_test")
	executor.recordTool("write_file")

	snapshot := recorder.Snapshot()
	if snapshot.ToolCalls != 3 {
		t.Errorf("tool calls = %d, vill ha 3", snapshot.ToolCalls)
	}
	if len(snapshot.ToolNames) != 2 {
		t.Errorf("tool names = %v, vill ha två", snapshot.ToolNames)
	}
	if snapshot.ToolNames[0] != "run_test" || snapshot.ToolNames[1] != "write_file" {
		t.Errorf("tool names = %v, vill ha sorterade", snapshot.ToolNames)
	}
}
