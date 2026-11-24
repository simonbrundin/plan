package wiki

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Directory names in the wiki.
const (
	PagesDir    = "wiki" // genererade sidor: entities/, concepts/, summaries/, notes/
	InboxDir    = "inbox"
	ArchiveDir  = "archive"
	SandboxDir  = "sandbox"
	RawDir      = "raw"        // immutabla råkällor
	CatalogFile = "catalog.md" // genererad katalog över alla sidor
	GraphFile   = "graph.md"   // genererad länkgraph
	LogFile     = "log.md"     // kronologisk logg
)

// NotesDir är där arbetsnoter (från alla repor) landar.
const NotesDir = "notes"

// Wiki är en laddad kunskapsgraf.
type Wiki struct {
	Dir   string
	Notes []*Note
}

// Open loads all pages from dir/wiki.
func Open(dir string) (*Wiki, error) {
	pagesRoot := filepath.Join(dir, PagesDir)

	w := &Wiki{Dir: dir}
	err := filepath.WalkDir(pagesRoot, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			if os.IsNotExist(err) {
				return fs.SkipAll
			}
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			return nil
		}

		note, err := ParseNote(path)
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(pagesRoot, path)
		if err != nil {
			return err
		}
		note.ID = strings.TrimSuffix(filepath.ToSlash(rel), ".md")
		w.Notes = append(w.Notes, note)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("kunde inte läsa %s: %w", pagesRoot, err)
	}

	w.sortNotes()
	return w, nil
}

func (w *Wiki) sortNotes() {
	sort.Slice(w.Notes, func(i, j int) bool {
		return w.Notes[i].ID < w.Notes[j].ID
	})
}

// Note returns the note with the given id, or nil.
func (w *Wiki) Note(id string) *Note {
	for _, n := range w.Notes {
		if strings.EqualFold(n.ID, id) {
			return n
		}
	}
	return nil
}

// Resolve finds the page a link target points at.
// Länkar kan skrivas som [[concepts/Flux]], [[Concepts/Flux]] eller
// som markdown-länkar, och matchningen är tolerant mot skiftläge,
// bindestreck och slug-form.
func (w *Wiki) Resolve(target string) *Note {
	target = strings.TrimSuffix(strings.TrimSpace(target), ".md")
	if target == "" {
		return nil
	}

	// 1. Exakt id, case-insensitive.
	for _, n := range w.Notes {
		if strings.EqualFold(n.ID, target) {
			return n
		}
	}

	// 2. Samma sista segment - "Summaries/Longhorn disk guide" ->
	//    summaries/longhorn-disk-guide.
	targetBase := path.Base(strings.ReplaceAll(target, "\\", "/"))
	for _, n := range w.Notes {
		if strings.EqualFold(path.Base(n.ID), targetBase) {
			return n
		}
	}

	// 3. Samma slug - "Longhorn disk guide" -> "longhorn-disk-guide".
	targetSlug := Slug(targetBase)
	for _, n := range w.Notes {
		if Slug(path.Base(n.ID)) == targetSlug {
			return n
		}
	}

	return nil
}

// Match is a search hit in a page.
type Match struct {
	Note  *Note
	Score int
	Lines []string
}

// Search finds pages matching the query, ranked by relevance.
// The title and tags weigh heavier than the body.
func (w *Wiki) Search(query string) []Match {
	terms := strings.Fields(strings.ToLower(query))
	if len(terms) == 0 {
		return nil
	}

	var matches []Match
	for _, note := range w.Notes {
		score := 0
		title := strings.ToLower(note.Title)
		tags := strings.ToLower(strings.Join(note.Tags, " "))
		body := strings.ToLower(note.Body)

		for _, term := range terms {
			if strings.Contains(title, term) {
				score += 10
			}
			if strings.Contains(tags, term) {
				score += 6
			}
			score += strings.Count(body, term)
		}

		if score == 0 {
			continue
		}
		matches = append(matches, Match{Note: note, Score: score, Lines: contextLines(note, terms)})
	}

	sort.SliceStable(matches, func(i, j int) bool {
		if matches[i].Score != matches[j].Score {
			return matches[i].Score > matches[j].Score
		}
		return matches[i].Note.ID < matches[j].Note.ID
	})

	return matches
}

// contextLines returns up to three body lines containing the search terms.
func contextLines(note *Note, terms []string) []string {
	const maxLines = 3

	var lines []string
	for i, line := range strings.Split(note.Body, "\n") {
		lower := strings.ToLower(line)
		for _, term := range terms {
			if strings.Contains(lower, term) {
				lines = append(lines, fmt.Sprintf("%4d| %s", i+1, strings.TrimSpace(line)))
				break
			}
		}
		if len(lines) == maxLines {
			break
		}
	}
	return lines
}

// Backlinks returns the pages linking to the given id.
func (w *Wiki) Backlinks(id string) []*Note {
	var out []*Note
	for _, note := range w.Notes {
		for _, link := range note.Links(w) {
			if strings.EqualFold(link.ID, id) {
				out = append(out, note)
				break
			}
		}
	}
	return out
}

// Tags returns all tags in the wiki with their frequency.
func (w *Wiki) Tags() map[string]int {
	counts := map[string]int{}
	for _, note := range w.Notes {
		for _, tag := range note.Tags {
			counts[tag]++
		}
	}
	return counts
}

