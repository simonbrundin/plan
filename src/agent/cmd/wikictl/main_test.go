package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// runCLI utför wikictl med givna argument och fångar utdata.
func runCLI(t *testing.T, args ...string) string {
	t.Helper()

	var out bytes.Buffer
	if err := run(args, &out); err != nil {
		t.Fatalf("run(%v): %v", args, err)
	}
	return out.String()
}

// newTestWiki skapar en tom wiki att jobba mot.
func newTestWiki(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	for _, sub := range []string{"wiki/notes", "inbox"} {
		if err := os.MkdirAll(filepath.Join(dir, sub), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestNoteThenIngest(t *testing.T) {
	dir := newTestWiki(t)

	runCLI(t, "--wiki", dir, "note", "--title", "En viktig not", "--tags", "bug,session", "--body", "Texten i noten.")

	// Noten ska ligga i inboxen, inte bland sidorna.
	if _, err := os.Stat(filepath.Join(dir, "inbox", "en-viktig-not.md")); err != nil {
		t.Fatalf("noten skrevs inte till inboxen: %v", err)
	}

	runCLI(t, "--wiki", dir, "ingest")

	// Efter ingestion ska noten ligga bland sidorna och råmaterialet i archive/.
	if _, err := os.Stat(filepath.Join(dir, "wiki", "notes", "en-viktig-not.md")); err != nil {
		t.Errorf("sidan skapades inte: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "inbox", "en-viktig-not.md")); !os.IsNotExist(err) {
		t.Error("noten ligger kvar i inboxen")
	}
	if _, err := os.Stat(filepath.Join(dir, "archive")); err != nil {
		t.Errorf("arkivet saknas: %v", err)
	}

	// Genererade filer ska finnas - index.md byggs nu om vid varje ingest.
	for _, name := range []string{"index.md", "catalog.md", "graph.md", "log.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); err != nil {
			t.Errorf("%s skapades inte: %v", name, err)
		}
	}
}

func TestIngestRebuildsIndex(t *testing.T) {
	dir := newTestWiki(t)

	runCLI(t, "--wiki", dir, "note", "--title", "Not", "--body", "Text.")
	runCLI(t, "--wiki", dir, "ingest")

	content, err := os.ReadFile(filepath.Join(dir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "notes/not") {
		t.Errorf("index.md saknar den nya noten:\n%s", content)
	}
}

func TestSearchFindsNote(t *testing.T) {
	dir := newTestWiki(t)
	runCLI(t, "--wiki", dir, "note", "--title", "Sprint timer", "--tags", "bug", "--body", "timeLeft räknas aldrig ned.")
	runCLI(t, "--wiki", dir, "ingest")

	out := runCLI(t, "--wiki", dir, "search", "timeLeft")

	if !strings.Contains(out, "notes/sprint-timer") {
		t.Errorf("sökningen hittade inte noten:\n%s", out)
	}
}

func TestSearchResolvesWikilinks(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/concepts/Flux.md", "---\ntype: concept\n---\n\n# Flux\n\nSe [[concepts/Talos]].\n")
	writePage(t, dir, "wiki/concepts/Talos.md", "---\ntype: concept\n---\n\n# Talos\n\nInnehåll.\n")

	out := runCLI(t, "--wiki", dir, "search", "Talos")
	if !strings.Contains(out, "concepts/Talos") {
		t.Errorf("wikilänken hittades inte:\n%s", out)
	}

	graph := runCLI(t, "--wiki", dir, "graph", "--mermaid")
	if !strings.Contains(graph, "n_concepts_Flux --> n_concepts_Talos") {
		t.Errorf("grafen saknar kant mellan sidorna:\n%s", graph)
	}
}

func TestShowUnknownPageFails(t *testing.T) {
	dir := newTestWiki(t)

	var out bytes.Buffer
	if err := run([]string{"--wiki", dir, "show", "finns-inte"}, &out); err == nil {
		t.Fatal("expected error for unknown page")
	}
}

func TestIngestDryRunKeepsInbox(t *testing.T) {
	dir := newTestWiki(t)
	runCLI(t, "--wiki", dir, "note", "--title", "Kvar i inboxen", "--body", "Text.")

	runCLI(t, "--wiki", dir, "ingest", "--dry-run")

	if _, err := os.Stat(filepath.Join(dir, "inbox", "kvar-i-inboxen.md")); err != nil {
		t.Errorf("torrkörning rörde inboxen: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "wiki", "notes", "kvar-i-inboxen.md")); !os.IsNotExist(err) {
		t.Error("torrkörning skrev en sida")
	}
}

func TestInitAddsBlockToAgentsFile(t *testing.T) {
	repo := t.TempDir()

	runCLI(t, "init", repo)

	content, err := os.ReadFile(filepath.Join(repo, "AGENTS.md"))
	if err != nil {
		t.Fatalf("AGENTS.md skapades inte: %v", err)
	}
	if !strings.Contains(string(content), "wiki:start") {
		t.Errorf("AGENTS.md saknar wikinblocket:\n%s", content)
	}
}

// writePage hjälper testerna att skriva en sida till disk.
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

func TestPageCommandWritesPage(t *testing.T) {
	dir := newTestWiki(t)

	out := runCLI(t, "--wiki", dir, "page", "--type", "concept", "--title", "Kyverno CEL",
		"--summary", "CEL är standarden.", "--body", "# Kyverno CEL\n\nInnehåll.")

	if !strings.Contains(out, "concepts/kyverno-cel") {
		t.Errorf("utdata saknar sid-id:\n%s", out)
	}

	content, err := os.ReadFile(filepath.Join(dir, "wiki", "concepts", "kyverno-cel.md"))
	if err != nil {
		t.Fatalf("sidan skapades inte: %v", err)
	}
	if !strings.Contains(string(content), "summary: CEL är standarden.") {
		t.Errorf("sidan saknar summary:\n%s", content)
	}

	index, err := os.ReadFile(filepath.Join(dir, "index.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(index), "CEL är standarden.") {
		t.Errorf("index.md saknar sidan:\n%s", index)
	}
}

func TestLinkCommandAddsRelation(t *testing.T) {
	dir := newTestWiki(t)
	runCLI(t, "--wiki", dir, "page", "--type", "entity", "--title", "Talos", "--body", "Text.")
	runCLI(t, "--wiki", dir, "page", "--type", "concept", "--title", "GitOps", "--body", "Text.")

	runCLI(t, "--wiki", dir, "link", "--from", "entities/talos", "--to", "concepts/gitops")

	content, err := os.ReadFile(filepath.Join(dir, "wiki", "entities", "talos.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(content), "concepts/gitops") {
		t.Errorf("länken saknas:\n%s", content)
	}
}

func TestSourceAddAndList(t *testing.T) {
	dir := newTestWiki(t)
	src := filepath.Join(dir, "kalla.md")
	if err := os.WriteFile(src, []byte("Råtext.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runCLI(t, "--wiki", dir, "source", "add", "--file", src, "--title", "Min källa", "--origin", "https://example.com")

	out := runCLI(t, "--wiki", dir, "source", "list")
	if !strings.Contains(out, "min-kalla") {
		t.Errorf("listan saknar källan:\n%s", out)
	}
	if !strings.Contains(out, "ointegrerad") {
		t.Errorf("listan saknar status:\n%s", out)
	}
}

func TestLinkLogsOperation(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/entities/Talos.md", "---\ntype: entity\n---\n\n# Talos\n")
	writePage(t, dir, "wiki/concepts/GitOps.md", "---\ntype: concept\n---\n\n# GitOps\n")

	runCLI(t, "--wiki", dir, "link", "--from", "entities/talos", "--to", "concepts/gitops")

	log, err := os.ReadFile(filepath.Join(dir, "log.md"))
	if err != nil {
		t.Fatalf("loggen saknas: %v", err)
	}
	if !strings.Contains(string(log), "] link | entities/talos -> concepts/gitops") {
		t.Errorf("loggen saknar länkningen:\n%s", log)
	}
}

func TestSourceAddLogsOperation(t *testing.T) {
	dir := newTestWiki(t)
	src := filepath.Join(dir, "kalla.md")
	if err := os.WriteFile(src, []byte("Råtext.\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	runCLI(t, "--wiki", dir, "source", "add", "--file", src, "--title", "Min källa")

	log, err := os.ReadFile(filepath.Join(dir, "log.md"))
	if err != nil {
		t.Fatalf("loggen saknas: %v", err)
	}
	if !strings.Contains(string(log), "] source | ") {
		t.Errorf("loggen saknar råkällan:\n%s", log)
	}
}

func TestLintReportsBrokenLink(t *testing.T) {
	dir := newTestWiki(t)
	writePage(t, dir, "wiki/concepts/Flux.md", "---\ntype: concept\n---\n\n# Flux\n\nSe [[concepts/Saknas]].\n")

	out := runCLI(t, "--wiki", dir, "lint")
	if !strings.Contains(out, "trasig länk") {
		t.Errorf("lint hittade inte den trasiga länken:\n%s", out)
	}
}
