package wiki

import (
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"strings"
	"time"
)

// PageKind är en sidtyp i wiki-lagret. Varje typ har en egen katalog.
type PageKind string

const (
	KindNote      PageKind = "note"
	KindEntity    PageKind = "entity"
	KindConcept   PageKind = "concept"
	KindSummary   PageKind = "summary"
	KindSynthesis PageKind = "synthesis"
)

// KindDir mappar en sidtyp till katalogen den ligger i.
func (k PageKind) KindDir() string {
	switch k {
	case KindEntity:
		return "entities"
	case KindConcept:
		return "concepts"
	case KindSummary:
		return "summaries"
	case KindSynthesis:
		return "synthesis"
	default:
		return NotesDir
	}
}

// KindTitle ger svenska rubriker för indexet.
func (k PageKind) KindTitle() string {
	switch k {
	case KindEntity:
		return "Entiteter"
	case KindConcept:
		return "Koncept"
	case KindSummary:
		return "Sammanfattningar"
	case KindSynthesis:
		return "Synteser"
	case KindNote:
		return "Noter"
	default:
		return "Övrigt"
	}
}

// Kinds är de sidtyper som visas i indexet, i den ordning de presenteras.
var Kinds = []PageKind{KindEntity, KindConcept, KindSummary, KindSynthesis, KindNote}

// ParseKind tolkar ett sidtypnamn.
func ParseKind(s string) (PageKind, error) {
	switch PageKind(strings.ToLower(strings.TrimSpace(s))) {
	case KindNote, "":
		return KindNote, nil
	case KindEntity:
		return KindEntity, nil
	case KindConcept:
		return KindConcept, nil
	case KindSummary:
		return KindSummary, nil
	case KindSynthesis:
		return KindSynthesis, nil
	default:
		return "", fmt.Errorf("okänd sidtyp %q (giltiga: note, entity, concept, summary, synthesis)", s)
	}
}

// WriteOptions styr hur en sida skrivs.
type WriteOptions struct {
	Kind    PageKind
	Title   string
	Body    string
	Summary string
	Tags    []string
	Related []string
	Sources []string
	Status  string
	// Replace ersätter brödtexten i stället för att lägga till en
	// daterad uppdateringssektion.
	Replace bool
	// Now gör körningen reproducerbar i tester.
	Now time.Time
}

// WritePage skriver en sida i wiki-lagret och returnerar dess id.
// Finns sidan redan uppdateras den: metadata slås ihop, taggar och
// relaterade sidor adderas, och brödtexten antingen ersätts eller läggs
// till under en daterad uppdateringsrubrik.
func WritePage(dir string, opts WriteOptions) (string, error) {
	if strings.TrimSpace(opts.Title) == "" {
		return "", fmt.Errorf("en sida måste ha en rubrik")
	}

	now := opts.Now
	if now.IsZero() {
		now = time.Now()
	}
	date := now.Format("2006-01-02")

	kind := opts.Kind
	if kind == "" {
		kind = KindNote
	}

	id := Slug(opts.Title)
	if id == "" {
		return "", fmt.Errorf("rubriken %q går inte att göra till ett id", opts.Title)
	}
	pageID := kind.KindDir() + "/" + id
	path := filepath.Join(dir, PagesDir, kind.KindDir(), id+".md")

	existing, err := ParseNote(path)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return "", err
	}

	if existing == nil {
		note := &Note{
			ID:      pageID,
			Title:   opts.Title,
			Type:    string(kind),
			Tags:    opts.Tags,
			Related: opts.Related,
			Sources: opts.Sources,
			Status:  opts.Status,
			Summary: opts.Summary,
			Created: date,
			Updated: date,
			Body:    strings.TrimSpace(opts.Body),
			Path:    path,
		}
		if note.Status == "" {
			note.Status = "active"
		}
		if err := writeFile(path, note.Render()); err != nil {
			return "", err
		}
		return pageID, nil
	}

	existing.Title = opts.Title
	existing.Type = string(kind)
	existing.Tags = MergeTags(existing.Tags, opts.Tags)
	existing.Related = MergeTags(existing.Related, opts.Related)
	existing.Sources = MergeTags(existing.Sources, opts.Sources)
	if opts.Summary != "" {
		existing.Summary = opts.Summary
	}
	if opts.Status != "" {
		existing.Status = opts.Status
	}
	existing.Updated = date

	body := strings.TrimSpace(opts.Body)
	if body == "" {
		body = existing.Body
	} else if opts.Replace {
		existing.Body = body
	} else {
		existing.Body = appendUpdate(existing.Body, body, date)
	}

	return pageID, writeFile(path, existing.Render())
}

