package definer

import (
	"strings"
	"testing"
)

// TestParseDefinitionLäserJSON kontrollerar att ett JSON-svar tolkas.
func TestParseDefinitionLäserJSON(t *testing.T) {
	response := `{
		"title": "Bygg en rapporteringsvy",
		"description": "Vyn sammanfattar månadens siffror.",
		"icon": "heroicons:chart-bar",
		"now": true,
		"due": "2026-12-01",
		"estimated_minutes": 240,
		"recurrence_rule": "FREQ=WEEKLY;BYDAY=MO",
		"tags": ["rapportering"],
		"places": ["Hem"],
		"notes": ["Antog att siffrorna finns i en tabell"]
	}`

	definition, err := parseDefinition(response)
	if err != nil {
		t.Fatalf("parseDefinition: %v", err)
	}

	if definition.Title != "Bygg en rapporteringsvy" {
		t.Errorf("titel = %q", definition.Title)
	}
	if definition.Description != "Vyn sammanfattar månadens siffror." {
		t.Errorf("beskrivning = %q", definition.Description)
	}
	if definition.Icon != "heroicons:chart-bar" {
		t.Errorf("ikon = %q", definition.Icon)
	}
	if !definition.Now {
		t.Error("now ska vara sant")
	}
	if definition.Due != "2026-12-01" {
		t.Errorf("due = %q", definition.Due)
	}
	if definition.EstimatedMinutes != 240 {
		t.Errorf("uppskattning = %d", definition.EstimatedMinutes)
	}
	if definition.RecurrenceRule != "FREQ=WEEKLY;BYDAY=MO" {
		t.Errorf("upprepningsregel = %q", definition.RecurrenceRule)
	}
	if len(definition.Tags) != 1 || definition.Tags[0] != "rapportering" {
		t.Errorf("taggar = %v", definition.Tags)
	}
	if len(definition.Places) != 1 || definition.Places[0] != "Hem" {
		t.Errorf("platser = %v", definition.Places)
	}
	if len(definition.Notes) != 1 {
		t.Errorf("noteringar = %v", definition.Notes)
	}
}

// TestParseDefinitionHanterarFälder kontrollerar att modellen får skicka
// med fält den inte vet något om.
func TestParseDefinitionHanterarFälder(t *testing.T) {
	response := `{"title": "Ett mål", "description": "Beskrivning", "start_date": null, "tags": []}`

	definition, err := parseDefinition(response)
	if err != nil {
		t.Fatalf("parseDefinition: %v", err)
	}

	if definition.StartDate != "" {
		t.Errorf("startdatum = %q, vill ha tomt", definition.StartDate)
	}
	if len(definition.Tags) != 0 {
		t.Errorf("taggar = %v, vill ha tom lista", definition.Tags)
	}
}

// TestParseDefinitionHanterarKodblock kontrollerar att en JSON i ett
// kodblock tolkas - modellen gör det ibland.
func TestParseDefinitionHanterarKodblock(t *testing.T) {
	response := "Självklart! Här är definitionen:\n```json\n{\"title\": \"Ett mål\", \"description\": \"Beskrivning\"}\n```\n"

	definition, err := parseDefinition(response)
	if err != nil {
		t.Fatalf("parseDefinition: %v", err)
	}

	if definition.Title != "Ett mål" {
		t.Errorf("titel = %q", definition.Title)
	}
}

// TestParseDefinitionHanterarBracesITexten kontrollerar att ett "}"
// inne i en text inte avslutar JSON-objektet.
func TestParseDefinitionHanterarBracesITexten(t *testing.T) {
	response := `{"title": "Ett mål", "description": "Beskrivning med { och } i sig", "icon": "heroicons:flag"}`

	definition, err := parseDefinition(response)
	if err != nil {
		t.Fatalf("parseDefinition: %v", err)
	}

	if definition.Description != "Beskrivning med { och } i sig" {
		t.Errorf("beskrivning = %q", definition.Description)
	}
}

// TestParseDefinitionFallerTillbakaPåText kontrollerar att det gamla
// textformatet fortfarande tolkas.
func TestParseDefinitionFallerTillbakaPåText(t *testing.T) {
	response := "TITLE: Ett mål\n\nDESCRIPTION: En beskrivning\nmed flera rader."

	definition, err := parseDefinition(response)
	if err != nil {
		t.Fatalf("parseDefinition: %v", err)
	}

	if definition.Title != "Ett mål" {
		t.Errorf("titel = %q", definition.Title)
	}
	if definition.Description != "En beskrivning med flera rader." {
		t.Errorf("beskrivning = %q", definition.Description)
	}
}

// TestParseDefinitionAvvisarSkräp kontrollerar att ett svar utan vare
// sig JSON eller en titel inte blir en definition.
func TestParseDefinitionAvvisarSkräp(t *testing.T) {
	if _, err := parseDefinition("Jag kan inte hjälpa dig med det."); err == nil {
		t.Fatal("ett svar utan titel ska avvisas")
	}
}

// TestParseDefinitionKortarTitel kontrollerar att en för lång titel
// kortas av.
func TestParseDefinitionKortarTitel(t *testing.T) {
	response := `{"title": "` + strings.Repeat("a", 80) + `", "description": "Kort"}`

	definition, err := parseDefinition(response)
	if err != nil {
		t.Fatalf("parseDefinition: %v", err)
	}

	if len(definition.Title) > maxTitleLength {
		t.Errorf("titeln är %d tecken lång, vill ha max %d", len(definition.Title), maxTitleLength)
	}
}
