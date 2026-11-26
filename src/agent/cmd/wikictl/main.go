// Command wikictl hanterar wikin: sök, lista, skriv inlägg och
// ingestera inboxen in i kunskapsgrafen.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"plan/agent/internal/wiki"
)

func main() {
	if err := run(os.Args[1:], os.Stdout); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, "fel:", err)
		os.Exit(1)
	}
}

// run dispatchar till rätt delkommando.
func run(args []string, out io.Writer) error {
	wikiDir, args := splitWikiFlag(args)
	if len(args) == 0 {
		usage(out)
		return nil
	}

	cmd, rest := args[0], args[1:]
	switch cmd {
	case "search":
		return runSearch(rest, out, wikiDir)
	case "list":
		return runList(rest, out, wikiDir)
	case "show":
		return runShow(rest, out, wikiDir)
	case "graph":
		return runGraph(rest, out, wikiDir)
	case "inbox":
		return runInbox(rest, out, wikiDir)
	case "note":
		return runNote(rest, out, wikiDir)
	case "page":
		return runPage(rest, out, wikiDir)
	case "link":
		return runLink(rest, out, wikiDir)
	case "source":
		return runSource(rest, out, wikiDir)
	case "ingest":
		return runIngest(rest, out, wikiDir)
	case "index":
		return runIndex(rest, out, wikiDir)
	case "lint":
		return runLint(rest, out, wikiDir)
	case "init":
		return runInit(rest, out)
	case "help", "-h", "--help":
		usage(out)
		return nil
	default:
		return fmt.Errorf("okänt kommando: %q (kör 'wikictl help')", cmd)
	}
}

func usage(out io.Writer) {
	_, _ = fmt.Fprint(out, `wikictl - hantera kunskaps-wikin

Användning:
  wikictl <kommando> [flaggor]

Kommandon:
  search <fråga>      Sök i wikin (--tag, --project, --limit)
  list                Lista alla sidor (--project)
  show <id>           Visa en sida
  graph               Visa kunskapsgrafen (--mermaid)
  inbox               Visa inlägg som väntar på att ingesteras
  note                Skriv en not till inboxen (--title, --tags, --body, --project)
  page                Skriv en sida (--type, --title, --body, --file, --tags, ...)
  link                Länka en sida till en annan (--from, --to, --remove)
  source              Hantera råkällor (add, list)
  ingest              Ingestera inboxen (--dry-run)
  index               Bygg om index.md, catalog.md och graph.md
  lint                Kolla wikin (trasiga länkar, ödelappar, dubbler, gammal info)
  init [repo]         Lägg in wikinreglerna i ett repos AGENTS.md

Globala flaggor:
  --wiki <sökväg>     Wiki-katalogen (default: hittas uppifrån, eller $WIKI_DIR)
`)
}

// splitWikiFlag plockar ut --wiki ur argumentlistan.
func splitWikiFlag(args []string) (string, []string) {
	var wikiDir string
	rest := make([]string, 0, len(args))

	for i := 0; i < len(args); i++ {
		if args[i] == "--wiki" && i+1 < len(args) {
			wikiDir = args[i+1]
			i++
			continue
		}
		if strings.HasPrefix(args[i], "--wiki=") {
			wikiDir = strings.TrimPrefix(args[i], "--wiki=")
			continue
		}
		rest = append(rest, args[i])
	}

	return wikiDir, rest
}

// newFlagSet bygger en flaggsats för ett delkommando.
func newFlagSet(name string, args []string) (*flag.FlagSet, []string) {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	return fs, args
}

func openWiki(wikiDir string) (*wiki.Wiki, error) {
	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return nil, err
	}
	return wiki.Open(dir)
}

func runSearch(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("search", args)
	tag := fs.String("tag", "", "begränsa sökningen till en tagg")
	project := fs.String("project", "", "begränsa sökningen till ett projekt")
	limit := fs.Int("limit", 10, "max antal träffar")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("sökning kräver en fråga")
	}

	w, err := openWiki(wikiDir)
	if err != nil {
		return err
	}

	matches := w.Search(strings.Join(fs.Args(), " "))
	shown := 0
	for _, m := range matches {
		if *tag != "" && !m.Note.HasTag(*tag) {
			continue
		}
		if *project != "" && !m.Note.HasProject(*project) {
			continue
		}
		if *limit > 0 && shown >= *limit {
			break
		}
		shown++

		_, _ = fmt.Fprintf(out, "%s  (poäng %d)\n", m.Note.ID, m.Score)
		if len(m.Note.Tags) > 0 {
			_, _ = fmt.Fprintf(out, "  taggar: %s\n", strings.Join(m.Note.Tags, ", "))
		}
		for _, line := range m.Lines {
			_, _ = fmt.Fprintf(out, "  %s\n", line)
		}
		_, _ = fmt.Fprintln(out)
	}

	if shown == 0 {
		_, _ = fmt.Fprintln(out, "Inga träffar. Skriv en not i inboxen om detta borde finnas i grafen.")
	}
	return nil
}