// Catalog renders the generated catalog.md - en fullständig lista över alla
// sidor. Den kuraterade index.md lämnas orörd.
func (w *Wiki) Catalog() string {
	var b strings.Builder

	b.WriteString("<!-- GENERERAD av `wikictl index`. Ändra inte manuellt. -->\n")
	b.WriteString("# Katalog\n\n")
	b.WriteString("Alla sidor i wikin, sorterade på id. Den kuraterade katalogen\n")
	b.WriteString("finns i `index.md`.\n\n")

	if len(w.Notes) == 0 {
		b.WriteString("_Inga sidor ännu._\n")
		return b.String()
	}

	b.WriteString("| Sida | Titel | Typ | Taggar | Uppdaterad | Länkar |\n")
	b.WriteString("|---|---|---|---|---|---|\n")
	for _, note := range w.Notes {
		title := strings.ReplaceAll(note.Title, "|", "\\|")
		tags := strings.Join(note.Tags, ", ")
		if tags == "" {
			tags = "-"
		}
		noteType := note.Type
		if noteType == "" {
			noteType = "-"
		}
		updated := note.Updated
		if updated == "" {
			updated = "-"
		}
		fmt.Fprintf(&b, "| [[%s]] | %s | %s | %s | %s | %d |\n",
			note.ID, title, noteType, tags, updated, len(note.Links(w)))
	}

	b.WriteString("\n## Taggar\n\n")
	tags := w.Tags()
	names := make([]string, 0, len(tags))
	for tag := range tags {
		names = append(names, tag)
	}
	sort.Strings(names)
	for _, tag := range names {
		fmt.Fprintf(&b, "- `%s` (%d)\n", tag, tags[tag])
	}

	return b.String()
}

// Graph renders the generated graph.md with a mermaid diagram and backlinks.
func (w *Wiki) Graph() string {
	var b strings.Builder

	b.WriteString("<!-- GENERERAD av `wikictl graph`. Ändra inte manuellt. -->\n")
	b.WriteString("# Länkgraph\n\n")
	b.WriteString("Kunskapsgrafen: varje nod är en sida, varje kant är en länk mellan dem.\n\n")

	b.WriteString("```mermaid\n")
	b.WriteString("graph LR\n")
	for _, note := range w.Notes {
		fmt.Fprintf(&b, "  %s[\"%s\"]\n", nodeID(note.ID), escapeLabel(note.Title))
	}
	for _, note := range w.Notes {
		for _, link := range note.Links(w) {
			fmt.Fprintf(&b, "  %s --> %s\n", nodeID(note.ID), nodeID(link.ID))
		}
	}
	b.WriteString("```\n\n")

	b.WriteString("## Backlinks\n\n")
	for _, note := range w.Notes {
		back := w.Backlinks(note.ID)
		if len(back) == 0 {
			continue
		}
		var names []string
		for _, b := range back {
			names = append(names, fmt.Sprintf("[[%s]]", b.ID))
		}
		fmt.Fprintf(&b, "- **%s** <- %s\n", note.ID, strings.Join(names, ", "))
	}

	return b.String()
}

// nodeID makes an id safe to use as a mermaid node identifier.
func nodeID(id string) string {
	replacer := strings.NewReplacer("/", "_", " ", "_", "-", "_", ".", "_")
	return "n_" + replacer.Replace(id)
}

// escapeLabel escapes quotes in mermaid labels.
func escapeLabel(label string) string {
	return strings.ReplaceAll(label, "\"", "'")
}

// WriteCatalog writes the generated catalog.md.
func (w *Wiki) WriteCatalog() error {
	return writeFile(filepath.Join(w.Dir, CatalogFile), w.Catalog())
}

// WriteGraph writes the generated graph.md.
func (w *Wiki) WriteGraph() error {
	return writeFile(filepath.Join(w.Dir, GraphFile), w.Graph())
}

// writeFile writes content, creating parent directories as needed.
func writeFile(path, content string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("kunde inte skapa %s: %w", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("kunde inte skriva %s: %w", path, err)
	}
	return nil
}

// ResolveDir finds the wiki directory.
// Priority: explicit path, then $WIKI_DIR, then a wiki/ directory
// found by walking upwards from the working directory, and finally the
// global ~/wiki.
func ResolveDir(explicit string) (string, error) {
	if explicit != "" {
		return explicit, nil
	}

	if env := os.Getenv("WIKI_DIR"); env != "" {
		return env, nil
	}

	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("kunde inte läsa arbetskatalogen: %w", err)
	}

	for {
		candidate := filepath.Join(dir, "wiki")
		if info, err := os.Stat(filepath.Join(candidate, PagesDir)); err == nil && info.IsDir() {
			return candidate, nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	if home, err := os.UserHomeDir(); err == nil {
		global := filepath.Join(home, "wiki")
		if info, err := os.Stat(filepath.Join(global, PagesDir)); err == nil && info.IsDir() {
			return global, nil
		}
	}

	return "", fmt.Errorf("hittade inte wikin - kör från repot eller sätt WIKI_DIR")
}

// markdownLinkPattern matchar [text](fil.md).
var markdownLinkPattern = regexp.MustCompile(`\]\(([^)]+\.md)\)`)

// wikilinkPattern matchar [[Sida/Med Namn]].
var wikilinkPattern = regexp.MustCompile(`\[\[([^\]|]+)(?:\|[^\]]*)?\]\]`)

// codeSpanPattern matchar inline-kod (`...`).
var codeSpanPattern = regexp.MustCompile("`[^`]*`")
