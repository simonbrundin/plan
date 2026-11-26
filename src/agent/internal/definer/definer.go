// Package definer tar fram agentens förslag på hur ett mål ska se ut.
//
// Definitionen är agentens egen - den är ett förslag, inte ett facit.
// Agenten fyller så många fält som möjligt, men skriver aldrig till målet
// själv. Förslaget lämnas till en människa som bestämmer vad som ska
// gälla (se paketet goaldefinition).
package definer

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"plan/agent/internal/core/messages"
	"plan/agent/internal/core/provider"
)

// Maximal längd på en titel (antal tecken).
const maxTitleLength = 50

// Definition represents a goal definition
//
// Fälten är de målfält agenten kan fylla. Ett tomt fält betyder att
// agenten inte har något att föreslå för det - då lämnas målets värde
// orört.
type Definition struct {
	Title       string `json:"title"`
	Description string `json:"description"`
	// Icon är en emoji eller ett Iconify-namn ("heroicons:flag").
	Icon string `json:"icon,omitempty"`
	// Now betyder att målet ska göras nu.
	Now bool `json:"now,omitempty"`
	// Eternal betyder att målet aldrig slutar.
	Eternal bool `json:"eternal,omitempty"`
	// Agent betyder att målet är avsett för en agent.
	Agent bool `json:"agent,omitempty"`
	// Due är det datum målet ska vara klart ("YYYY-MM-DD").
	Due string `json:"due,omitempty"`
	// StartDate är det datum målet får dyka upp i prioriteringslistan.
	StartDate string `json:"start_date,omitempty"`
	// EstimatedMinutes är den uppskattade tidsåtgången, i minuter.
	EstimatedMinutes int `json:"estimated_minutes,omitempty"`
	// RecurrenceRule är upprepningsregeln ("FREQ=WEEKLY;BYDAY=MO").
	RecurrenceRule string `json:"recurrence_rule,omitempty"`
	// RecurrenceMode är schemaläggningsstrategin ("schedule" eller "push").
	RecurrenceMode string `json:"recurrence_mode,omitempty"`
	// RecurrenceStartDate är seriens startdatum ("YYYY-MM-DD").
	RecurrenceStartDate string `json:"recurrence_start_date,omitempty"`
	// Tags är taggar som passar målet.
	Tags []string `json:"tags,omitempty"`
	// Places är namn på platser där målet kan slutföras.
	//
	// Namnen måste finnas bland användarens platser - agenten skapar
	// inga nya platser.
	Places []string `json:"places,omitempty"`
	// Notes är antaganden agenten gjort, som en människa bör känna till.
	Notes []string `json:"notes,omitempty"`
}

// Input är målet som ska definieras, tillsammans med de värden som
// redan finns.
//
// De befintliga värdena skickas med så att agenten inte föreslår något
// som redan är satt - och så att den kan bygga vidare på dem.
type Input struct {
	Title       string
	Description string
	// Feedback är en återkoppling från en människa på ett tidigare förslag.
	Feedback string
	// Icon är målets nuvarande ikon.
	Icon string
	// Now, Eternal och Agent är målets nuvarande flaggor.
	Now     bool
	Eternal bool
	Agent   bool
	// Due och StartDate är målets nuvarande datum ("YYYY-MM-DD").
	Due       string
	StartDate string
	// EstimatedMinutes är målets nuvarande uppskattning, i minuter.
	EstimatedMinutes int
	// RecurrenceRule är målets nuvarande upprepningsregel.
	RecurrenceRule string
	// Tags är målets nuvarande taggar.
	Tags []string
	// Places är de platser målet gäller, med namn.
	Places []string
}

// Definer uses an LLM to define goals
type Definer struct {
	provider provider.LLMProvider
	model    provider.Model
}

// New creates a new goal definer
func New(provider provider.LLMProvider, model provider.Model) *Definer {
	return &Definer{
		provider: provider,
		model:    model,
	}
}

