// Package goaldefinition bygger agentens förslag på hur ett mål ska se ut.
//
// Paketet bär policyn: agenten får fylla luckor, aldrig skriva över
// värden som användaren redan angett. Förslaget lämnas till en människa
// som godkänner det, och det är exakt de godkända värdena som sedan
// appliceras - inte en ny generering.
package goaldefinition

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"plan/agent/internal/definer"
	"plan/agent/internal/planclient"

	"github.com/teambition/rrule-go"
)

// Kind är godkännandets typ för en måldefinition.
const Kind = planclient.ApprovalKindGoalDefinition

// Values är de exakta värden som föreslås.
//
// Ett fält som inte finns här lämnas orört. PeKAR används så att "sätt
// värdet" går att skilja från "lämna oförändrat".
type Values struct {
	Title       *string `json:"title,omitempty"`
	Description *string `json:"description,omitempty"`
	Icon        *string `json:"icon,omitempty"`
	Now         *bool   `json:"now,omitempty"`
	Eternal     *bool   `json:"eternal,omitempty"`
	Agent       *bool   `json:"agent,omitempty"`
	Due         *string `json:"due,omitempty"`
	StartDate   *string `json:"start_date,omitempty"`
	// EstimatedMinutes är den uppskattade tidsåtgången, i minuter.
	EstimatedMinutes *int `json:"estimated_minutes,omitempty"`
	// RecurrenceRule är upprepningsregeln ("FREQ=WEEKLY;BYDAY=MO").
	RecurrenceRule *string `json:"recurrence_rule,omitempty"`
	// RecurrenceMode är schemaläggningsstrategin ("schedule" eller "push").
	RecurrenceMode *string `json:"recurrence_mode,omitempty"`
	// RecurrenceStartDate är seriens startdatum ("YYYY-MM-DD").
	RecurrenceStartDate *string `json:"recurrence_start_date,omitempty"`
	// Tags är målets fulla tagglista efter ändringen.
	//
	// API:t ersätter hela uppsättningen, så listan måste innehålla de
	// taggar målet redan har.
	Tags []string `json:"tags,omitempty"`
	// PlaceIDs är målets fulla platslista efter ändringen.
	PlaceIDs []int64 `json:"place_ids,omitempty"`
	// Status är fasen målet flyttas till.
	Status *string `json:"status,omitempty"`
}

// Field är en ändring som föreslås - den visas för människan.
type Field struct {
	// Key är fältets namn, så att det går att hitta det i values.
	Key string `json:"field"`
	// Label är fältets namn på svenska.
	Label string `json:"label"`
	// From är målets nuvarande värde. Tomt = inget värde.
	From string `json:"from"`
	// To är det föreslagna värdet.
	To string `json:"to"`
}

// Proposal är agentens förslag på hur ett mål ska se ut.
type Proposal struct {
	// Kind är godkännandets typ.
	Kind string `json:"kind"`
	// GoalID är målet förslaget gäller.
	GoalID int64 `json:"goal_id"`
	// Fields är ändringarna som föreslås, i läsbar form.
	Fields []Field `json:"fields"`
	// Values är de exakta värdena som appliceras om förslaget godkänns.
	Values Values `json:"values"`
	// Notes är antaganden agenten gjort.
	Notes []string `json:"notes,omitempty"`
}

// Context är sammanhanget ett förslag byggs i.
type Context struct {
	// Goal är målet som ska definieras.
	Goal planclient.Goal
	// Definition är agentens förslag.
	Definition definer.Definition
	// Places är användarens platser. Agenten föreslår bara platser som
	// finns bland dem.
	Places []planclient.Place
	// CurrentPhase är målets nuvarande fas.
	CurrentPhase string
	// NextPhase är fasen målet flyttas till när definitionen är godkänd.
	NextPhase string
}