// Link lägger till en länk från en sida till en annan.
// Länken hamnar både i frontmatter (related) och i brödtextens
// "## Relaterade"-sektion, så att både maskin och läsare ser den.
func Link(dir, fromID, toID string) error {
	from, err := findPageByID(dir, fromID)
	if err != nil {
		return err
	}
	if _, err := findPageByID(dir, toID); err != nil {
		return fmt.Errorf("målsidan %q finns inte", toID)
	}

	from.Related = MergeTags(from.Related, []string{toID})
	from.Updated = time.Now().Format("2006-01-02")

	body := from.Body
	if !hasRelatedEntry(body, toID) {
		body = appendRelated(body, toID)
	}
	from.Body = body

	return writeFile(from.Path, from.Render())
}

// Unlink tar bort en länk mellan två sidor.
func Unlink(dir, fromID, toID string) error {
	from, err := findPageByID(dir, fromID)
	if err != nil {
		return err
	}

	from.Related = removeTag(from.Related, toID)
	from.Updated = time.Now().Format("2006-01-02")
	from.Body = removeRelated(from.Body, toID)

	return writeFile(from.Path, from.Render())
}

// findPageByID hittar en sida på dess id (case-insensitive) eller,
// om inget hittas, på sista segmentet av id:t.
func findPageByID(dir, id string) (*Note, error) {
	w, err := Open(dir)
	if err != nil {
		return nil, err
	}
	if note := w.Note(id); note != nil {
		return note, nil
	}
	base := filepath.Base(strings.ReplaceAll(id, "\\", "/"))
	for _, note := range w.Notes {
		if strings.EqualFold(filepath.Base(note.ID), base) {
			return note, nil
		}
	}
	return nil, fmt.Errorf("sidan %q finns inte", id)
}

// hasRelatedEntry svarar om brödtexten redan länkar till målet.
func hasRelatedEntry(body, toID string) bool {
	for _, line := range strings.Split(body, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(line), "- ") {
			continue
		}
		if strings.Contains(line, "[["+toID+"]]") || strings.Contains(line, "[["+toID+"|") {
			return true
		}
	}
	return false
}

// appendRelated lägger till en post i "## Relaterade"-sektionen.
// Listan hålls sammanhållen: inga tomma rader mellan posterna.
func appendRelated(body, toID string) string {
	const heading = "## Relaterade"
	entry := "- [[" + toID + "]]"

	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.TrimSpace(line) != heading {
			continue
		}

		// Samla in posterna som hör till sektionen.
		var entries []string
		j := i + 1
		for ; j < len(lines); j++ {
			trimmed := strings.TrimSpace(lines[j])
			if trimmed == "" {
				continue
			}
			if !strings.HasPrefix(trimmed, "- ") {
				break
			}
			entries = append(entries, trimmed)
		}

		entries = append([]string{entry}, entries...)
		out := append([]string{}, lines[:i+1]...)
		out = append(out, "")
		out = append(out, entries...)
		out = append(out, lines[j:]...)
		return strings.Join(out, "\n")
	}

	if strings.TrimSpace(body) == "" {
		return heading + "\n\n" + entry
	}
	return strings.TrimRight(body, "\n") + "\n\n" + heading + "\n\n" + entry + "\n"
}

// removeRelated tar bort en post ur "## Relaterade"-sektionen.
func removeRelated(body, toID string) string {
	lines := strings.Split(body, "\n")
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "- ") &&
			(strings.Contains(line, "[["+toID+"]]") || strings.Contains(line, "[["+toID+"|")) {
			continue
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// removeTag tar bort en tagg ur en lista.
func removeTag(list []string, tag string) []string {
	var out []string
	for _, t := range list {
		if strings.EqualFold(t, tag) {
			continue
		}
		out = append(out, t)
	}
	return out
}
