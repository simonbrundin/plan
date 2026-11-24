package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestWritePageCreatesPage(t *testing.T) {
	dir := t.TempDir()

	id, err := WritePage(dir, WriteOptions{
		Kind:    KindConcept,
		Title:   "Kyverno CEL",
		Body:    "# Kyverno CEL\n\nCEL är standarden.",
		Summary: "CEL är standarden för policies.",
		Tags:    []string{"kyverno"},
	})
	if err != nil {
		t.Fatalf("WritePage: %v", err)
	}
	if id != "concepts/kyverno-cel" {
		t.Errorf("id = %q, want concepts/kyverno-cel", id)
	}

	note, err := ParseNote(filepath.Join(dir, PagesDir, "concepts", "kyverno-cel.md"))
	if err != nil {
		t.Fatal(err)
	}
	if note.Type != "concept" {
		t.Errorf("Type = %q, want concept", note.Type)
	}
	if note.Summary != "CEL är standarden för policies." {
		t.Errorf("Summary = %q", note.Summary)
	}
}

func TestWritePageAppendsWhenExisting(t *testing.T) {
	dir := t.TempDir()

	opts := WriteOptions{Kind: KindEntity, Title: "Talos", Body: "Första versionen."}
	if _, err := WritePage(dir, opts); err != nil {
		t.Fatal(err)
	}

	opts.Body = "Andra versionen."
	id, err := WritePage(dir, opts)
	if err != nil {
		t.Fatal(err)
	}
	if id != "entities/talos" {
		t.Errorf("id = %q", id)
	}

	note, err := ParseNote(filepath.Join(dir, PagesDir, "entities", "talos.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Body, "Första versionen.") || !strings.Contains(note.Body, "Andra versionen.") {
		t.Errorf("brödtext saknar en version:\n%s", note.Body)
	}
}

func TestWritePageReplace(t *testing.T) {
	dir := t.TempDir()
	opts := WriteOptions{Kind: KindEntity, Title: "Talos", Body: "Första versionen."}
	if _, err := WritePage(dir, opts); err != nil {
		t.Fatal(err)
	}

	opts.Body = "Bara den här."
	opts.Replace = true
	if _, err := WritePage(dir, opts); err != nil {
		t.Fatal(err)
	}

	note, err := ParseNote(filepath.Join(dir, PagesDir, "entities", "talos.md"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(note.Body, "Första versionen.") {
		t.Errorf("replace lämnade den gamla texten:\n%s", note.Body)
	}
}

func TestLinkAddsRelationBothWays(t *testing.T) {
	dir := t.TempDir()
	if _, err := WritePage(dir, WriteOptions{Kind: KindEntity, Title: "Talos", Body: "Text."}); err != nil {
		t.Fatal(err)
	}
	if _, err := WritePage(dir, WriteOptions{Kind: KindConcept, Title: "GitOps", Body: "Text."}); err != nil {
		t.Fatal(err)
	}

	if err := Link(dir, "entities/talos", "concepts/gitops"); err != nil {
		t.Fatalf("Link: %v", err)
	}

	note, err := ParseNote(filepath.Join(dir, PagesDir, "entities", "talos.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(strings.Join(note.Related, ","), "concepts/gitops") {
		t.Errorf("Related = %v, saknar concepts/gitops", note.Related)
	}
	if !strings.Contains(note.Body, "[[concepts/gitops]]") {
		t.Errorf("brödtexten saknar länken:\n%s", note.Body)
	}

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if back := w.Backlinks("concepts/gitops"); len(back) != 1 {
		t.Errorf("Backlinks = %v, want 1", back)
	}
}

func TestLinkToMissingPageFails(t *testing.T) {
	dir := t.TempDir()
	if _, err := WritePage(dir, WriteOptions{Kind: KindEntity, Title: "Talos", Body: "Text."}); err != nil {
		t.Fatal(err)
	}
	if err := Link(dir, "entities/talos", "concepts/finns-inte"); err == nil {
		t.Error("Link till saknad sida borde ge fel")
	}
}

func TestIndexListsPagesByKind(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/entities/Talos.md", "---\ntype: entity\nsummary: Operativsystemet.\n---\n\n# Talos\n")
	writePage(t, dir, "wiki/concepts/GitOps.md", "---\ntype: concept\n---\n\n# GitOps\n\nGit är sanningskällan. Mer text.\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	index := w.IndexByKind()
	if !strings.Contains(index, "## Entiteter") || !strings.Contains(index, "## Koncept") {
		t.Errorf("index saknar typavsnitt:\n%s", index)
	}
	if !strings.Contains(index, "Operativsystemet.") {
		t.Errorf("index saknar summary:\n%s", index)
	}
	if !strings.Contains(index, "Git är sanningskällan.") {
		t.Errorf("index saknar härledd summary:\n%s", index)
	}
}

func TestSourceFileAddsRawSource(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "kalla.md")
	if err := os.WriteFile(src, []byte("Råtext.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dest, err := SourceFile(dir, src, "Min källa", "https://example.com", time.Date(2026, 10, 23, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("SourceFile: %v", err)
	}
	if filepath.Base(dest) != "2026-10-23-min-kalla.md" {
		t.Errorf("filnamn = %q", filepath.Base(dest))
	}

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	sources := w.Sources()
	if len(sources) != 1 {
		t.Fatalf("Sources = %d, want 1", len(sources))
	}
	if sources[0].Indexed {
		t.Error("källan ska vara ointegrerad innan någon sida hänvisar till den")
	}
}

func TestLintFindsOrphanAndUnintegratedSource(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/concepts/Ensam.md", "---\ntype: concept\n---\n\n# Ensam\n\nText.\n")
	src := filepath.Join(dir, "kalla.md")
	if err := os.WriteFile(src, []byte("Råtext.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := SourceFile(dir, src, "Källa", "", time.Now()); err != nil {
		t.Fatal(err)
	}

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	findings := w.Lint(LintOptions{})

	var kinds []LintKind
	for _, f := range findings {
		kinds = append(kinds, f.Kind)
	}
	if !containsKind(kinds, LintOrphan) {
		t.Errorf("lint hittade inte ödelapp: %v", findings)
	}
	if !containsKind(kinds, LintUnintegrated) {
		t.Errorf("lint hittade inte ointegrerad källa: %v", findings)
	}
}

func containsKind(kinds []LintKind, want LintKind) bool {
	for _, k := range kinds {
		if k == want {
			return true
		}
	}
	return false
}

func TestAppendLogUsesGreppableFormat(t *testing.T) {
	dir := t.TempDir()
	entries := []LogEntry{{Date: "2026-10-23", Op: "ingest", Title: "create notes/foo"}}
	if err := AppendLog(dir, entries, []string{"[create] [[notes/foo]]"}); err != nil {
		t.Fatalf("AppendLog: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, LogFile))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "## [2026-10-23] ingest | create notes/foo") {
		t.Errorf("loggen har fel format:\n%s", data)
	}
}

func TestMergeSourcesAcceptsBothKeys(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "wiki", "summaries", "x.md")
	if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ntype: summary\nsource: a.md\nsources: [b.md, a.md]\n---\n\n# X\n"
	if err := os.WriteFile(page, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Notes) != 1 {
		t.Fatalf("väntade en sida, fick %d", len(w.Notes))
	}
	got := w.Notes[0].Sources
	want := []string{"b.md", "a.md"}
	if len(got) != len(want) {
		t.Fatalf("källor = %v, ville ha %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("källor = %v, ville ha %v", got, want)
		}
	}
}

func TestLintReportsDuplicateSection(t *testing.T) {
	dir := t.TempDir()
	page := filepath.Join(dir, "wiki", "concepts", "x.md")
	if err := os.MkdirAll(filepath.Dir(page), 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\ntype: concept\n---\n\n# X\n\n## Kopplingar\n\n- [[a]]\n\n## Kopplingar\n\n- [[b]]\n"
	if err := os.WriteFile(page, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	findings := w.Lint(LintOptions{})
	var found bool
	for _, f := range findings {
		if f.Kind == LintDuplicateSec {
			found = true
		}
	}
	if !found {
		t.Errorf("lint hittade inte den dubblerade sektionen: %v", findings)
	}
}

func TestSourceAddInsideRawMovesInsteadOfCopies(t *testing.T) {
	dir := t.TempDir()
	rawDir := filepath.Join(dir, "raw")
	if err := os.MkdirAll(rawDir, 0o755); err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(rawDir, "kalla.md")
	if err := os.WriteFile(src, []byte("Råtext.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	dest, err := SourceFile(dir, src, "Min källa", "https://example.com", time.Now())
	if err != nil {
		t.Fatalf("SourceFile: %v", err)
	}
	if _, err := os.Stat(src); !os.IsNotExist(err) {
		t.Error("originalen finns kvar i raw/")
	}
	data, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("kunde inte läsa dest: %v", err)
	}
	if !strings.Contains(string(data), "type: source") {
		t.Errorf("frontmatter saknas:\n%s", data)
	}
}
