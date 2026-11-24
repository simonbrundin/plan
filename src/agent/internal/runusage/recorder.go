// Package runusage samlar tokenanvändningen för en agentkörning.
//
// Agenten rapporterar bara användningen. Kostnaden räknas fram i Plan, ur
// en versionsstyrd prislista - så att samma prislista gäller för alla
// körningar och kostnaden kan reproduceras.
package runusage

import (
	"sync"
	"time"

	"plan-agent/internal/core/messages"
)

// Snapshot är den samlade användningen för en körning.
type Snapshot struct {
	// Provider och Model visar vilken modell som användes.
	Provider string `json:"provider"`
	Model    string `json:"model"`
	// ModelVersion är modellens version, om providern uppger den.
	ModelVersion string `json:"model_version,omitempty"`
	// Tokenantal för de olika sorterna.
	InputTokens     int `json:"input_tokens"`
	OutputTokens    int `json:"output_tokens"`
	CacheReadTokens int `json:"cache_read_tokens"`
	// ToolCalls är antalet verktygsanrop.
	ToolCalls int `json:"tool_calls"`
	// ToolNames är de verktyg som anropats.
	ToolNames []string `json:"tool_names,omitempty"`
	// StartedAt och FinishedAt avgränsar körningen.
	StartedAt  time.Time `json:"started_at"`
	FinishedAt time.Time `json:"finished_at"`
}

// Recorder samlar användningen för en körning.
//
// Den är trådsäker, eftersom agentens verktyg kan köras parallellt.
type Recorder struct {
	mu       sync.Mutex
	provider string
	model    string
	snapshot Snapshot
	tools    map[string]int
}

// New skapar en recorder för en provider och modell.
func New(provider, model string) *Recorder {
	return &Recorder{
		provider: provider,
		model:    model,
		tools:    map[string]int{},
		snapshot: Snapshot{
			Provider:  provider,
			Model:     model,
			StartedAt: time.Now(),
		},
	}
}

// Record användningen från ett modellsvar.
func (r *Recorder) Record(usage *messages.Usage) {
	if usage == nil {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.snapshot.InputTokens += usage.InputTokens
	r.snapshot.OutputTokens += usage.OutputTokens
	r.snapshot.CacheReadTokens += usage.CacheTokens
}

// RecordTool registrerar ett verktygsanrop.
func (r *Recorder) RecordTool(name string) {
	if name == "" {
		return
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.tools[name]++
	r.snapshot.ToolCalls++
}

// SetModelVersion sparar modellens version.
func (r *Recorder) SetModelVersion(version string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.snapshot.ModelVersion = version
}

// Snapshot returnerar den samlade användningen.
//
// Verktygsnamnen sorteras, så att samma körning alltid ger samma svar.
func (r *Recorder) Snapshot() Snapshot {
	r.mu.Lock()
	defer r.mu.Unlock()

	snapshot := r.snapshot
	snapshot.FinishedAt = time.Now()

	names := make([]string, 0, len(r.tools))
	for name := range r.tools {
		names = append(names, name)
	}
	sortStrings(names)
	snapshot.ToolNames = names

	return snapshot
}

// Reset tömmer insamlingen, så att samma recorder kan användas för nästa
// körning.
func (r *Recorder) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.snapshot = Snapshot{
		Provider:  r.provider,
		Model:     r.model,
		StartedAt: time.Now(),
	}
	r.tools = map[string]int{}
}

// sortStrings sorterar en lista på plats.
func sortStrings(values []string) {
	for i := 1; i < len(values); i++ {
		for j := i; j > 0 && values[j] < values[j-1]; j-- {
			values[j], values[j-1] = values[j-1], values[j]
		}
	}
}
