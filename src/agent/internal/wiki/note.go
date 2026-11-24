// Package wiki implementerar en enkel, fil-baserad kunskapsgraf:
// atomiska anteckningar i markdown, kopplade via länkar, med en inbox
// som agenter skriver till och en ingest-motor som flyttar inläggen in
// i grafen.
package wiki

import (
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
)

// Anteckningsmetadata-nycklar i frontmatter.
const (
	MetaTitle   = "title"
	MetaType    = "type"
	MetaTags    = "tags"
	MetaCreated = "created"
	MetaUpdated = "updated"
	MetaRelated = "related"
	MetaStatus  = "status"
	MetaSources = "sources"
	MetaSource  = "source"
	MetaSummary = "summary"
	MetaOrigin  = "origin"
	MetaAdded   = "added"
)

// Note är en enskild anteckning i wikin.
type Note struct {
	ID      string   // unikt id i grafen
	Title   string   // mänsklig titel
	Type    string   // entity, concept, summary, note ...
	Tags    []string // nyckelord för uppslagning
	Created string   // YYYY-MM-DD
	Updated string   // YYYY-MM-DD
	Related []string // id:n till anteckningar som hör ihop
	Status  string   // t.ex. active, deprecated, open
	Sources []string // råkällor sidan är skriven utifrån
	Summary string   // enradig beskrivning, visas i indexet
	Origin  string   // var källan kommer ifrån (url, fil, etc.)
	Added   string   // datum källan lades in (YYYY-MM-DD)
	Body    string   // själva texten
	Path    string   // sökväg på disk
}

// Slug normaliserar en rubrik till ett filnamnsanpassat id.
func Slug(text string) string {
	replacer := strings.NewReplacer(
		"å", "a", "ä", "a", "ö", "o",
		"Å", "a", "Ä", "a", "Ö", "o",
		"ø", "o", "æ", "ae",
		"_", "-", " ", "-",
	)
	s := strings.ToLower(replacer.Replace(text))

	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
		default:
			b.WriteRune('-')
		}
	}

	slug := strings.Trim(b.String(), "-")
	for strings.Contains(slug, "--") {
		slug = strings.ReplaceAll(slug, "--", "-")
	}
	return slug
}

// ParseNote läser en anteckning från en sökväg.
func ParseNote(path string) (*Note, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kunde inte läsa %s: %w", path, err)
	}

	meta, body, err := splitFrontmatter(string(data))
	if err != nil {
		return nil, fmt.Errorf("kunde inte tolka frontmatter i %s: %w", path, err)
	}

	id := strings.TrimSuffix(filepath.Base(path), ".md")
	title := meta[MetaTitle]
	if title == "" {
		title = headingTitle(body)
	}
	if title == "" {
		title = id
	}

	return &Note{
		ID:      id,
		Title:   title,
		Type:    meta[MetaType],
		Tags:    splitList(meta[MetaTags]),
		Created: meta[MetaCreated],
		Updated: meta[MetaUpdated],
		Related: splitList(meta[MetaRelated]),
		Status:  meta[MetaStatus],
		Sources: mergeSources(meta[MetaSources], meta[MetaSource]),
		Summary: meta[MetaSummary],
		Origin:  meta[MetaOrigin],
		Added:   meta[MetaAdded],
		Body:    strings.TrimSpace(body),
		Path:    path,
	}, nil
}

// headingTitle letar upp första rubriken på nivå 1 i brödtexten.
func headingTitle(body string) string {
	for _, line := range strings.Split(body, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "# "))
		}
	}
	return ""
}

// splitFrontmatter delar upp en fil i metadata och brödtext.
// En fil utan frontmatter tolkas som ren brödtext.
func splitFrontmatter(content string) (map[string]string, string, error) {
	meta := map[string]string{}

	normalized := strings.ReplaceAll(content, "\r\n", "\n")
	if !strings.HasPrefix(normalized, "---\n") {
		return meta, normalized, nil
	}

	rest := normalized[len("---\n"):]
	end := strings.Index(rest, "\n---")
	if end < 0 {
		return nil, "", fmt.Errorf("frontmatter saknar avslutande ---")
	}

	header := rest[:end]
	body := rest[end+len("\n---"):]
	body = strings.TrimPrefix(body, "\n")

	for _, line := range strings.Split(header, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, found := strings.Cut(line, ":")
		if !found {
			return nil, "", fmt.Errorf("ogiltig metadata-rad: %q", line)
		}
		meta[strings.TrimSpace(key)] = strings.TrimSpace(value)
	}

	return meta, body, nil
}

