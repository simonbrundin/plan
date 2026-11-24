package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// IngestOptions styr hur en ingest körs.
type IngestOptions struct {
	// DryRun beräknar allt men skriver ingenting till disk.
	DryRun bool
	// Now används för att göra körningen reproducerbar i tester.
	Now time.Time
}

// IngestReport sammanfattar vad en ingest gjorde.
type IngestReport struct {
	Created  []string // id:n för nya sidor
	Updated  []string // id:n för uppdaterade sidor
	Archived []string // arkiverade inbox-filer
	Suggest  []Merge  // förslag på sammanslagning
}

// Merge är ett förslag på att slå ihop två sidor.
type Merge struct {
	A    string
	B    string
	Tags []string
}

// String beskriver en merge på svenska.
func (m Merge) String() string {
	return fmt.Sprintf("%s + %s (delade taggar: %s)", m.A, m.B, strings.Join(m.Tags, ", "))
}

// Ingest tar hand om allt som ligger i inboxen:
// varje inlägg blir en sida (eller ett tillägg till en befintlig),
// råmaterialet arkiveras och katalog/graph/log byggs om.
func Ingest(dir string, opts IngestOptions) (*IngestReport, error) {
	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	date := now.Format("2006-01-02")

	items, err := readInbox(dir)
	if err != nil {
		return nil, err
	}

	report := &IngestReport{}
	w, err := Open(dir)
	if err != nil {
		return nil, err
	}

	archiveDir := filepath.Join(dir, ArchiveDir)
	if !opts.DryRun {
		if err := os.MkdirAll(archiveDir, 0o755); err != nil {
			return nil, fmt.Errorf("kunde inte skapa %s: %w", archiveDir, err)
		}
	}

	for _, item := range items {
		id := item.Slug()
		if id == "" {
			continue
		}
		pageID := NotesDir + "/" + id

		if existing := w.Note(pageID); existing != nil {
			existing.Body = appendUpdate(existing.Body, item.Body, date)
			existing.Tags = MergeTags(existing.Tags, item.Tags)
			existing.Updated = date
			if existing.Status == "" {
				existing.Status = "active"
			}
			existing.Related = MergeTags(existing.Related, item.AllRelated())

			report.Updated = append(report.Updated, pageID)
			if !opts.DryRun {
				if err := writeFile(existing.Path, existing.Render()); err != nil {
					return nil, err
				}
			}
		} else {
			note := &Note{
				ID:      pageID,
				Title:   item.Title,
				Type:    item.Type,
				Tags:    item.Tags,
				Created: item.Created,
				Updated: item.Updated,
				Related: item.AllRelated(),
				Status:  item.Status,
				Sources: item.Sources,
				Body:    item.Body,
				Path:    filepath.Join(dir, PagesDir, NotesDir, id+".md"),
			}
			if note.Created == "" {
				note.Created = date
			}
			if note.Updated == "" {
				note.Updated = date
			}
			if note.Status == "" {
				note.Status = "active"
			}
			if note.Type == "" {
				note.Type = "note"
			}

			report.Created = append(report.Created, pageID)
			if !opts.DryRun {
				if err := writeFile(note.Path, note.Render()); err != nil {
					return nil, err
				}
			}
			w.Notes = append(w.Notes, note)
		}

		archived := filepath.Join(archiveDir, fmt.Sprintf("%s-%s.md", date, id))
		report.Archived = append(report.Archived, filepath.Base(archived))
		if !opts.DryRun {
			if err := os.Rename(item.Path, archived); err != nil {
				return nil, fmt.Errorf("kunde inte arkivera %s: %w", item.Path, err)
			}
		}
	}

	w.sortNotes()
	report.Suggest = SuggestMerges(w)

	if !opts.DryRun {
		if err := w.WriteIndex(); err != nil {
			return nil, err
		}
		if err := w.WriteCatalog(); err != nil {
			return nil, err
		}
		if err := w.WriteGraph(); err != nil {
			return nil, err
		}
		if err := appendLog(dir, date, report); err != nil {
			return nil, err
		}
	}

	return report, nil
}

// readInbox loads the pending inbox items.
func readInbox(dir string) ([]*Note, error) {
	inboxDir := filepath.Join(dir, InboxDir)
	entries, err := os.ReadDir(inboxDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("kunde inte läsa %s: %w", inboxDir, err)
	}

	var items []*Note
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(name, ".md") {
			continue
		}
		if name == "README.md" || strings.HasPrefix(name, "_") {
			continue
		}

		item, err := ParseNote(filepath.Join(inboxDir, name))
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].Path < items[j].Path
	})
	return items, nil
}

