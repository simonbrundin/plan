// Package executor gör kodändringar i en isolerad workspace och verifierar
// dem med projektets egna tester.
//
// Paketet är medvetet litet: agenten föreslår en ändring, workern utför den
// i en kopia av repot och rapporterar resultatet tillbaka till Plan. Originalet
// rörs aldrig - det är en människa som beslutar om ändringen ska appliceras.
package executor

import (
	"context"
	"fmt"
	"strings"

	"plan-agent/internal/workspace"
)

// Change är en föreslagen ändring i en fil.
type Change struct {
	// Path är sökvägen relativt repo-roten.
	Path string
	// Content är det nya innehållet.
	Content string
	// Reason förklarar varför ändringen görs.
	Reason string
}

// Request är en kodändring som ska utföras.
type Request struct {
	// Changes är de ändringar som ska göras.
	Changes []Change
	// TestCommand är kommandot som verifierar ändringen.
	TestCommand string
}

// Result är resultatet av en utförd ändring.
type Result struct {
	// ChangedFiles är de filer som ändrats.
	ChangedFiles []string
	// Diff är skillnaden mot originalet.
	Diff string
	// TestOutput är utdata från testkörningen.
	TestOutput string
	// TestsPassed berättar om testerna gick igenom.
	TestsPassed bool
	// TestError beskriver varför testerna inte gick igenom.
	TestError string
	// Branch är branchen ändringen ligger på i källrepoet.
	Branch string
}

// Executor utför kodändringar i en isolerad workspace.
type Executor struct {
	ws *workspace.Workspace
}

// New skapar en executor som arbetar i en kopia av source.
func New(source string) (*Executor, error) {
	ws, err := workspace.New(source)
	if err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}
	return &Executor{ws: ws}, nil
}

// Workspace ger tillgång till arbetsytan - till exempel för att läsa filer.
func (e *Executor) Workspace() *workspace.Workspace {
	return e.ws
}

// Execute applicerar ändringarna och kör verifieringen.
//
// Ändringarna appliceras först, sedan körs testerna. Ett fel i testerna är
// inte ett fel i utförandet - resultatet ska beskriva utfallet, inte
// kasta bort det.
//
// Testkommandot körs med skalning i arbetsytan. Det är säkert därför att
// arbetsytan är en isolerad kopia och körningen saknar hemligheter.
func (e *Executor) Execute(ctx context.Context, req Request) (*Result, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("execution cancelled: %w", err)
	}

	for _, change := range req.Changes {
		if err := e.ws.Write(change.Path, change.Content); err != nil {
			return nil, fmt.Errorf("failed to apply change to %s: %w", change.Path, err)
		}
	}

	changed, err := e.ws.ChangedFiles()
	if err != nil {
		return nil, fmt.Errorf("failed to list changed files: %w", err)
	}

	diff, err := e.ws.Diff()
	if err != nil {
		return nil, fmt.Errorf("failed to produce diff: %w", err)
	}

	result := &Result{
		ChangedFiles: changed,
		Diff:         diff,
	}

	if req.TestCommand == "" {
		return result, nil
	}

	output, testErr := e.ws.RunShell(req.TestCommand)
	result.TestOutput = output
	result.TestsPassed = testErr == nil
	if testErr != nil {
		result.TestError = testErr.Error()
	}

	return result, nil
}

// Commit lägger ändringen på en branch i källrepoet.
//
// Branchen gör ändringen granskbar: huvudgrenen rörs aldrig, och en människa
// väljer om den ska mergeas.
func (e *Executor) Commit(branch, message string) error {
	if err := e.ws.CommitToBranch(branch, message); err != nil {
		return fmt.Errorf("failed to commit the change: %w", err)
	}
	return nil
}

// Cleanup tar bort arbetsytan.
func (e *Executor) Cleanup() error {
	return e.ws.Cleanup()
}

// Summary beskriver resultatet på ett sätt som passar i en händelse.
func (r *Result) Summary() map[string]any {
	summary := map[string]any{
		"changed_files": r.ChangedFiles,
		"tests_passed":  r.TestsPassed,
	}
	if r.TestError != "" {
		summary["test_error"] = r.TestError
	}
	if r.Diff != "" {
		summary["diff_size"] = len(r.Diff)
	}
	return summary
}

// DiffPreview returnerar de första raderna av diffen - en hel diff kan vara
// för stor för ett händelse-payload.
func (r *Result) DiffPreview(maxLines int) string {
	if maxLines <= 0 {
		return ""
	}

	lines := strings.Split(r.Diff, "\n")
	if len(lines) > maxLines {
		lines = lines[:maxLines]
	}
	return strings.Join(lines, "\n")
}