func runList(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("list", args)
	project := fs.String("project", "", "visa bara sidor för ett projekt")
	if err := fs.Parse(rest); err != nil {
		return err
	}

	w, err := openWiki(wikiDir)
	if err != nil {
		return err
	}
	if len(w.Notes) == 0 {
		_, _ = fmt.Fprintln(out, "Wikin är tom.")
		return nil
	}

	for _, note := range w.Notes {
		if *project != "" && !note.HasProject(*project) {
			continue
		}
		_, _ = fmt.Fprintf(out, "%-40s %-35s %s\n", note.ID, note.Title, strings.Join(note.Tags, ", "))
	}
	return nil
}

func runShow(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("show", args)
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if fs.NArg() == 0 {
		return fmt.Errorf("show kräver en id")
	}

	w, err := openWiki(wikiDir)
	if err != nil {
		return err
	}
	note := w.Note(fs.Arg(0))
	if note == nil {
		return fmt.Errorf("sidan %q finns inte", fs.Arg(0))
	}
	_, _ = fmt.Fprint(out, note.Render())
	return nil
}

func runGraph(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("graph", args)
	mermaidOnly := fs.Bool("mermaid", false, "skriv endast ut mermaid-diagrammet")
	if err := fs.Parse(rest); err != nil {
		return err
	}

	w, err := openWiki(wikiDir)
	if err != nil {
		return err
	}
	if *mermaidOnly {
		_, _ = fmt.Fprint(out, mermaidBlock(w.Graph()))
		return nil
	}
	_, _ = fmt.Fprint(out, w.Graph())
	return nil
}

func runInbox(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("inbox", args)
	if err := fs.Parse(rest); err != nil {
		return err
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}
	entries, err := os.ReadDir(filepath.Join(dir, wiki.InboxDir))
	if err != nil {
		if os.IsNotExist(err) {
			_, _ = fmt.Fprintln(out, "Inboxen är tom.")
			return nil
		}
		return err
	}

	found := false
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		if name == "README.md" || strings.HasPrefix(name, "_") {
			continue
		}
		found = true
		_, _ = fmt.Fprintln(out, name)
	}
	if !found {
		_, _ = fmt.Fprintln(out, "Inboxen är tom.")
	}
	return nil
}

func runNote(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("note", args)
	title := fs.String("title", "", "rubrik på noten")
	tags := fs.String("tags", "", "taggar, kommaseparerade")
	body := fs.String("body", "", "texten i noten")
	project := fs.String("project", "", "projektet noten gäller (standard: repoet du kör i)")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *title == "" {
		return fmt.Errorf("--title krävs")
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}

	inbox := filepath.Join(dir, wiki.InboxDir)
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		return err
	}

	note := &wiki.Note{
		ID:    wiki.Slug(*title),
		Title: *title,
		Type:  "note",
		Tags:  splitTags(*tags),
		Body:  strings.TrimSpace(*body),
	}

	// Tagga med projektet om det inte redan finns en sådan tagg.
	if *project == "" {
		*project = wiki.DetectProject("")
	}
	if *project != "" && note.ProjectOf() == "" {
		note.Tags = append(note.Tags, wiki.ProjectTag(*project))
	}

	path := filepath.Join(inbox, note.ID+".md")
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("det finns redan en not med samma rubrik i inboxen: %s", path)
	}

	if err := os.WriteFile(path, []byte(note.Render()), 0o644); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "Skrev %s\n", path)
	_, _ = fmt.Fprintln(out, "Kör 'wikictl ingest' för att lägga noten i grafen.")
	return nil
}

func splitTags(raw string) []string {
	if raw == "" {
		return nil
	}
	parts := strings.Split(raw, ",")
	tags := make([]string, 0, len(parts))
	for _, p := range parts {
		if p = strings.TrimSpace(p); p != "" {
			tags = append(tags, p)
		}
	}
	return tags
}

