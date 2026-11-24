package wiki

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// LintKind kategoriserar ett lintfynd.
type LintKind string

const (
	LintBrokenLink   LintKind = "trasig länk"
	LintDuplicate    LintKind = "dubblerad sida"
	LintDuplicateSec LintKind = "dubblerad sektion"
	LintOrphan       LintKind = "ödelapp"
	LintNoSummary    LintKind = "saknar sammanfattning"
	LintStale        LintKind = "föråldrad"
	LintUnintegrated LintKind = "ointegrerad källa"
)

// LintFinding är ett problem som linten hittade.
type LintFinding struct {
	Kind LintKind
	Page string
	Msg  string
}

// String beskriver fyndet på svenska.
func (f LintFinding) String() string {
	if f.Page == "" {
		return fmt.Sprintf("%s: %s", f.Kind, f.Msg)
	}
	return fmt.Sprintf("%s: %s - %s", f.Kind, f.Page, f.Msg)
}

// LintOptions styr hur linten körs.
type LintOptions struct {
	// StaleAfter anger hur länge en sida får vara ouppdaterad
	// innan den flaggas. Noll betyder att ingen ålderskontroll körs.
	StaleAfter time.Duration
	// Now används för att göra körningen reproducerbar i tester.
	Now time.Time
}

// Lint undersöker wikin och returnerar en lista med problem.
// Den letar efter trasiga länkar, dubblerade sidor, ödelappar utan
// inkommande länkar, sidor utan sammanfattning, föråldrade sidor
// och råkällor som aldrig integrerats i wikin.
func (w *Wiki) Lint(opts LintOptions) []LintFinding {
	var findings []LintFinding

	// Trasiga länkar.
	for _, note := range w.Notes {
		for _, target := range note.AllRelated() {
			if w.Resolve(target) == nil {
				findings = append(findings, LintFinding{
					Kind: LintBrokenLink,
					Page: note.ID,
					Msg:  fmt.Sprintf("pekar på %q som inte finns", target),
				})
			}
		}
	}

	// Dubblerade sidor: samma rubrik på flera ställen.
	byTitle := map[string][]string{}
	for _, note := range w.Notes {
		key := strings.ToLower(strings.TrimSpace(note.Title))
		byTitle[key] = append(byTitle[key], note.ID)
	}
	titles := make([]string, 0, len(byTitle))
	for title := range byTitle {
		titles = append(titles, title)
	}
	sort.Strings(titles)
	for _, title := range titles {
		ids := byTitle[title]
		if len(ids) < 2 {
			continue
		}
		findings = append(findings, LintFinding{
			Kind: LintDuplicate,
			Page: strings.Join(ids, ", "),
			Msg:  fmt.Sprintf("samma rubrik %q finns på flera ställen", title),
		})
	}

	// Dubblerade sektioner: samma rubrik två gånger på en sida.
	// Tyder på att innehåll har klistrats in flera gånger.
	for _, note := range w.Notes {
		seen := map[string]bool{}
		for _, heading := range sectionHeadings(note.Body) {
			if seen[heading] {
				findings = append(findings, LintFinding{
					Kind: LintDuplicateSec,
					Page: note.ID,
					Msg:  fmt.Sprintf("rubriken %q förekommer flera gånger", heading),
				})
				continue
			}
			seen[heading] = true
		}
	}

	// Ödelappar: sidor som ingen annan sida länkar till.
	for _, note := range w.Notes {
		if len(w.Backlinks(note.ID)) == 0 {
			findings = append(findings, LintFinding{
				Kind: LintOrphan,
				Page: note.ID,
				Msg:  "inga inkommande länkar",
			})
		}
	}

	// Sidor utan beskrivning gör indexet svårläst.
	for _, note := range w.Notes {
		if strings.TrimSpace(note.Summary) == "" && DeriveSummary(note.Body) == "" {
			findings = append(findings, LintFinding{
				Kind: LintNoSummary,
				Page: note.ID,
				Msg:  "sidan saknar beskrivning - lägg till summary i frontmatter",
			})
		}
	}

	// Föråldrade sidor: äldre än StaleAfter och inte uppdaterade sedan dess.
	if opts.StaleAfter > 0 {
		now := opts.Now
		if now.IsZero() {
			now = time.Now()
		}
		cutoff := now.Add(-opts.StaleAfter)
		for _, note := range w.Notes {
			updated, err := time.Parse("2006-01-02", note.Updated)
			if err != nil {
				continue
			}
			if updated.Before(cutoff) {
				findings = append(findings, LintFinding{
					Kind: LintStale,
					Page: note.ID,
					Msg:  fmt.Sprintf("inte uppdaterad sedan %s", note.Updated),
				})
			}
		}
	}

	// Råkällor som aldrig integrerats: en datalucka i wikin.
	for _, source := range w.Sources() {
		if !source.Indexed {
			findings = append(findings, LintFinding{
				Kind: LintUnintegrated,
				Page: source.ID,
				Msg:  "råkällan nämns inte av någon sida - skriv en sammanfattning",
			})
		}
	}

	return findings
}

// sectionHeadings returnerar alla rubriker på nivå 2 i brödtexten.
func sectionHeadings(body string) []string {
	var out []string
	inCode := false
	for _, line := range strings.Split(body, "\n") {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inCode = !inCode
			continue
		}
		if inCode {
			continue
		}
		if strings.HasPrefix(trimmed, "## ") {
			out = append(out, strings.TrimSpace(strings.TrimPrefix(trimmed, "## ")))
		}
	}
	return out
}
