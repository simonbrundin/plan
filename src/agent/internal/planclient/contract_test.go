package planclient

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

// kontraktet är filen som beskriver vad agenten får använda.
type kontrakt struct {
	Version            int        `yaml:"version"`
	Phases             []string   `yaml:"phases"`
	Statuses           []string   `yaml:"statuses"`
	TerminalStatuses   []string   `yaml:"terminal_statuses"`
	StatusNeedingHuman []string   `yaml:"status_needing_human"`
	Triggers           []string   `yaml:"triggers"`
	Profiles           []string   `yaml:"profiles"`
	Endpoints          []endpoint `yaml:"endpoints"`
}

type endpoint struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	// Agent betyder att agenten gör anropet själv.
	Agent bool `yaml:"agent"`
}

// läsKontrakt läser kontraktet från src/contracts/plan-agent.yaml.
//
// Sökningen går uppåt från katalogen testet körs i tills filen hittas -
// så fungerar den oavsett vilken paketkatalog testet ligger i.
func läsKontrakt(t *testing.T) *kontrakt {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("kan inte läsa arbetskatalogen: %v", err)
	}

	var path string
	for {
		kandidat := filepath.Join(dir, "contracts", "plan-agent.yaml")
		if _, err := os.Stat(kandidat); err == nil {
			path = kandidat
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("hittade inte kontraktet contracts/plan-agent.yaml")
		}
		dir = parent
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("kan inte läsa kontraktet: %v", err)
	}

	var k kontrakt
	if err := yaml.Unmarshal(raw, &k); err != nil {
		t.Fatalf("kan inte tolka kontraktet: %v", err)
	}

	if k.Version <= 0 {
		t.Fatal("kontraktet ska ha en version")
	}
	return &k
}

// TestStatusMatcharKontraktet kontrollerar att agentens tillstånd är de
// som kontraktet anger - annars skulle agenten få avslag på sin egen
// statusuppdatering.
func TestStatusMatcharKontraktet(t *testing.T) {
	k := läsKontrakt(t)

	agentens := []string{
		RunStatusQueued,
		RunStatusRunning,
		RunStatusWaitingApproval,
		RunStatusSucceeded,
		RunStatusFailed,
		RunStatusCancelled,
	}

	if len(agentens) != len(k.Statuses) {
		t.Fatalf("kontraktet har %d tillstånd, agenten har %d", len(k.Statuses), len(agentens))
	}

	for i, status := range k.Statuses {
		if agentens[i] != status {
			t.Errorf("tillstånd %d: kontraktet säger %q, agenten säger %q", i, status, agentens[i])
		}
	}
}

// TestSokvagarMatcharKontraktet kontrollerar att varje sökväg agenten
// använder finns i kontraktet.
func TestSokvagarMatcharKontraktet(t *testing.T) {
	k := läsKontrakt(t)

	kända := map[string]bool{}
	for _, e := range k.Endpoints {
		kända[e.Method+" "+e.Path] = true
	}

	for _, sökväg := range AllaSökvägar {
		// Parametrar skrivs %d i koden och {id} i kontraktet.
		normaliserad := strings.ReplaceAll(sökväg, "%d", "{id}")

		found := false
		for _, e := range k.Endpoints {
			if e.Path == normaliserad {
				found = true
				break
			}
		}

		if !found {
			t.Errorf("agenten använder %s, men den vägen finns inte i kontraktet", normaliserad)
		}
	}
}

// TestAgentensAnropArMarkade kontrollerar att varje sökväg agenten
// använder är markerad som agentens eget anrop i kontraktet.
//
// Det markerade anropet är Plan:s agent-scope: agenten får bara göra de
// anrop den är markerad för. Ett anrop agenten gör men som inte är
// markerat ger 403 i prod, där agentens nyckel har scope 'agent'.
func TestAgentensAnropArMarkade(t *testing.T) {
	k := läsKontrakt(t)

	markerade := map[string]bool{}
	for _, e := range k.Endpoints {
		if e.Agent {
			markerade[e.Path] = true
		}
	}

	for _, sökväg := range AllaSökvägar {
		normaliserad := strings.ReplaceAll(sökväg, "%d", "{id}")

		if !markerade[normaliserad] {
			t.Errorf("agenten använder %s, men den vägen är inte markerad som agentens eget anrop", normaliserad)
		}
	}
}

// TestKontraktetsAnropFinnsIPlan kontrollerar att varje anrop i kontraktet
// finns i Plan API:ets OpenAPI-specifikation - annars står det i kontraktet
// något som inte finns.
func TestKontraktetsAnropFinnsIPlan(t *testing.T) {
	k := läsKontrakt(t)

	spec := läsSwagger(t)

	for _, e := range k.Endpoints {
		paths, ok := spec["paths"].(map[string]any)
		if !ok {
			t.Fatal("swagger-specifikationen saknar paths")
		}

		operationer, ok := paths[e.Path].(map[string]any)
		if !ok {
			t.Errorf("kontraktet använder %s %s, men den vägen finns inte i Plan", e.Method, e.Path)
			continue
		}

		if _, ok := operationer[strings.ToLower(e.Method)]; !ok {
			t.Errorf("Plan saknar %s %s", e.Method, e.Path)
		}
	}
}

// läsSwagger läser Plan API:ets OpenAPI-specifikation.
//
// Sökningen utgår från modulroten (katalogen med go.mod) - så hittas
// rätt fil oavsett vilken paketkatalog testet ligger i. Vi letar inte
// bara efter "api/docs/swagger.json" uppåt: agenten har en egen
// swagger.json i internal/api/docs som ligger närmare.
func läsSwagger(t *testing.T) map[string]any {
	t.Helper()

	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("kan inte läsa arbetskatalogen: %v", err)
	}

	var path string
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			path = filepath.Join(dir, "api", "docs", "swagger.json")
			break
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("hittade inte modulroten (go.mod)")
		}
		dir = parent
	}

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