// runPage skriver en sida i wiki-lagret. Det är agentens sätt att
// underhålla entity-, concept-, summary- och synthesis-sidor - de
// filerna ägs av LLM:en, inte av inboxen.
func runPage(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("page", args)
	kind := fs.String("type", "note", "sidtyp: note, entity, concept, summary, synthesis")
	title := fs.String("title", "", "rubrik på sidan")
	body := fs.String("body", "", "brödtext")
	file := fs.String("file", "", "läs brödtexten från en fil")
	summary := fs.String("summary", "", "enradig sammanfattning (visas i indexet)")
	tags := fs.String("tags", "", "taggar, kommaseparerade")
	related := fs.String("related", "", "relaterade sidor, kommaseparerade")
	sources := fs.String("sources", "", "råkällor, kommaseparerade")
	replace := fs.Bool("replace", false, "ersätt brödtexten i stället för att lägga till")
	project := fs.String("project", "", "projektet sidan gäller (standard: repoet du kör i)")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *title == "" {
		return fmt.Errorf("--title krävs")
	}

	pageKind, err := wiki.ParseKind(*kind)
	if err != nil {
		return err
	}

	text := *body
	if *file != "" {
		data, err := os.ReadFile(*file)
		if err != nil {
			return fmt.Errorf("kunde inte läsa %s: %w", *file, err)
		}
		text = string(data)
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("--body eller --file krävs")
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}

	pageTags := splitTags(*tags)
	if *project == "" {
		*project = wiki.DetectProject("")
	}
	if *project != "" && !hasProjectTag(pageTags, *project) {
		pageTags = append(pageTags, wiki.ProjectTag(*project))
	}

	id, err := wiki.WritePage(dir, wiki.WriteOptions{
		Kind:    pageKind,
		Title:   *title,
		Body:    text,
		Summary: *summary,
		Tags:    pageTags,
		Related: splitTags(*related),
		Sources: splitTags(*sources),
		Replace: *replace,
	})
	if err != nil {
		return err
	}

	if err := refreshIndex(dir); err != nil {
		return err
	}
	if err := wiki.AppendLog(dir, []wiki.LogEntry{{
		Date:  time.Now().Format("2006-01-02"),
		Op:    "page",
		Title: id,
	}}, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "Skrev [[%s]]\n", id)
	return nil
}

// hasProjectTag svarar om listan redan innehåller projekttaggen.
func hasProjectTag(tags []string, project string) bool {
	want := wiki.ProjectTag(project)
	for _, tag := range tags {
		if strings.EqualFold(tag, want) {
			return true
		}
	}
	return false
}

// runLink lägger till eller tar bort en länk mellan två sidor.
func runLink(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("link", args)
	from := fs.String("from", "", "sidan som ska länka")
	to := fs.String("to", "", "sidan som ska länkas till")
	remove := fs.Bool("remove", false, "ta bort länken i stället för att lägga till")
	if err := fs.Parse(rest); err != nil {
		return err
	}
	if *from == "" || *to == "" {
		return fmt.Errorf("--from och --to krävs")
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}

	if *remove {
		if err := wiki.Unlink(dir, *from, *to); err != nil {
			return err
		}
	} else if err := wiki.Link(dir, *from, *to); err != nil {
		return err
	}

	if err := refreshIndex(dir); err != nil {
		return err
	}

	op := "link"
	verb := "Länkade"
	if *remove {
		op = "unlink"
		verb = "Tog bort länk"
	}

	if err := wiki.AppendLog(dir, []wiki.LogEntry{{
		Date:  time.Now().Format("2006-01-02"),
		Op:    op,
		Title: *from + " -> " + *to,
	}}, nil); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "%s %s -> %s\n", verb, *from, *to)
	return nil
}

// runSource hanterar råkällager: add lägger in en fil, list visar vad
// som finns och vad som redan integrerats.
func runSource(args []string, out io.Writer, wikiDir string) error {
	if len(args) == 0 {
		return fmt.Errorf("source kräver add eller list")
	}

	sub := args[0]
	rest := args[1:]

	fs, rest := newFlagSet("source "+sub, rest)
	path := fs.String("file", "", "filen som ska läggas in som råkälla")
	title := fs.String("title", "", "rubrik på källan")
	origin := fs.String("origin", "", "varifrån källan kommer (url, sökväg)")
	if err := fs.Parse(rest); err != nil {
		return err
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}

	switch sub {
	case "add":
		if *path == "" && fs.NArg() > 0 {
			*path = fs.Arg(0)
		}
		if *path == "" {
			return fmt.Errorf("--file krävs")
		}
		dest, err := wiki.SourceFile(dir, *path, *title, *origin, time.Now())
		if err != nil {
			return err
		}
		if err := wiki.AppendLog(dir, []wiki.LogEntry{{
			Date:  time.Now().Format("2006-01-02"),
			Op:    "source",
			Title: filepath.Base(dest),
		}}, nil); err != nil {
			return err
		}
		_, _ = fmt.Fprintf(out, "Lade in råkälla %s\n", dest)
		_, _ = fmt.Fprintln(out, "Skriv nu en sammanfattning i wikin och länka den till källan.")

	case "list":
		w, err := wiki.Open(dir)
		if err != nil {
			return err
		}
		sources := w.Sources()
		if len(sources) == 0 {
			_, _ = fmt.Fprintln(out, "Inga råkällor.")
			return nil
		}
		for _, source := range sources {
			state := "ointegrerad"
			if source.Indexed {
				state = "integrerad"
			}
			origin := source.Origin
			if origin == "" {
				origin = "-"
			}
			_, _ = fmt.Fprintf(out, "%-40s %-12s %-14s %s\n", source.ID, state, source.Added, origin)
		}

	default:
		return fmt.Errorf("okänd source-kommando: %q (giltiga: add, list)", sub)
	}

	return nil
}

