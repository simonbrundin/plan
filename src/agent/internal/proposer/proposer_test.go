package proposer

import (
	"path/filepath"
	"testing"

	"plan-agent/internal/executor"
)

// TestParseProposal kontrollerar att ett välformaterat svar blir ett förslag.
func TestParseProposal(t *testing.T) {
	response := `TITLE: Lägg till test för agentprofil
DESCRIPTION: Testet kontrollerar att agent_profile sparas korrekt.
TEST_COMMAND: go test ./internal/models/...

FILE: internal/models/agent_run_profile_test.go
CONTENT:
package models

func TestAgentProfile(t *testing.T) {
	if !ValidAgentProfile(AgentProfileCoding) {
		t.Error("coding ska vara giltig")
	}
}
END_FILE

FILE: internal/models/agent_run.go
CONTENT:
package models
END_FILE
`

	proposal, err := parseProposal(response)
	if err != nil {
		t.Fatalf("parseProposal: %v", err)
	}

	if proposal.Title != "Lägg till test för agentprofil" {
		t.Errorf("title = %q", proposal.Title)
	}
	if proposal.Description != "Testet kontrollerar att agent_profile sparas korrekt." {
		t.Errorf("description = %q", proposal.Description)
	}
	if proposal.TestCommand != "go test ./internal/models/..." {
		t.Errorf("test command = %q", proposal.TestCommand)
	}
	if len(proposal.Changes) != 2 {
		t.Fatalf("antal ändringar = %d, vill ha 2", len(proposal.Changes))
	}

	first := proposal.Changes[0]
	if first.Path != "internal/models/agent_run_profile_test.go" {
		t.Errorf("path = %q", first.Path)
	}
	if first.Content != "package models\n\nfunc TestAgentProfile(t *testing.T) {\n\tif !ValidAgentProfile(AgentProfileCoding) {\n\t\tt.Error(\"coding ska vara giltig\")\n\t}\n}" {
		t.Errorf("content = %q", first.Content)
	}
}

// TestParseProposalSaknarÄndringar kontrollerar att ett svar utan
// filändringar ger ett fel.
func TestParseProposalSaknarÄndringar(t *testing.T) {
	response := `TITLE: Ingenting
DESCRIPTION: Ingen kod
TEST_COMMAND: go test ./...
`

	if _, err := parseProposal(response); err == nil {
		t.Error("ett svar utan ändringar ska ge ett fel")
	}
}

// TestParseProposalUtanHeader kontrollerar att ett förslag utan titel får
// ett standardvärde i stället för att fallera.
func TestParseProposalUtanHeader(t *testing.T) {
	response := `FILE: a.go
CONTENT:
package a
END_FILE
`

	proposal, err := parseProposal(response)
	if err != nil {
		t.Fatalf("parseProposal: %v", err)
	}
	if proposal.Title == "" {
		t.Error("titel ska ha ett standardvärde")
	}
	if len(proposal.Changes) != 1 {
		t.Errorf("antal ändringar = %d, vill ha 1", len(proposal.Changes))
	}
}

// TestIsRelevantSource kontrollerar vilka filer som läses in.
func TestIsRelevantSource(t *testing.T) {
	relevant := []string{"main.go", "agent_run_test.go", "app.vue", "types.ts"}
	for _, name := range relevant {
		if !isRelevantSource(name) {
			t.Errorf("filen %q ska vara relevant", name)
		}
	}

	irrelevant := []string{"Makefile", "logo.png"}
	for _, name := range irrelevant {
		if isRelevantSource(name) {
			t.Errorf("filen %q ska inte vara relevant", name)
		}
	}
}

// TestProposalÄndringarÄrExecutorChange kontrollerar att förslagets ändringar
// kan användas av exekveraren.
func TestProposalÄndringarÄrExecutorChange(t *testing.T) {
	proposal := &Proposal{
		Changes: []executor.Change{{Path: "a.go", Content: "package a"}},
	}

	if proposal.Changes[0].Path != "a.go" {
		t.Error("ändringen överfördes inte")
	}
}

// TestRelativePathGörOmAbsolutaSökvägar kontrollerar att sökvägar i ett
// förslag pekar på filer i repot och inte på en temporär arbetskatalog.
func TestRelativePathGörOmAbsolutaSökvägar(t *testing.T) {
	root := "/tmp/agent-workspace-123"
	proposal := &Proposal{
		Changes: []executor.Change{
			{Path: filepath.Join(root, "calculator.go")},
			{Path: "calculator_test.go"},
		},
	}

	relativize(root, proposal)

	if proposal.Changes[0].Path != "calculator.go" {
		t.Errorf("sökväg = %q, vill ha calculator.go", proposal.Changes[0].Path)
	}
	if proposal.Changes[1].Path != "calculator_test.go" {
		t.Errorf("sökväg = %q", proposal.Changes[1].Path)
	}
}

// TestRelativePathBehållerUtanförRepo kontrollerar att en sökväg utanför
// repot inte görs om till en relativ sökväg.
func TestRelativePathBehållerUtanförRepo(t *testing.T) {
	root := "/tmp/agent-workspace-123"
	proposal := &Proposal{
		Changes: []executor.Change{{Path: "/etc/passwd"}},
	}

	relativize(root, proposal)

	if proposal.Changes[0].Path != "/etc/passwd" {
		t.Errorf("sökväg = %q, vill ha /etc/passwd", proposal.Changes[0].Path)
	}
}

// TestRelativePathGörOmAbsolutaSökvägar kontrollerar att sökvägar i ett
// förslag pekar på filer i repot och inte på en temporär arbetskatalog.
// skriva om dem från grunden.
func TestIsRelevantSourceTarMedDokumentation(t *testing.T) {
	relevant := []string{
		"main.go", "agent_run.go", "app.vue", "useAgentApi.ts",
		"README.md", "AGENTS.md", "deployment.yaml", "clone.sh",
	}
	for _, name := range relevant {
		if !isRelevantSource(name) {
			t.Errorf("%s borde vara relevant", name)
		}
	}

	irrelevant := []string{"go.sum", "package-lock.json", "image.png"}
	for _, name := range irrelevant {
		if isRelevantSource(name) {
			t.Errorf("%s borde inte vara relevant", name)
		}
	}
}

// TestParseProposalLäserHeredocTestkommando kontrollerar att ett
// flerradigt testkommando läses i sin helhet.
//
// Utan detta blir testkommandot en trunkad rad som inte kan köras.
func TestParseProposalLäserHeredocTestkommando(t *testing.T) {
	response := `TITLE: Lägg till en sektion
DESCRIPTION: En kort sektion
TEST_COMMAND: python3 - <<'PY'
print("kontrollerar innehållet")
PY

FILE: README.md
CONTENT:
# Plan
END_FILE`

	proposal, err := parseProposal(response)
	if err != nil {
		t.Fatalf("parseProposal: %v", err)
	}

	expected := "python3 - <<'PY'\nprint(\"kontrollerar innehållet\")\nPY"
	if proposal.TestCommand != expected {
		t.Errorf("testkommando = %q, vill ha %q", proposal.TestCommand, expected)
	}
}