// Define creates a goal definition from an initial goal title
func (d *Definer) Define(ctx context.Context, input Input) (*Definition, error) {
	systemPrompt := `You are a software requirements analyst. Your job is to take a loose goal or idea and turn it into a well-defined goal.

Answer with ONE JSON object and nothing else. No prose, no markdown fences.

The object has these fields:
{
  "title": "concise title, max 50 characters",
  "description": "what the goal accomplishes, key requirements, success criteria and scope",
  "icon": "one emoji or an iconify name like heroicons:flag",
  "now": false,
  "eternal": false,
  "agent": false,
  "due": "YYYY-MM-DD or null",
  "start_date": "YYYY-MM-DD or null",
  "estimated_minutes": 0,
  "recurrence_rule": "RFC 5545 RRULE such as FREQ=WEEKLY;BYDAY=MO, or null",
  "recurrence_mode": "schedule or push",
  "recurrence_start_date": "YYYY-MM-DD or null",
  "tags": ["short tags"],
  "places": ["names of places the goal applies to"],
  "notes": ["assumptions you made"]
}

Rules:
- Leave a field out or set it to null when you have nothing to say about it. Never invent a value just to fill the field.
- due, start_date and recurrence_start_date are calendar dates in YYYY-MM-DD.
- estimated_minutes is a realistic estimate for one occurrence of the goal. Use 0 when you cannot estimate.
- Only use places that are listed in the goal's existing places, or that you are told about. Use the exact names.
- now is true only for something that should be done right now. eternal is true for goals that never end.
- agent is true when a person should hand the goal over to an agent.
- tags are short and lowercase.

Be specific and actionable. Avoid vague language like "improve" or "optimize" without specifics.`

	userPrompt := buildUserPrompt(input)

	msgs := []messages.Message{
		&messages.SystemMessage{
			Role:      messages.RoleSystem,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: systemPrompt}},
			Timestamp: 0,
		},
		&messages.UserMessage{
			Role:      messages.RoleUser,
			Content:   []messages.ContentBlock{&messages.TextBlock{Text: userPrompt}},
			Timestamp: 0,
		},
	}

	stream, err := d.provider.Stream(ctx, d.model, msgs, nil, provider.StreamingOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to stream response: %w", err)
	}

	var response strings.Builder
	for event := range stream {
		switch e := event.(type) {
		case *provider.TextDeltaEvent:
			response.WriteString(e.Delta)
		case *provider.DoneEvent:
			return parseDefinition(response.String())
		case *provider.ErrorEvent:
			return nil, e.Error
		}
	}

	return nil, fmt.Errorf("stream ended without completion")
}