// Build bygger förslaget utifrån målet och agentens definition.
//
// Policyn är att agenten fyller luckor. Ett fält som redan har ett värde
// lämnas orört - även om agenten föreslår något annat. Taggar och platser
// slås ihop med målets befintliga, så inget användardata försvinner.
func Build(ctx Context) Proposal {
	proposal := Proposal{
		Kind:   Kind,
		GoalID: ctx.Goal.ID,
		Fields: []Field{},
		Notes:  append([]string{}, ctx.Definition.Notes...),
	}

	definition := ctx.Definition

	// Titel och beskrivning är definitionens kärna - de föreslås alltid,
	// men bara om de faktiskt ändrar något.
	if definition.Title != "" && definition.Title != ctx.Goal.Title {
		value := definition.Title
		proposal.Values.Title = &value
		proposal.add("title", "Titel", ctx.Goal.Title, value)
	}

	if definition.Description != "" && definition.Description != ctx.Goal.Description {
		value := definition.Description
		proposal.Values.Description = &value
		proposal.add("description", "Beskrivning", ctx.Goal.Description, value)
	}

	// Ikon - bara om målet saknar en.
	if ctx.Goal.Icon == "" && definition.Icon != "" {
		if validIcon(definition.Icon) {
			value := definition.Icon
			proposal.Values.Icon = &value
			proposal.add("icon", "Ikon", "", value)
		} else {
			proposal.note(fmt.Sprintf("ikonen %q är inte ett giltigt ikonnamn och hoppades över", definition.Icon))
		}
	}

	// Operativa flaggor - bara sätts på, aldrig av.
	if !ctx.Goal.Now && definition.Now {
		value := true
		proposal.Values.Now = &value
		proposal.add("now", "Gör nu", "nej", "ja")
	}

	if !ctx.Goal.Eternal && definition.Eternal {
		value := true
		proposal.Values.Eternal = &value
		proposal.add("eternal", "Evigt mål", "nej", "ja")
	}

	if !ctx.Goal.Agent && definition.Agent {
		value := true
		proposal.Values.Agent = &value
		proposal.add("agent", "Agentmål", "nej", "ja")
	}

	// Datum - bara om målet saknar dem.
	if ctx.Goal.Due == nil && definition.Due != "" {
		if date, ok := validDate(definition.Due); ok {
			proposal.Values.Due = &date
			proposal.add("due", "Slutdatum", "", date)
		} else {
			proposal.note(fmt.Sprintf("slutdatumet %q är inte ett giltigt datum (YYYY-MM-DD) och hoppades över", definition.Due))
		}
	}

	if ctx.Goal.StartDate == nil && definition.StartDate != "" {
		if date, ok := validDate(definition.StartDate); ok {
			proposal.Values.StartDate = &date
			proposal.add("start_date", "Startdatum", "", date)
		} else {
			proposal.note(fmt.Sprintf("startdatumet %q är inte ett giltigt datum (YYYY-MM-DD) och hoppades över", definition.StartDate))
		}
	}

	// Uppskattad tid - bara om målet saknar en.
	if ctx.Goal.EstimatedMinutes == nil && definition.EstimatedMinutes > 0 {
		value := definition.EstimatedMinutes
		proposal.Values.EstimatedMinutes = &value
		proposal.add("estimated_minutes", "Uppskattad tid", "", strconv.Itoa(value)+" min")
	}

	// Upprepning - bara om målet saknar en regel.
	if ctx.Goal.RecurrenceRule == nil && definition.RecurrenceRule != "" {
		if err := validateRule(definition.RecurrenceRule); err != nil {
			proposal.note(fmt.Sprintf("upprepningsregeln %q är inte giltig och hoppades över", definition.RecurrenceRule))
		} else {
			rule := definition.RecurrenceRule
			proposal.Values.RecurrenceRule = &rule
			proposal.add("recurrence_rule", "Upprepning", "", rule)

			mode := definition.RecurrenceMode
			if mode == "" {
				mode = "schedule"
			}
			if mode != "schedule" && mode != "push" {
				proposal.note(fmt.Sprintf("upprepningsstrategin %q är okänd - \"schedule\" användes i stället", mode))
				mode = "schedule"
			}
			proposal.Values.RecurrenceMode = &mode
			proposal.add("recurrence_mode", "Upprepningsstrategi", "", mode)

			start := definition.RecurrenceStartDate
			if start == "" {
				start = definition.StartDate
			}
			if start == "" {
				start = definition.Due
			}
			if start != "" {
				if date, ok := validDate(start); ok {
					proposal.Values.RecurrenceStartDate = &date
					proposal.add("recurrence_start_date", "Upprepning startar", "", date)
				} else {
					proposal.note(fmt.Sprintf("startdatumet för upprepningen %q är ogiltigt och hoppades över", start))
				}
			}
		}
	}

	// Taggar - slås ihop med målets egna.
	if tags, added := mergeTags(ctx.Goal.Tags, definition.Tags); added {
		proposal.Values.Tags = tags
		proposal.add("tags", "Taggar", strings.Join(ctx.Goal.Tags, ", "), strings.Join(tags, ", "))
	}

	// Platser - slås ihop med målets egna.
	if ids, names, added := mergePlaces(ctx.Goal.Places, definition.Places, ctx.Places); added {
		proposal.Values.PlaceIDs = ids
		proposal.add("places", "Platser", strings.Join(placeNames(ctx.Goal.Places), ", "), strings.Join(names, ", "))
	}

	// Fasen - målet går vidare när definitionen är godkänd.
	if ctx.NextPhase != "" {
		value := ctx.NextPhase
		proposal.Values.Status = &value
		proposal.add("status", "Fas", ctx.CurrentPhase, value)
	}

	return proposal
}

