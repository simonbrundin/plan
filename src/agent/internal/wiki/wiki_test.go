package wiki

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// writePage skriver en sida till disk.
func writePage(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// newTestWiki skapar en wiki med en sida i varje typ.
func newTestWiki(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writePage(t, dir, "wiki/concepts/Flux.md", "---\ntype: concept\ntags: [flux]\n---\n\n# Flux\n\nEn operator.\n")
	return dir
}

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Arkitektur":              "arkitektur",
		"Sessionens livscykel":    "sessionens-livscykel",
		"Återkoppling av och-och": "aterkoppling-av-och-och",
		"  Bigram-algoritmen  ":   "bigram-algoritmen",
		"SQL & NoSQL":             "sql-nosql",
	}
	for input, want := range cases {
		if got := Slug(input); got != want {
			t.Errorf("Slug(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseNote(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/notes/arkitektur.md", `---
title: Arkitektur
type: note
tags: [arkitektur, go]
created: 2026-10-21
related: [bigram-algoritmen]
---

Paketen är delade i lager.
`)

	note, err := ParseNote(filepath.Join(dir, "wiki", "notes", "arkitektur.md"))
	if err != nil {
		t.Fatalf("ParseNote: %v", err)
	}
	if note.ID != "arkitektur" {
		t.Errorf("ID = %q, want arkitektur", note.ID)
	}
	if note.Title != "Arkitektur" {
		t.Errorf("Title = %q, want Arkitektur", note.Title)
	}
	if note.Type != "note" {
		t.Errorf("Type = %q, want note", note.Type)
	}
	if len(note.Tags) != 2 || note.Tags[0] != "arkitektur" {
		t.Errorf("Tags = %v, want [arkitektur go]", note.Tags)
	}
	if note.Created != "2026-10-21" {
		t.Errorf("Created = %q, want 2026-10-21", note.Created)
	}
	if len(note.Related) != 1 || note.Related[0] != "bigram-algoritmen" {
		t.Errorf("Related = %v, want [bigram-algoritmen]", note.Related)
	}
	if !strings.Contains(note.Body, "Paketen är delade i lager") {
		t.Errorf("Body = %q, saknar innehåll", note.Body)
	}
}

func TestParseNoteUsesHeadingAsTitle(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/entities/flux.md", "---\ntype: entity\n---\n\n# Flux GitOps\n\nEn operator.\n")

	note, err := ParseNote(filepath.Join(dir, "wiki", "entities", "flux.md"))
	if err != nil {
		t.Fatalf("ParseNote: %v", err)
	}
	if note.Title != "Flux GitOps" {
		t.Errorf("Title = %q, want Flux GitOps (rubrik används)", note.Title)
	}
}

func TestParseNoteWithoutFrontmatter(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/notes/utan.md", "Bara text.\n")

	note, err := ParseNote(filepath.Join(dir, "wiki", "notes", "utan.md"))
	if err != nil {
		t.Fatalf("ParseNote: %v", err)
	}
	if note.Title != "utan" {
		t.Errorf("Title = %q, want utan (fallback till id)", note.Title)
	}
	if note.Body != "Bara text." {
		t.Errorf("Body = %q, want 'Bara text.'", note.Body)
	}
}

func TestParseNoteBrokenFrontmatter(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/notes/trasig.md", "---\ntitle: Trasig\n")

	if _, err := ParseNote(filepath.Join(dir, "wiki", "notes", "trasig.md")); err == nil {
		t.Fatal("expected error for unterminated frontmatter")
	}
}

func TestBodyLinksHandlesBothStyles(t *testing.T) {
	note := &Note{
		ID:      "notes/a",
		Related: []string{"notes/b"},
		Body:    "Se [b](b.md) och [[concepts/C]] och [[notes/a]].",
	}

	links := note.BodyLinks()
	if len(links) != 3 {
		t.Fatalf("BodyLinks = %v, want 3", links)
	}

	all := note.AllRelated()
	if len(all) != 3 {
		t.Fatalf("AllRelated = %v, want 3 (a ska filtreras bort)", all)
	}
}

func TestOpenFindsPagesRecursively(t *testing.T) {
	dir := newTestWiki(t)

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(w.Notes) != 1 {
		t.Fatalf("Notes = %d, want 1", len(w.Notes))
	}
	if w.Notes[0].ID != "concepts/Flux" {
		t.Errorf("ID = %q, want concepts/Flux", w.Notes[0].ID)
	}
}

func TestResolveFindsWikilinkTarget(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/entities/Talos.md", "---\ntype: entity\n---\n\n# Talos\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got := w.Resolve("concepts/Flux"); got == nil {
		t.Error("Resolve(concepts/Flux) = nil, want noten")
	}
	if got := w.Resolve("entities/talos"); got == nil {
		t.Error("Resolve ska vara case-insensitive")
	}
	if got := w.Resolve("saknas"); got != nil {
		t.Error("Resolve(saknas) ska vara nil")
	}
}

