package contracts

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"plan-api/internal/models"
)

// TestKontraktetFinns kontrollerar att kontraktet går att läsa - utan det
// finns ingen sanningskälla att jämföra mot.
func TestKontraktetFinns(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	if contract.Version <= 0 {
		t.Errorf("kontraktet ska ha en version, fick %d", contract.Version)
	}

	if len(contract.Phases) == 0 || len(contract.Statuses) == 0 || len(contract.Endpoints) == 0 {
		t.Fatal("kontraktet ska innehålla faser, tillstånd och endpoints")
	}
}

// TestFaserMatcharModellen kontrollerar att Plan API:t och kontraktet
// använder samma faser.
func TestFaserMatcharModellen(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	if len(contract.Phases) != len(models.AgentRunPhases) {
		t.Fatalf("kontraktet har %d faser, modellen har %d", len(contract.Phases), len(models.AgentRunPhases))
	}

	for i, phase := range contract.Phases {
		if models.AgentRunPhases[i] != phase {
			t.Errorf("fas %d: kontraktet säger %q, modellen säger %q", i, phase, models.AgentRunPhases[i])
		}
	}
}

// TestStatusMatcharModellen kontrollerar att Plan API:t och kontraktet
// använder samma tillstånd.
func TestStatusMatcharModellen(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	if len(contract.Statuses) != len(models.AgentRunStatuses) {
		t.Fatalf("kontraktet har %d tillstånd, modellen har %d", len(contract.Statuses), len(models.AgentRunStatuses))
	}

	for i, status := range contract.Statuses {
		if models.AgentRunStatuses[i] != status {
			t.Errorf("tillstånd %d: kontraktet säger %q, modellen säger %q", i, status, models.AgentRunStatuses[i])
		}
	}
}

// TestTerminalaStatusMatchar kontrollerar att listan över terminala
// tillstånd stämmer med modellen.
func TestTerminalaStatusMatchar(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	for _, status := range contract.TerminalStatuses {
		if !models.IsTerminalAgentRunStatus(status) {
			t.Errorf("kontraktet säger att %q är terminalt, modellen säger motsatsen", status)
		}
	}

	for _, status := range models.AgentRunStatuses {
		terminal := models.IsTerminalAgentRunStatus(status)
		finns := false
		for _, kandidat := range contract.TerminalStatuses {
			if kandidat == status {
				finns = true
				break
			}
		}
		if terminal != finns {
			t.Errorf("tillståndet %q: terminalt i modellen = %t, i kontraktet = %t", status, terminal, finns)
		}
	}
}

// TestStatusSomKraverManskaMatchar kontrollerar att listan över tillstånd
// där en människa måste agera stämmer med modellen.
func TestStatusSomKraverManskaMatchar(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	for _, status := range contract.StatusNeedingHuman {
		if !models.AgentRunStatusNeedsHuman(status) {
			t.Errorf("kontraktet säger att %q kräver en människa, modellen säger motsatsen", status)
		}
	}
}

// TestOvergångarAnvanderKandaStatus kontrollerar att övergångstabellen
// bara pekar på tillstånd som finns.
func TestOvergångarAnvanderKandaStatus(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	for från, till := range models.AgentRunStatusTransitions() {
		if !models.ValidAgentRunStatus(från) {
			t.Errorf("övergången startar i okänt tillstånd %q", från)
		}
		for _, status := range till {
			if !models.ValidAgentRunStatus(status) {
				t.Errorf("övergången %q -> %q pekar på ett okänt tillstånd", från, status)
			}
		}
	}

	// Varje icke-terminalt tillstånd måste kunna avbrytas.
	for _, status := range contract.Statuses {
		if models.IsTerminalAgentRunStatus(status) {
			continue
		}
		if !models.IsTerminalAgentRunStatus(status) && !contains(contract.Statuses, status) {
			t.Errorf("tillståndet %q finns inte i kontraktet", status)
		}
	}
}

// TestEndpointFinnsISwagger kontrollerar att varje anrop i kontraktet finns
// i API:ets OpenAPI-specifikation - annars har agenten något att anropa som
// inte finns.
func TestEndpointFinnsISwagger(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	spec := läsSwagger(t)

	for _, endpoint := range contract.Endpoints {
		path, ok := spec["paths"].(map[string]any)
		if !ok {
			t.Fatal("swagger-specifikationen saknar paths")
		}

		operationer, ok := path[endpoint.Path].(map[string]any)
		if !ok {
			t.Errorf("kontraktet använder %s %s, men den vägen finns inte i swagger", endpoint.Method, endpoint.Path)
			continue
		}

		if _, ok := operationer[strings.ToLower(endpoint.Method)]; !ok {
			t.Errorf("swagger saknar %s %s", endpoint.Method, endpoint.Path)
		}
	}
}

// TestAgentEndpointsIkontraktet kontrollerar åt andra hållet: varje
// agent-endpoint i swagger ska finnas i kontraktet.
func TestAgentEndpointsIkontraktet(t *testing.T) {
	contract, err := Load()
	if err != nil {
		t.Fatalf("kontraktet ska gå att läsa: %v", err)
	}

	spec := läsSwagger(t)
	paths, ok := spec["paths"].(map[string]any)
	if !ok {
		t.Fatal("swagger-specifikationen saknar paths")
	}

	for sökväg, operationer := range paths {
		if !strings.HasPrefix(sökväg, "/api/v1/agent-run") {
			continue
		}

		for metod := range operationer.(map[string]any) {
			if !contract.HasEndpoint(strings.ToUpper(metod), sökväg) {
				t.Errorf("swagger har %s %s, men det finns inte i kontraktet", strings.ToUpper(metod), sökväg)
			}
		}
	}
}

func läsSwagger(t *testing.T) map[string]any {
	t.Helper()

	path := filepath.Join("..", "..", "docs", "swagger.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("kan inte läsa swagger-specifikationen: %v", err)
	}

	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		t.Fatalf("kan inte tolka swagger-specifikationen: %v", err)
	}
	return spec
}

func contains(list []string, value string) bool {
	for _, item := range list {
		if item == value {
			return true
		}
	}
	return false
}