// add lägger till en ändring i förslaget.
func (p *Proposal) add(key, label, from, to string) {
	p.Fields = append(p.Fields, Field{Key: key, Label: label, From: from, To: to})
}

// note lägger till en notering om något som hoppades över.
func (p *Proposal) note(text string) {
	if text == "" {
		return
	}
	p.Notes = append(p.Notes, text)
}

// Payload gör om förslaget till något som kan sparas i Plan.
//
// Värdet är exakt JSON-strukturen - det är den som godkänns och sedan
// appliceras.
func (p Proposal) Payload() map[string]any {
	raw, err := json.Marshal(p)
	if err != nil {
		return map[string]any{}
	}

	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

// FromPayload läser ett förslag ur ett godkännandes payload.
//
// Det är så här exakt det som människan godkände läses tillbaka - utan
// att något genereras om.
func FromPayload(payload map[string]any) (Proposal, error) {
	if payload == nil {
		return Proposal{}, fmt.Errorf("godkännandet saknar det föreslagna innehållet")
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return Proposal{}, fmt.Errorf("kunde inte tolka förslaget: %w", err)
	}

	var proposal Proposal
	if err := json.Unmarshal(raw, &proposal); err != nil {
		return Proposal{}, fmt.Errorf("kunde inte tolka förslaget: %w", err)
	}

	if proposal.Kind == "" {
		proposal.Kind = Kind
	}

	return proposal, nil
}

// Apply skriver de föreslagna värdena till målet.
//
// Funktionen applicerar exakt de värden som stod i förslaget - den
// genererar inte om något.
func (p Proposal) Apply(ctx context.Context, client *planclient.Client, goalID int64) error {
	request := planclient.UpdateGoalRequest{
		Title:               p.Values.Title,
		Description:         p.Values.Description,
		Icon:                p.Values.Icon,
		Now:                 p.Values.Now,
		Eternal:             p.Values.Eternal,
		Agent:               p.Values.Agent,
		Due:                 p.Values.Due,
		StartDate:           p.Values.StartDate,
		EstimatedMinutes:    p.Values.EstimatedMinutes,
		RecurrenceRule:      p.Values.RecurrenceRule,
		RecurrenceMode:      p.Values.RecurrenceMode,
		RecurrenceStartDate: p.Values.RecurrenceStartDate,
	}

	if !isEmpty(request) {
		if _, err := client.UpdateGoal(ctx, goalID, request); err != nil {
			return fmt.Errorf("failed to update goal: %w", err)
		}
	}

	// Taggar och platser ersätter hela uppsättningen, så de kommer
	// efter målets övriga fält.
	if len(p.Values.Tags) > 0 {
		if err := client.SetGoalTags(ctx, goalID, p.Values.Tags); err != nil {
			return fmt.Errorf("failed to set goal tags: %w", err)
		}
	}

	if len(p.Values.PlaceIDs) > 0 {
		if err := client.SetGoalPlaces(ctx, goalID, p.Values.PlaceIDs); err != nil {
			return fmt.Errorf("failed to set goal places: %w", err)
		}
	}

	// Statusen ändras via den dedikerade endpointen - den vanliga
	// målendpointen ignorerar status_id.
	if p.Values.Status != nil && *p.Values.Status != "" {
		if err := client.UpdateGoalPhase(ctx, goalID, *p.Values.Status); err != nil {
			return fmt.Errorf("failed to update goal phase: %w", err)
		}
	}

	return nil
}

// isEmpty berättar om en uppdatering innehåller något alls.
func isEmpty(request planclient.UpdateGoalRequest) bool {
	return request.Title == nil &&
		request.Description == nil &&
		request.Icon == nil &&
		request.Now == nil &&
		request.Eternal == nil &&
		request.Agent == nil &&
		request.Due == nil &&
		request.StartDate == nil &&
		request.EstimatedMinutes == nil &&
		request.RecurrenceRule == nil &&
		request.RecurrenceMode == nil &&
		request.RecurrenceStartDate == nil
}

// HasDataFields berättar om förslaget innehåller andra data än fasen.
//
// Fasen är agentens eget steg i flödet - den behöver inte godkännas av
// en människa. Övriga fält är uppgifter på målet, och de ska en människa
// ta ställning till.
func (p Proposal) HasDataFields() bool {
	for _, field := range p.Fields {
		if field.Key != "status" {
			return true
		}
	}
	return false
}

// Summary beskriver förslaget för människan som ska ta ställning.
func (p Proposal) Summary() string {
	if len(p.Fields) == 0 {
		return "Agenten har inget nytt att föreslå för målet."
	}

	var sb strings.Builder
	sb.WriteString("Agenten föreslår att målet får följande ändringar:\n\n")

	for _, field := range p.Fields {
		from := field.From
		if from == "" {
			from = "ingen"
		}
		fmt.Fprintf(&sb, "- %s: %s → %s\n", field.Label, from, field.To)
	}

	if len(p.Notes) > 0 {
		sb.WriteString("\nNoteringar:\n")
		for _, note := range p.Notes {
			sb.WriteString("- " + note + "\n")
		}
	}

	return strings.TrimRight(sb.String(), "\n")
}

// Consequence beskriver vad förslaget innebär.
func (p Proposal) Consequence() string {
	if len(p.Fields) == 0 {
		return "Inget ändras."
	}

	return fmt.Sprintf("%d fält ändras på målet. Inget värde tas bort.", len(p.Fields))
}

// mergeTags slår ihop målets taggar med agentens förslag.
//
// Agentens taggar läggs till - målets egna tas aldrig bort. Agentens
// förslag kan aldrig ta bort en användares tagg.
func mergeTags(existing []string, proposed []string) ([]string, bool) {
	merged := make([]string, 0, len(existing)+len(proposed))
	seen := make(map[string]bool)

	for _, tag := range existing {
		name := normalizeTag(tag)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		merged = append(merged, name)
	}

	added := false
	for _, tag := range proposed {
		name := normalizeTag(tag)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		merged = append(merged, name)
		added = true
	}

	return merged, added
}

// normalizeTag gör om ett taggnamn till den form Plan sparar det i.
//
// Samma regler som i Plan: namnet Trimmas, "#" tas bort och versaler
// förvandlas till gemener.
func normalizeTag(raw string) string {
	name := strings.TrimSpace(raw)
	name = strings.TrimPrefix(name, "#")
	name = strings.ToLower(name)
	name = strings.Join(strings.Fields(name), " ")
	return name
}

// mergePlaces slår ihop målets platser med agentens förslag.
//
// Agenten föreslår platsnamn. Bara platser som redan finns hos användaren
// godtas - agenten skapar inga nya platser.
func mergePlaces(existing []planclient.Place, proposed []string, known []planclient.Place) ([]int64, []string, bool) {
	ids := make([]int64, 0, len(existing)+len(proposed))
	names := make([]string, 0, len(existing)+len(proposed))
	seen := make(map[int64]bool)

	for _, place := range existing {
		if seen[place.ID] {
			continue
		}
		seen[place.ID] = true
		ids = append(ids, place.ID)
		names = append(names, place.Name)
	}

	added := false
	for _, name := range proposed {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}

		place, found := findPlace(trimmed, known)
		if !found {
			continue
		}
		if seen[place.ID] {
			continue
		}

		seen[place.ID] = true
		ids = append(ids, place.ID)
		names = append(names, place.Name)
		added = true
	}

	return ids, names, added
}