func TestSearchRanksTitleAboveBody(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/concepts/Arkitektur.md", "---\ntitle: Arkitektur\ntags: [go]\n---\n\nPaketstrukturen.\n")
	writePage(t, dir, "wiki/concepts/annat.md", "---\ntitle: Annat\ntags: [noter]\n---\n\nIngen arkitektur här.\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	matches := w.Search("arkitektur")
	if len(matches) == 0 {
		t.Fatal("expected at least one match")
	}
	if matches[0].Note.ID != "concepts/Arkitektur" {
		t.Errorf("top match = %q, want concepts/Arkitektur", matches[0].Note.ID)
	}
}

func TestSearchWithTagFilter(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/notes/a.md", "---\ntitle: A\ntags: [go, test]\n---\n\nBigram\n")
	writePage(t, dir, "wiki/notes/b.md", "---\ntitle: B\ntags: [go]\n---\n\nBigram\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	matches := w.Search("bigram")
	if len(matches) != 2 {
		t.Fatalf("matches = %d, want 2", len(matches))
	}

	tagged := 0
	for _, m := range matches {
		if m.Note.HasTag("test") {
			tagged++
		}
	}
	if tagged != 1 {
		t.Errorf("HasTag(test) matched %d notes, want 1", tagged)
	}
}

func TestBacklinks(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/concepts/kallare.md", "---\ntitle: Källa\nrelated: [concepts/Flux]\n---\n\nSe [[concepts/Flux]].\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	back := w.Backlinks("concepts/Flux")
	if len(back) != 1 || back[0].ID != "concepts/kallare" {
		t.Errorf("Backlinks = %v, want [concepts/kallare]", back)
	}
}

func TestCatalogAndGraphAreGenerated(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/concepts/Talos.md", "---\ntype: concept\ntags: [x]\n---\n\n# Talos\n\nSe [[concepts/Flux]].\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}

	catalog := w.Catalog()
	if !strings.Contains(catalog, "concepts/Flux") {
		t.Errorf("katalogen saknar sidan:\n%s", catalog)
	}
	if !strings.Contains(catalog, "`x`") {
		t.Errorf("katalogen saknar taggsammanfattning:\n%s", catalog)
	}

	graph := w.Graph()
	if !strings.Contains(graph, "graph LR") {
		t.Errorf("graph saknar mermaid-header:\n%s", graph)
	}
	if !strings.Contains(graph, "n_concepts_Talos --> n_concepts_Flux") {
		t.Errorf("graph saknar kant mellan sidorna:\n%s", graph)
	}
	if !strings.Contains(graph, "Backlinks") {
		t.Errorf("graph saknar backlink-lista:\n%s", graph)
	}
}

func TestIngestCreatesPageAndArchivesInbox(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, InboxDir), 0o755); err != nil {
		t.Fatal(err)
	}
	inbox := filepath.Join(dir, InboxDir, "ny-not.md")
	content := "---\ntitle: Ny not\ntags: [nytt]\n---\n\nInnehåll i noten.\n"
	if err := os.WriteFile(inbox, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Ingest(dir, IngestOptions{Now: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if len(report.Created) != 1 || report.Created[0] != "notes/ny-not" {
		t.Errorf("Created = %v, want [notes/ny-not]", report.Created)
	}
	if len(report.Archived) != 1 {
		t.Errorf("Archived = %v, want 1", report.Archived)
	}

	pagePath := filepath.Join(dir, PagesDir, NotesDir, "ny-not.md")
	if _, err := os.Stat(pagePath); err != nil {
		t.Fatalf("sidan skapades inte: %v", err)
	}
	if _, err := os.Stat(inbox); !os.IsNotExist(err) {
		t.Error("inbox-filen ligger kvar efter ingest")
	}
	archived := filepath.Join(dir, ArchiveDir, "2026-10-21-ny-not.md")
	if _, err := os.Stat(archived); err != nil {
		t.Errorf("arkiv saknas: %v", err)
	}

	for _, generated := range []string{CatalogFile, GraphFile, LogFile} {
		if _, err := os.Stat(filepath.Join(dir, generated)); err != nil {
			t.Errorf("%s skapades inte: %v", generated, err)
		}
	}
}

func TestIngestAppendsToExistingPage(t *testing.T) {
	dir := t.TempDir()
	writePage(t, dir, "wiki/notes/befintlig.md", "---\ntitle: Befintlig\ntags: [a]\n---\n\nFörsta versionen.\n")

	if err := os.MkdirAll(filepath.Join(dir, InboxDir), 0o755); err != nil {
		t.Fatal(err)
	}
	inbox := filepath.Join(dir, InboxDir, "uppdatering.md")
	content := "---\ntitle: Befintlig\ntags: [b]\n---\n\nAndra versionen.\n"
	if err := os.WriteFile(inbox, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Ingest(dir, IngestOptions{Now: time.Date(2026, 10, 21, 12, 0, 0, 0, time.UTC)})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if len(report.Updated) != 1 || report.Updated[0] != "notes/befintlig" {
		t.Errorf("Updated = %v, want [notes/befintlig]", report.Updated)
	}
	if len(report.Created) != 0 {
		t.Errorf("Created = %v, want none", report.Created)
	}

	note, err := ParseNote(filepath.Join(dir, PagesDir, NotesDir, "befintlig.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(note.Body, "Första versionen.") || !strings.Contains(note.Body, "Andra versionen.") {
		t.Errorf("Body saknar en av versionerna:\n%s", note.Body)
	}
	if !strings.Contains(note.Body, "## Uppdatering 2026-10-21") {
		t.Errorf("Body saknar uppdateringsrubrik:\n%s", note.Body)
	}
	if len(note.Tags) != 2 {
		t.Errorf("Tags = %v, want 2 (a och b)", note.Tags)
	}
}

func TestIngestDryRunChangesNothing(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, InboxDir), 0o755); err != nil {
		t.Fatal(err)
	}
	inbox := filepath.Join(dir, InboxDir, "not.md")
	if err := os.WriteFile(inbox, []byte("---\ntitle: Not\n---\n\nText\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Ingest(dir, IngestOptions{DryRun: true})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if len(report.Created) != 1 {
		t.Errorf("Created = %v, want 1 (bara beräknat)", report.Created)
	}
	if _, err := os.Stat(filepath.Join(dir, PagesDir, NotesDir, "not.md")); !os.IsNotExist(err) {
		t.Error("torrkörning skrev en sida")
	}
	if _, err := os.Stat(inbox); err != nil {
		t.Error("torrkörning rörde inboxen")
	}
}

func TestIngestSkipsInboxReadme(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, InboxDir), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, InboxDir, "README.md"), []byte("# Inbox\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	report, err := Ingest(dir, IngestOptions{})
	if err != nil {
		t.Fatalf("Ingest: %v", err)
	}
	if len(report.Created) != 0 || len(report.Updated) != 0 {
		t.Errorf("README.md ingesterades: %+v", report)
	}
}

func TestSuggestMerges(t *testing.T) {
	dir := t.TempDir()
	// Samma ämne med olika skrivställe - ska föreslås.
	writePage(t, dir, "wiki/notes/arkitektur.md", "---\ntitle: Arkitektur\ntags: [go, session]\n---\n\nText\n")
	writePage(t, dir, "wiki/notes/arkitektur-2.md", "---\ntitle: arkitektur\ntags: [go]\n---\n\nText\n")
	// Olika ämnen som delar generiska taggar - ska inte föreslås.
	writePage(t, dir, "wiki/notes/algoritm.md", "---\ntitle: Algoritm\ntags: [go, session]\n---\n\nText\n")

	w, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	suggestions := SuggestMerges(w)
	if len(suggestions) != 1 {
		t.Fatalf("suggestions = %d, want 1 (bara liknande titlar)", len(suggestions))
	}
	if suggestions[0].A != "notes/arkitektur" || suggestions[0].B != "notes/arkitektur-2" {
		t.Errorf("suggestion = %+v, want notes/arkitektur + notes/arkitektur-2", suggestions[0])
	}
}

func TestResolveDirFindsWikiUpwards(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "wiki", PagesDir), 0o755); err != nil {
		t.Fatal(err)
	}
	nested := filepath.Join(root, "src", "go")
	if err := os.MkdirAll(nested, 0o755); err != nil {
		t.Fatal(err)
	}

	previous, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(nested); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(previous)
	})

	dir, err := ResolveDir("")
	if err != nil {
		t.Fatalf("ResolveDir: %v", err)
	}
	if dir != filepath.Join(root, "wiki") {
		t.Errorf("ResolveDir = %q, want %q", dir, filepath.Join(root, "wiki"))
	}
}