// Slug returns the id an inbox item will get when ingested.
func (n *Note) Slug() string {
	if n.Title != "" {
		if slug := Slug(n.Title); slug != "" {
			return slug
		}
	}
	return Slug(n.ID)
}

// appendUpdate adds a dated section at the end of a page.
func appendUpdate(body, addition, date string) string {
	if strings.TrimSpace(addition) == "" {
		return body
	}
	if strings.TrimSpace(body) == "" {
		return strings.TrimSpace(addition)
	}
	return strings.TrimRight(body, "\n") + fmt.Sprintf("\n\n## Uppdatering %s\n\n%s\n", date, strings.TrimSpace(addition))
}

// SuggestMerges hittar par av sidor som sannolikt beskriver samma sak.
// Tre villkor måste vara uppfyllda:
//   - samma sidtyp (en concept och en summary är inte samma sak),
//   - nästan lika titel (slug lika, eller den ena titeln är ett prefix),
//   - minst en gemensam tagg.
//
// Ett brett tagg-delande räcker inte - generiska taggar som "kubernetes"
// gör att halva wikin matchar varandra.
func SuggestMerges(w *Wiki) []Merge {
	var out []Merge

	for i := 0; i < len(w.Notes); i++ {
		for j := i + 1; j < len(w.Notes); j++ {
			a, b := w.Notes[i], w.Notes[j]
			if a.Type != b.Type {
				continue
			}
			if !similarTitles(a.Title, b.Title) {
				continue
			}
			shared := sharedTags(a.Tags, b.Tags)
			if len(shared) == 0 {
				continue
			}
			out = append(out, Merge{A: a.ID, B: b.ID, Tags: shared})
		}
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].A != out[j].A {
			return out[i].A < out[j].A
		}
		return out[i].B < out[j].B
	})
	return out
}

// similarTitles avgör om två titlar troligen beskriver samma sak.
// Jämförelsen sker på slug utan bindestreck, så "FluxOperator" och
// "flux-operator" räknas som samma. Prefix räknas bara när det följs av
// en ny ord - "GitOps" är inte samma som "Flux GitOps".
func similarTitles(a, b string) bool {
	sa, sb := Slug(a), Slug(b)
	if stripSeparators(sa) == stripSeparators(sb) {
		return true
	}
	return isWordPrefix(sa, sb) || isWordPrefix(sb, sa)
}

// isWordPrefix avgör om a är ett prefix av b på ordgräns.
func isWordPrefix(a, b string) bool {
	if len(a) >= len(b) {
		return false
	}
	if !strings.HasPrefix(b, a) {
		return false
	}
	return b[len(a)] == '-'
}

// stripSeparators tar bort bindestreck och understreck ur en slug.
func stripSeparators(s string) string {
	return strings.NewReplacer("-", "", "_", "", " ", "").Replace(s)
}

func sharedTags(a, b []string) []string {
	inB := map[string]bool{}
	for _, tag := range b {
		inB[strings.ToLower(tag)] = true
	}

	var shared []string
	for _, tag := range a {
		if inB[strings.ToLower(tag)] {
			shared = append(shared, tag)
		}
	}
	return shared
}

// appendLog adds an entry to the wiki log, i wikins eget format:
// en grep-bart rad per händelse ("## [YYYY-MM-DD] ingest | Rubrik").
func appendLog(dir, date string, report *IngestReport) error {
	var entries []LogEntry
	var details []string

	for _, id := range report.Created {
		entries = append(entries, LogEntry{Date: date, Op: "ingest", Title: "create " + id})
		details = append(details, fmt.Sprintf("[create] [[%s]] — ny sida från inboxen", id))
	}
	for _, id := range report.Updated {
		entries = append(entries, LogEntry{Date: date, Op: "ingest", Title: "update " + id})
		details = append(details, fmt.Sprintf("[update] [[%s]] — tillägg från inboxen", id))
	}
	for _, name := range report.Archived {
		details = append(details, fmt.Sprintf("[archive] %s", name))
	}
	for _, s := range report.Suggest {
		details = append(details, fmt.Sprintf("[merge?] %s", s))
	}

	if len(entries) == 0 {
		return nil
	}
	return AppendLog(dir, entries, details)
}