// findPlace leter upp en plats på sitt namn.
func findPlace(name string, known []planclient.Place) (planclient.Place, bool) {
	needle := strings.ToLower(strings.TrimSpace(name))
	for _, place := range known {
		if strings.ToLower(strings.TrimSpace(place.Name)) == needle {
			return place, true
		}
	}
	return planclient.Place{}, false
}

// placeNames listar platsernas namn.
func placeNames(places []planclient.Place) []string {
	names := make([]string, 0, len(places))
	for _, place := range places {
		names = append(names, place.Name)
	}
	return names
}

// validDate kontrollerar att ett datum går att tolka som "YYYY-MM-DD".
func validDate(value string) (string, bool) {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return "", false
	}
	if _, err := time.Parse("2006-01-02", trimmed); err != nil {
		return "", false
	}
	return trimmed, true
}

// validIcon kontrollerar att ikonen går att visa.
//
// En ikon är antingen en emoji eller ett Iconify-namn ("kollektion:namn").
func validIcon(value string) bool {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return false
	}

	if strings.Contains(trimmed, ":") {
		parts := strings.SplitN(trimmed, ":", 2)
		return isIconName(parts[0]) && isIconName(parts[1])
	}

	// En emoji - högst några tecken, varav minst ett är utanför ASCII.
	runes := []rune(trimmed)
	if len(runes) > 4 {
		return false
	}
	for _, rune := range runes {
		if rune > 127 {
			return true
		}
	}
	return false
}