// refreshIndex bygger om index.md, catalog.md och graph.md.
func refreshIndex(dir string) error {
	w, err := wiki.Open(dir)
	if err != nil {
		return err
	}
	if err := w.WriteIndex(); err != nil {
		return err
	}
	if err := w.WriteCatalog(); err != nil {
		return err
	}
	return w.WriteGraph()
}

func runIngest(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("ingest", args)
	dryRun := fs.Bool("dry-run", false, "beräkna men skriv inget")
	if err := fs.Parse(rest); err != nil {
		return err
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}

	report, err := wiki.Ingest(dir, wiki.IngestOptions{DryRun: *dryRun})
	if err != nil {
		return err
	}

	if *dryRun {
		_, _ = fmt.Fprintln(out, "Torrkörning - inget skrevs.")
	}

	_, _ = fmt.Fprintf(out, "Skapade:     %s\n", joinOrNone(report.Created))
	_, _ = fmt.Fprintf(out, "Uppdaterade: %s\n", joinOrNone(report.Updated))
	_, _ = fmt.Fprintf(out, "Arkiverade:  %s\n", joinOrNone(report.Archived))

	if len(report.Suggest) > 0 {
		_, _ = fmt.Fprintln(out, "\nFörslag att slå ihop (kräver agent):")
		for _, s := range report.Suggest {
			_, _ = fmt.Fprintf(out, "  %s\n", s)
		}
	}
	return nil
}

func runIndex(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("index", args)
	if err := fs.Parse(rest); err != nil {
		return err
	}

	dir, err := wiki.ResolveDir(wikiDir)
	if err != nil {
		return err
	}
	if err := refreshIndex(dir); err != nil {
		return err
	}
	_, _ = fmt.Fprintln(out, "index.md, catalog.md och graph.md är uppdaterade.")
	return nil
}

func runInit(args []string, out io.Writer) error {
	fs, rest := newFlagSet("init", args)
	if err := fs.Parse(rest); err != nil {
		return err
	}

	repoDir := fs.Arg(0)
	if repoDir == "" {
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		repoDir = cwd
	}

	changed, err := wiki.EnsureAgentsBlock(repoDir)
	if err != nil {
		return err
	}

	if changed {
		_, _ = fmt.Fprintf(out, "Uppdaterade %s\n", filepath.Join(repoDir, wiki.AgentsFileName))
	} else {
		_, _ = fmt.Fprintf(out, "%s var redan uppdaterad\n", filepath.Join(repoDir, wiki.AgentsFileName))
	}
	return nil
}

func runLint(args []string, out io.Writer, wikiDir string) error {
	fs, rest := newFlagSet("lint", args)
	staleDays := fs.Int("stale-days", 0, "flagga sidor som inte uppdaterats på så många dagar")
	if err := fs.Parse(rest); err != nil {
		return err
	}

	w, err := openWiki(wikiDir)
	if err != nil {
		return err
	}

	opts := wiki.LintOptions{}
	if *staleDays > 0 {
		opts.StaleAfter = time.Duration(*staleDays) * 24 * time.Hour
	}

	findings := w.Lint(opts)
	for _, finding := range findings {
		_, _ = fmt.Fprintln(out, finding.String())
	}

	if len(findings) == 0 {
		_, _ = fmt.Fprintln(out, "Inga problem hittade.")
	} else {
		_, _ = fmt.Fprintf(out, "\n%d problem hittade.\n", len(findings))
	}
	return nil
}

// mermaidBlock extraherar mermaid-diagrammet ur den genererade graph.md.
func mermaidBlock(graph string) string {
	const fence = "```"

	start := strings.Index(graph, fence+"mermaid")
	if start < 0 {
		return ""
	}
	body := start + len(fence) + len("mermaid")
	end := strings.Index(graph[body:], fence)
	if end < 0 {
		return ""
	}
	return strings.TrimSpace(graph[body:body+end]) + "\n"
}

func joinOrNone(items []string) string {
	if len(items) == 0 {
		return "-"
	}
	return strings.Join(items, ", ")
}