// buildUserPrompt bygger frågan till modellen.
func buildUserPrompt(input Input) string {
	var sb strings.Builder

	sb.WriteString("## GOAL TO DEFINE\n\n")
	sb.WriteString(input.Title)
	sb.WriteString("\n\n")

	if input.Description != "" {
		sb.WriteString("## EXISTING DESCRIPTION\n\n")
		sb.WriteString(input.Description)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## WHAT THE GOAL ALREADY HAS\n\n")
	sb.WriteString("- ikon: " + orNone(input.Icon) + "\n")
	sb.WriteString("- ska göras nu: " + yesNo(input.Now) + "\n")
	sb.WriteString("- evigt mål: " + yesNo(input.Eternal) + "\n")
	sb.WriteString("- agentmål: " + yesNo(input.Agent) + "\n")
	sb.WriteString("- slutdatum: " + orNone(input.Due) + "\n")
	sb.WriteString("- startdatum: " + orNone(input.StartDate) + "\n")
	sb.WriteString("- uppskattad tid: " + orNone(minutesText(input.EstimatedMinutes)) + "\n")
	sb.WriteString("- upprepningsregel: " + orNone(input.RecurrenceRule) + "\n")
	sb.WriteString("- taggar: " + orNone(strings.Join(input.Tags, ", ")) + "\n")
	sb.WriteString("- platser: " + orNone(strings.Join(input.Places, ", ")) + "\n")
	sb.WriteString("\n")

	if input.Feedback != "" {
		sb.WriteString("## PREVIOUS FEEDBACK (please address this)\n\n")
		sb.WriteString(input.Feedback)
		sb.WriteString("\n\n")
	}

	sb.WriteString("## YOUR TASK\n\n")
	sb.WriteString("Answer with one JSON object. Fill in every field you have a good reason to fill in, ")
	sb.WriteString("and leave the rest out. If the goal is unclear, make reasonable assumptions and state them in notes.\n")

	return sb.String()
}

// orNone skriver "ingen" när värdet saknas.
func orNone(value string) string {
	if strings.TrimSpace(value) == "" {
		return "ingen"
	}
	return value
}

// yesNo skriver ja eller nej.
func yesNo(value bool) string {
	if value {
		return "ja"
	}
	return "nej"
}

// minutesText skriver en uppskattning som text.
func minutesText(minutes int) string {
	if minutes <= 0 {
		return ""
	}
	return fmt.Sprintf("%d minuter", minutes)
}

// parseDefinition tolkar svaret från modellen.
//
// Svaret ska vara ett JSON-objekt. Om modellen svarar med något annat
// faller parsningen tillbaka på det äldre textformatet - en trasig
// definition ska inte stoppa agenten.
func parseDefinition(response string) (*Definition, error) {
	definition, err := parseJSONDefinition(response)
	if err == nil {
		return definition, nil
	}

	return parseTextDefinition(response)
}

// parseJSONDefinition läser ett JSON-objekt ur svaret.
//
// Modellen lägger ibland till en kodblok runt JSON:en, och ibland en
// inledande mening. Därför letar vi efter det första objektet i svaret
// i stället för att kräva att svaret bara är JSON.
func parseJSONDefinition(response string) (*Definition, error) {
	start := strings.Index(response, "{")
	if start < 0 {
		return nil, fmt.Errorf("svaret innehåller inget JSON-objekt")
	}

	object, ok := jsonObject(response, start)
	if !ok {
		return nil, fmt.Errorf("JSON-objektet i svaret är inte komplett")
	}

	var definition Definition
	if err := json.Unmarshal([]byte(object), &definition); err != nil {
		return nil, fmt.Errorf("kunde inte tolka definitionen: %w", err)
	}

	if strings.TrimSpace(definition.Title) == "" {
		return nil, fmt.Errorf("definitionen saknar en titel")
	}

	normalizeDefinition(&definition)
	return &definition, nil
}

// jsonObject hittar det JSON-objekt som börjar på positionen start.
//
// Strängar och escapes räknas med, så att ett "}" inne i en text inte
// avslutar objektet.
func jsonObject(text string, start int) (string, bool) {
	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(text); i++ {
		char := text[i]

		if inString {
			switch {
			case escaped:
				escaped = false
			case char == '\\':
				escaped = true
			case char == '"':
				inString = false
			}
			continue
		}

		switch char {
		case '"':
			inString = true
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return text[start : i+1], true
			}
		}
	}

	return "", false
}

// parseTextDefinition tolkar det äldre textformatet ("TITLE:" och
// "DESCRIPTION:").
func parseTextDefinition(response string) (*Definition, error) {
	lines := strings.Split(response, "\n")

	var title, description string
	inDescription := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "TITLE:") {
			title = strings.TrimSpace(strings.TrimPrefix(trimmed, "TITLE:"))
			continue
		}

		if strings.HasPrefix(trimmed, "DESCRIPTION:") {
			description = strings.TrimSpace(strings.TrimPrefix(trimmed, "DESCRIPTION:"))
			inDescription = true
			continue
		}

		if inDescription {
			// Check for another section header
			if strings.HasPrefix(trimmed, "TITLE:") || strings.HasPrefix(trimmed, "ANALYSIS:") {
				break
			}
			description += " " + trimmed
		}
	}

	// Fallback if structured parsing failed
	if title == "" {
		return nil, fmt.Errorf("svaret innehåller ingen titel")
	}

	definition := &Definition{
		Title:       title,
		Description: strings.TrimSpace(description),
	}
	normalizeDefinition(definition)
	return definition, nil
}

// normalizeDefinition puts värdena i den form de ska sparas i.
func normalizeDefinition(definition *Definition) {
	definition.Title = strings.TrimSpace(definition.Title)
	definition.Description = strings.TrimSpace(definition.Description)
	definition.Icon = strings.TrimSpace(definition.Icon)
	definition.Due = strings.TrimSpace(definition.Due)
	definition.StartDate = strings.TrimSpace(definition.StartDate)
	definition.RecurrenceRule = strings.TrimSpace(definition.RecurrenceRule)
	definition.RecurrenceMode = strings.TrimSpace(definition.RecurrenceMode)
	definition.RecurrenceStartDate = strings.TrimSpace(definition.RecurrenceStartDate)

	// Truncate title if too long
	if len(definition.Title) > maxTitleLength {
		definition.Title = definition.Title[:maxTitleLength-3] + "..."
	}
}