// isIconName kontrollerar att en del av ett Iconify-namn är giltig.
func isIconName(value string) bool {
	if value == "" {
		return false
	}
	for _, rune := range value {
		switch {
		case rune >= 'a' && rune <= 'z':
		case rune >= '0' && rune <= '9':
		case rune == '-':
		default:
			return false
		}
	}
	return true
}

// validateRule kontrollerar att en upprepningsregel är giltig.
//
// Samma krav som Plan ställer: regeln ska vara en RRULE, och INTERVAL
// och COUNT ska vara minst 1.
func validateRule(rule string) error {
	normalized := strings.ToUpper(strings.TrimSpace(rule))
	if normalized == "" {
		return fmt.Errorf("regeln är tom")
	}

	normalized = strings.TrimPrefix(normalized, "RRULE:")

	for _, part := range strings.Split(normalized, ";") {
		key, value, found := strings.Cut(part, "=")
		if !found {
			continue
		}

		switch strings.TrimSpace(key) {
		case "INTERVAL", "COUNT":
			number, err := strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return fmt.Errorf("%s har ogiltigt värde %q", key, value)
			}
			if number < 1 {
				return fmt.Errorf("%s måste vara minst 1", key)
			}
		}
	}

	if _, err := rrule.StrToRRule(normalized); err != nil {
		return fmt.Errorf("regeln kunde inte tolkas: %w", err)
	}

	return nil
}