// splitList tolkar "a, b", "[a, b]" och "a" som en lista.
func splitList(raw string) []string {
	raw = strings.TrimSpace(raw)
	raw = strings.Trim(raw, "[]")
	if raw == "" {
		return nil
	}

	parts := strings.Split(raw, ",")
	list := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			list = append(list, p)
		}
	}
	return list
}

// mergeSources slår ihop "sources" och "source" så att båda
// skrivsätt fungerar.
func mergeSources(sources, source string) []string {
	list := splitList(sources)
	for _, s := range splitList(source) {
		if !slices.Contains(list, s) {
			list = append(list, s)
		}
	}
	return list
}

// BodyLinks returnerar de länkar som finns i brödtexten,
// antingen som markdown-länkar eller som [[wikilänkar]].
// Kods spans (`...`) hoppas över - prose som beskriver länksyntax
// ska inte räknas som en länk.
func (n *Note) BodyLinks() []string {
	seen := map[string]bool{}
	var links []string

	add := func(target string) {
		target = strings.TrimSpace(target)
		if target == "" {
			return
		}
		key := strings.ToLower(target)
		if seen[key] {
			return
		}
		seen[key] = true
		links = append(links, target)
	}

	body := codeSpanPattern.ReplaceAllString(n.Body, " ")

	for _, m := range markdownLinkPattern.FindAllStringSubmatch(body, -1) {
		add(m[1])
	}
	for _, m := range wikilinkPattern.FindAllStringSubmatch(body, -1) {
		add(m[1])
	}

	return links
}

// Links returnerar de sidor som anteckningen pekar på.
func (n *Note) Links(w *Wiki) []*Note {
	var out []*Note
	for _, target := range n.AllRelated() {
		if resolved := w.Resolve(target); resolved != nil {
			out = append(out, resolved)
		}
	}
	return out
}

// AllRelated returnerar både frontmatter-länkar och länkar i brödtexten.
func (n *Note) AllRelated() []string {
	seen := map[string]bool{n.ID: true}
	var all []string

	for _, id := range append(append([]string{}, n.Related...), n.BodyLinks()...) {
		if id == "" || seen[strings.ToLower(id)] {
			continue
		}
		seen[strings.ToLower(id)] = true
		all = append(all, id)
	}
	return all
}

// HasTag svarar på om anteckningen har en given tagg.
func (n *Note) HasTag(tag string) bool {
	for _, t := range n.Tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

// Render serialiserar anteckningen tillbaka till markdown med frontmatter.
func (n *Note) Render() string {
	var b strings.Builder

	b.WriteString("---\n")
	fmt.Fprintf(&b, "%s: %s\n", MetaTitle, n.Title)
	if n.Type != "" {
		fmt.Fprintf(&b, "%s: %s\n", MetaType, n.Type)
	}
	if n.Summary != "" {
		fmt.Fprintf(&b, "%s: %s\n", MetaSummary, n.Summary)
	}
	if len(n.Tags) > 0 {
		fmt.Fprintf(&b, "%s: %s\n", MetaTags, formatList(n.Tags))
	}
	if n.Origin != "" {
		fmt.Fprintf(&b, "%s: %s\n", MetaOrigin, n.Origin)
	}
	if len(n.Sources) > 0 {
		fmt.Fprintf(&b, "%s: %s\n", MetaSources, formatList(n.Sources))
	}
	if n.Created != "" {
		fmt.Fprintf(&b, "%s: %s\n", MetaCreated, n.Created)
	}
	if n.Updated != "" {
		fmt.Fprintf(&b, "%s: %s\n", MetaUpdated, n.Updated)
	}
	if n.Status != "" {
		fmt.Fprintf(&b, "%s: %s\n", MetaStatus, n.Status)
	}
	if len(n.Related) > 0 {
		fmt.Fprintf(&b, "%s: %s\n", MetaRelated, formatList(n.Related))
	}
	b.WriteString("---\n\n")
	b.WriteString(n.Body)
	b.WriteString("\n")

	return b.String()
}

func formatList(list []string) string {
	return "[" + strings.Join(list, ", ") + "]"
}

// MergeTags slår ihop två tagg-listor utan dubbletter, sorterade.
func MergeTags(a, b []string) []string {
	seen := map[string]bool{}
	var merged []string

	for _, tag := range append(append([]string{}, a...), b...) {
		key := strings.ToLower(tag)
		if tag == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, tag)
	}

	sort.Strings(merged)
	return merged
}
