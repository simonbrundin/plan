package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// IndexFile är den innehållsorienterade katalogen över wikin.
// Den byggs om vid varje ingest och är det första en agent läser
// när den söker i wikin.
const IndexFile = "index.md"

// IndexByKind renderar index.md - en katalog över alla sidor,
// uppdelad efter sidtyp, med en rad per sida.
func (w *Wiki) IndexByKind() string {
	var b strings.Builder

	b.WriteString("<!-- GENERERAD av `wikictl index`. Ändra inte manuellt. -->\n")
	b.WriteString("# Wiki Index\n\n")
	b.WriteString("Katalog över allt i wikin. **Läs den här först** när du söker,\n")
	b.WriteString("och gå sedan ner på sidorna som är relevanta.\n\n")

	if len(w.Notes) == 0 {
		b.WriteString("_Inga sidor ännu._\n")
		return b.String()
	}

	// Räkna per typ.
	counts := map[PageKind]int{}
	for _, note := range w.Notes {
		counts[kindOf(note)]++
	}

	b.WriteString("| Typ | Antal |\n|---|---|\n")
	for _, kind := range Kinds {
		if counts[kind] == 0 {
			continue
		}
		fmt.Fprintf(&b, "| %s | %d |\n", kind.KindTitle(), counts[kind])
	}
	b.WriteString("\n")

	for _, kind := range Kinds {
		pages := w.pagesOfKind(kind)
		if len(pages) == 0 {
			continue
		}
		fmt.Fprintf(&b, "## %s\n\n", kind.KindTitle())
		for _, note := range pages {
			updated := note.Updated
			if updated == "" {
				updated = "-"
			}
			fmt.Fprintf(&b, "- [[%s]] - %s _(uppdaterad %s)_\n", note.ID, note.SummaryLine(), updated)
		}
		b.WriteString("\n")
	}

	return b.String()
}

// WriteIndex skriver index.md.
func (w *Wiki) WriteIndex() error {
	return writeFile(filepath.Join(w.Dir, IndexFile), w.IndexByKind())
}

// pagesOfKind returnerar sidorna av en typ, sorterade på titel.
func (w *Wiki) pagesOfKind(kind PageKind) []*Note {
	var out []*Note
	for _, note := range w.Notes {
		if kindOf(note) == kind {
			out = append(out, note)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title)
	})
	return out
}

// kindOf gissar sidtypen utifrån frontmatter, med katalogen som fallback.
func kindOf(note *Note) PageKind {
	switch PageKind(strings.ToLower(note.Type)) {
	case KindEntity, KindConcept, KindSummary, KindSynthesis, KindNote:
		return PageKind(strings.ToLower(note.Type))
	}
	dir := strings.SplitN(note.ID, "/", 2)[0]
	switch dir {
	case "entities":
		return KindEntity
	case "concepts":
		return KindConcept
	case "summaries":
		return KindSummary
	case "synthesis":
		return KindSynthesis
	}
	return KindNote
}

// SummaryLine ger en rad som beskriver en sida.
// Frontmatter-fältet `summary` används om det finns, annars härleds
// en kort beskrivning ur brödtextens första mening.
func (n *Note) SummaryLine() string {
	if s := strings.TrimSpace(n.Summary); s != "" {
		return s
	}
	return DeriveSummary(n.Body)
}

// DeriveSummary plockar ut första meningen i en text och kortar av den.
func DeriveSummary(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, "-") {
			continue
		}
		if strings.HasPrefix(line, "|") || strings.HasPrefix(line, "```") {
			continue
		}
		summary := line
		if idx := strings.Index(summary, ". "); idx > 0 {
			summary = summary[:idx+1]
		}
		summary = strings.TrimSpace(summary)
		if len(summary) > 140 {
			summary = summary[:137] + "..."
		}
		return summary
	}
	return ""
}

// SourceFile kopierar en fil in i raw/ som en oförändlig råkälla.
// Källan får en header med metadata så att den går att hitta igen.
func SourceFile(dir, srcPath, title, origin string, now time.Time) (string, error) {
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return "", fmt.Errorf("kunde inte läsa %s: %w", srcPath, err)
	}

	if strings.TrimSpace(title) == "" {
		title = strings.TrimSuffix(filepath.Base(srcPath), filepath.Ext(srcPath))
	}

	date := now.Format("2006-01-02")
	name := date + "-" + Slug(title) + ".md"
	dest := filepath.Join(dir, RawDir, name)

	if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
		return "", fmt.Errorf("kunde inte skapa %s: %w", filepath.Dir(dest), err)
	}
	if _, err := os.Stat(dest); err == nil {
		return "", fmt.Errorf("råkällan finns redan: %s", dest)
	}

	var b strings.Builder
	b.WriteString("---\n")
	fmt.Fprintf(&b, "title: %s\n", title)
	b.WriteString("type: source\n")
	fmt.Fprintf(&b, "added: %s\n", date)
	if origin != "" {
		fmt.Fprintf(&b, "origin: %s\n", origin)
	}
	b.WriteString("immutable: true\n")
	b.WriteString("---\n\n")
	b.Write(data)

	if err := os.WriteFile(dest, []byte(b.String()), 0o644); err != nil {
		return "", fmt.Errorf("kunde inte skriva %s: %w", dest, err)
	}

	// Om källan redan låg i raw/ tar vi bort originalet - annars
	// finns samma text två gånger i källsamlingen.
	if inRaw, err := isInRawDir(dir, srcPath); err == nil && inRaw {
		if err := os.Remove(srcPath); err != nil {
			return "", fmt.Errorf("kunde inte ta bort %s: %w", srcPath, err)
		}
	}

	return dest, nil
}

// isInRawDir svarar om sökvägen ligger i katalogens raw/-katalog.
func isInRawDir(dir, path string) (bool, error) {
	rawDir := filepath.Join(dir, RawDir)
	absRaw, err := filepath.Abs(rawDir)
	if err != nil {
		return false, err
	}
	absPath, err := filepath.Abs(path)
	if err != nil {
		return false, err
	}
	rel, err := filepath.Rel(absRaw, absPath)
	if err != nil {
		return false, err
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)), nil
}

// Source är en råkälla som lagts in i raw/.
type Source struct {
	ID      string // filnamnet utan .md
	Title   string
	Origin  string
	Added   string
	Path    string
	Indexed bool // har en sida i wikin som hänvisar till källan
}

// Sources listar råkällorna i raw/ och markerar de som redan
// integrerats i wikin (någon sida listar dem i `sources`).
func (w *Wiki) Sources() []Source {
	rawDir := filepath.Join(w.Dir, RawDir)
	entries, err := os.ReadDir(rawDir)
	if err != nil {
		return nil
	}

	// Samla ihop alla källor som nämns i sidornas `sources`.
	mentioned := map[string]bool{}
	for _, note := range w.Notes {
		for _, src := range note.Sources {
			mentioned[strings.ToLower(strings.TrimSuffix(filepath.Base(src), ".md"))] = true
			mentioned[strings.ToLower(strings.TrimSuffix(src, ".md"))] = true
		}
	}

	var out []Source
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".md") {
			continue
		}
		path := filepath.Join(rawDir, entry.Name())
		note, err := ParseNote(path)
		if err != nil {
			continue
		}
		id := strings.TrimSuffix(entry.Name(), ".md")
		out = append(out, Source{
			ID:      id,
			Title:   note.Title,
			Origin:  note.Origin,
			Added:   note.Added,
			Path:    path,
			Indexed: mentioned[strings.ToLower(id)],
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}
