// Package contracts läser kontraktet mellan Plan och agenten.
//
// Kontraktet ligger i src/contracts/plan-agent.yaml och beskriver de
// delar agenten använder: faser, tillstånd, triggrar, profiler och
// API-anropen. Både Plan API:t och agenten testar mot samma fil - om en
// sida ändras utan den andra faller testerna.
package contracts

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"gopkg.in/yaml.v3"
)

// Endpoint är ett API-anrop agenten gör mot Plan.
type Endpoint struct {
	Method string `yaml:"method"`
	Path   string `yaml:"path"`
	// Agent betyder att agenten själv gör anropet. Plan ska tillåta det
	// för agent-scope - annars får agenten 403 när den använder sin egen
	// nyckel.
	Agent bool `yaml:"agent"`
}

// Key gör att ett anrop kan jämföras i en mängd.
func (e Endpoint) Key() string {
	return e.Method + " " + e.Path
}

// Contract är kontraktet mellan Plan och agenten.
type Contract struct {
	Version            int      `yaml:"version"`
	Phases             []string `yaml:"phases"`
	Statuses           []string `yaml:"statuses"`
	TerminalStatuses   []string `yaml:"terminal_statuses"`
	StatusNeedingHuman []string `yaml:"status_needing_human"`
	Triggers           []string `yaml:"triggers"`
	Profiles           []string `yaml:"profiles"`
	// ApprovalKinds är typerna av godkännanden agenten kan skapa.
	// De deklareras här så att Plan och agenten pratar om samma sak.
	ApprovalKinds []string   `yaml:"approval_kinds"`
	Endpoints     []Endpoint `yaml:"endpoints"`
}

// Filnamnet på kontraktet.
const Filename = "plan-agent.yaml"

// Path returnerar sökvägen till kontraktet. Sökningen börjar i den
// katalog testet körs i och går uppåt tills filen hittas - så fungerar
// den oavsett vilken modul eller paket koden ligger i.
func Path() (string, error) {
	start, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("kan inte läsa arbetskatalogen: %w", err)
	}

	// Börja i katalogen där anropet sker, så att sökningen inte
	// behöver gissa om paketets djup.
	if _, file, _, ok := runtime.Caller(1); ok {
		start = filepath.Dir(file)
	}

	dir := start
	for {
		kandidat := filepath.Join(dir, "contracts", Filename)
		if _, err := os.Stat(kandidat); err == nil {
			return kandidat, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", fmt.Errorf("hittade inte kontraktet contracts/%s", Filename)
}

// Load läser kontraktet från disk.
func Load() (*Contract, error) {
	path, err := Path()
	if err != nil {
		return nil, err
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("kan inte läsa kontraktet: %w", err)
	}

	var contract Contract
	if err := yaml.Unmarshal(raw, &contract); err != nil {
		return nil, fmt.Errorf("kan inte tolka kontraktet: %w", err)
	}

	if contract.Version <= 0 {
		return nil, fmt.Errorf("kontraktet saknar en version")
	}

	return &contract, nil
}

// shared håller kontraktet som laddats en gång för processen.
//
// Kontraktet är samma fil för hela API:t - det behöver inte läsas om
// för varje anrop.
var (
	sharedOnce sync.Once
	shared     *Contract
	sharedErr  error
)

// Shared returnerar kontraktet, som laddas vid första anropet.
func Shared() (*Contract, error) {
	sharedOnce.Do(func() {
		shared, sharedErr = Load()
	})
	return shared, sharedErr
}

// HasEndpoint berättar om ett anrop finns i kontraktet.
func (c *Contract) HasEndpoint(method, path string) bool {
	for _, endpoint := range c.Endpoints {
		if endpoint.Method == method && endpoint.Path == path {
			return true
		}
	}
	return false
}

// AgentEndpoints returnerar de anrop agenten gör själv. De ska gå att göra
// med agent-scope.
func (c *Contract) AgentEndpoints() []Endpoint {
	var agentens []Endpoint
	for _, endpoint := range c.Endpoints {
		if endpoint.Agent {
			agentens = append(agentens, endpoint)
		}
	}
	return agentens
}

// HasApprovalKind berättar om en godkännandetyp finns i kontraktet.
//
// Ett godkännande utan typ godtas - så att förslag som skapats före
// godkännandetyperna introducerades fortfarande fungerar.
func (c *Contract) HasApprovalKind(kind string) bool {
	for _, kandidat := range c.ApprovalKinds {
		if kandidat == kind {
			return true
		}
	}
	return false
}

// NormalizePath gör om en formatsökväg (%d) till kontraktets form ({id}),
// så att koden och kontraktet kan jämföras.
func NormalizePath(path string) string {
	return strings.ReplaceAll(path, "%d", "{id}")
}
