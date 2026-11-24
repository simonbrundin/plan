package coder

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"plan-agent/internal/executor"
	"plan-agent/internal/planclient"
	"plan-agent/internal/proposer"
	"plan-agent/internal/workspace"
)

// fakeProposer ger alltid samma förslag.
type fakeProposer struct {
	proposal *proposer.Proposal
	err      error
	calls    int
	// revised är förslaget som returneras när modellen blir ombedd att
	// rätta ett trasigt förslag.
	revised *proposer.Proposal
	// newTest är testkommandot som returneras när modellen blir ombedd att
	// skriva om ett trasigt test.
	newTest string
	// fixTests räknar antalet försök att skriva om testet.
	fixTests int
	// revisions räknar antalet revideringsförsök.
	revisions int
}

func (f *fakeProposer) Propose(context.Context, *workspace.Workspace, string, string) (*proposer.Proposal, error) {
	f.calls++
	if f.err != nil {
		return nil, f.err
	}
	return f.proposal, nil
}

// FixTest ger ett nytt testkommando, om det finns.
func (f *fakeProposer) FixTest(_ context.Context, _ *workspace.Workspace, _, _ string, _ *proposer.Proposal, _ string) (string, error) {
	f.fixTests++
	if f.newTest != "" {
		return f.newTest, nil
	}
	return "", fmt.Errorf("no new test command available")
}

// Revise ger det rättade förslaget, om det finns.
func (f *fakeProposer) Revise(_ context.Context, _ *workspace.Workspace, _, _ string, _ *proposer.Proposal, _ string) (*proposer.Proposal, error) {
	f.revisions++
	if f.revised != nil {
		return f.revised, nil
	}
	return nil, fmt.Errorf("no revision available")
}

// planServer är en miniskala Plan-API som svarar på de anrop codern gör.
type planServer struct {
	approvals   []planclient.AgentApproval
	created     []planclient.CreateAgentApprovalRequest
	events      []planclient.CreateAgentRunEventRequest
	updatedRuns []planclient.UpdateAgentRunRequest
}

// newPlanServer startar en testserver som beter sig som Plan.
func newPlanServer(t *testing.T, state *planServer) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: "define"},
				{ID: 2, Name: "plan"},
				{ID: 3, Name: "execute"},
				{ID: 4, Name: "verify"},
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/goals/42/status":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/agent-runs/7/approvals":
			_ = json.NewEncoder(w).Encode(state.approvals)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs/7/approvals":
			var req planclient.CreateAgentApprovalRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			state.created = append(state.created, req)
			_ = json.NewEncoder(w).Encode(planclient.AgentApproval{ID: 1, AgentRunID: 7, Status: "pending"})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs/7/events":
			var req planclient.CreateAgentRunEventRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			state.events = append(state.events, req)
			w.WriteHeader(http.StatusCreated)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/agent-runs/7":
			var req planclient.UpdateAgentRunRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			state.updatedRuns = append(state.updatedRuns, req)
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(server.Close)
	return server
}

// newTestRepo skapar ett litet git-repo att ändra.
func newTestRepo(t *testing.T) string {
	t.Helper()

	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "main.go"), "package main\n\nfunc main() {}\n")
	writeFile(t, filepath.Join(dir, "main_test.go"), "package main\n")

	runGit(t, dir, "init", "-q")
	runGit(t, dir, "add", "-A")
	runGit(t, dir, "-c", "user.email=agent@local", "-c", "user.name=agent", "commit", "-qm", "init")
	return dir
}

// runGit kör ett git-kommando i katalogen.
func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()

	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v failed: %v: %s", args, err, strings.TrimSpace(string(output)))
	}
}

// TestProposeVäntarPåGodkännande kontrollerar att agenten föreslår en
// ändring och att körningen blockeras tills en människa svarar.
func TestProposeVäntarPåGodkännande(t *testing.T) {
	state := &planServer{}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	proposal := &proposer.Proposal{
		Title:       "Lägg till en funktion",
		Description: "Målet kräver en ny funktion.",
		TestCommand: "go test ./...",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n\nfunc Ny() {}\n",
			Reason:  "målet kräver det",
		}},
	}

	coder := New(client, &fakeProposer{proposal: proposal}, Config{RepoPath: newTestRepo(t)})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if len(state.created) != 1 {
		t.Fatalf("antal godkännanden = %d, vill ha 1", len(state.created))
	}
	if state.created[0].Title != proposal.Title {
		t.Errorf("titel = %q", state.created[0].Title)
	}
	if state.created[0].VerificationPlan != proposal.TestCommand {
		t.Errorf("verifieringsplan = %q", state.created[0].VerificationPlan)
	}

	// Förslaget ska sparas så att exakt det kan appliceras senare.
	if state.created[0].Payload == nil {
		t.Fatal("förslaget sparades inte i godkännandet")
	}
	if _, ok := state.created[0].Payload["changes"]; !ok {
		t.Error("ändringarna saknas i payload")
	}

	// Körningen ska blockeras.
	if len(state.updatedRuns) == 0 {
		t.Fatal("körningen uppdaterades inte")
	}
	last := state.updatedRuns[len(state.updatedRuns)-1]
	if last.Status == nil || *last.Status != "blocked" {
		t.Errorf("status = %v, vill ha blocked", last.Status)
	}
}

// TestExecuteAnvänderGodkäntFörslag kontrollerar att det är det godkända
// förslaget som appliceras - inte ett nytt.
func TestExecuteAnvänderGodkäntFörslag(t *testing.T) {
	repo := newTestRepo(t)

	// Godkännandet innehåller en specifik ändring.
	approved := map[string]any{
		"changes": []any{
			map[string]any{
				"path":    "main.go",
				"content": "package main\n\nfunc main() {}\n\nfunc Godkand() {}\n",
				"reason":  "godkänd",
			},
		},
		"test_command": "cat main.go",
	}

	state := &planServer{
		approvals: []planclient.AgentApproval{{
			ID: 1, AgentRunID: 7, Status: "approved", Payload: approved,
		}},
	}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	// Proposern ger ett HELT ANNAT förslag - det får inte användas.
	other := &proposer.Proposal{
		Title:       "Något annat",
		TestCommand: "cat main.go",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n\nfunc Oskar(){}\n",
		}},
	}

	coder := New(client, &fakeProposer{proposal: other}, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	// En diff-händelse ska ha skrivits med det godkända innehållet.
	found := false
	for _, event := range state.events {
		if event.Type != planclient.EventArtifact {
			continue
		}
		diff, _ := event.Payload["diff"].(string)
		if contains(diff, "Godkand") {
			found = true
		}
		if contains(diff, "Oskar") {
			t.Error("det ogodkända förslaget applicerades")
		}
	}
	if !found {
		t.Error("den godkända ändringen finns inte i diffen")
	}
}

// TestExecuteVerifierarMedTestkommando kontrollerar att testkommandot körs
// och att resultatet skrivs till historiken.
func TestExecuteVerifierarMedTestkommando(t *testing.T) {
	approved := map[string]any{
		"changes": []any{
			map[string]any{"path": "fil.txt", "content": "hej hej\n"},
		},
		"test_command": "cat fil.txt",
	}

	state := &planServer{
		approvals: []planclient.AgentApproval{{ID: 1, AgentRunID: 7, Status: "approved", Payload: approved}},
	}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	repo := newTestRepo(t)
	writeFile(t, filepath.Join(repo, "fil.txt"), "hej\n")
	runGit(t, repo, "add", "-A")
	runGit(t, repo, "-c", "user.email=agent@local", "-c", "user.name=agent", "commit", "-qm", "fil")

	coder := New(client, &fakeProposer{}, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Ändra fil"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	found := false
	for _, event := range state.events {
		if event.Type != planclient.EventVerification {
			continue
		}
		if passed, ok := event.Payload["tests_passed"].(bool); ok && passed {
			found = true
		}
	}
	if !found {
		t.Error("ingen verifieringshändelse med godkänt resultat")
	}
}

// TestExecuteMisslyckatTestMarkerarKörningen kontrollerar att ett misslyckat
// test gör att körningen markeras som misslyckad.
func TestExecuteMisslyckatTestMarkerarKörningen(t *testing.T) {
	approved := map[string]any{
		"changes":      []any{map[string]any{"path": "fil.txt", "content": "hej hej\n"}},
		"test_command": "cat fil-saknas.txt",
	}

	state := &planServer{
		approvals: []planclient.AgentApproval{{ID: 1, AgentRunID: 7, Status: "approved", Payload: approved}},
	}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	repo := newTestRepo(t)
	writeFile(t, filepath.Join(repo, "fil.txt"), "hej\n")
	runGit(t, repo, "add", "-A")
	runGit(t, repo, "-c", "user.email=agent@local", "-c", "user.name=agent", "commit", "-qm", "fil")

	coder := New(client, &fakeProposer{}, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Ändra fil"})
	if err == nil {
		t.Fatal("ett misslyckat test ska ge ett fel")
	}

	failed := false
	for _, update := range state.updatedRuns {
		if update.Status != nil && *update.Status == "failed" {
			failed = true
		}
	}
	if !failed {
		t.Error("körningen markerades inte som misslyckad")
	}
}

// TestChangesFromApproval kontrollerar att förslaget läses tillbaka korrekt.
func TestChangesFromApproval(t *testing.T) {
	approval := &planclient.AgentApproval{
		Payload: map[string]any{
			"changes": []any{
				map[string]any{"path": "a.go", "content": "package a", "reason": "test"},
			},
			"test_command": "go test ./...",
		},
	}

	changes, testCommand, err := changesFromApproval(approval)
	if err != nil {
		t.Fatalf("changesFromApproval: %v", err)
	}
	if len(changes) != 1 {
		t.Fatalf("antal ändringar = %d, vill ha 1", len(changes))
	}
	if changes[0].Path != "a.go" {
		t.Errorf("path = %q", changes[0].Path)
	}
	if changes[0].Content != "package a" {
		t.Errorf("content = %q", changes[0].Content)
	}
	if changes[0].Reason != "test" {
		t.Errorf("reason = %q", changes[0].Reason)
	}
	if testCommand != "go test ./..." {
		t.Errorf("test command = %q", testCommand)
	}
}

// TestChangesFromApprovalSaknarInnehåll kontrollerar att ett godkännande
// utan förslag ger ett tydligt fel.
func TestChangesFromApprovalSaknarInnehåll(t *testing.T) {
	if _, _, err := changesFromApproval(&planclient.AgentApproval{}); err == nil {
		t.Error("ett tomt godkännande ska ge ett fel")
	}
	if _, _, err := changesFromApproval(nil); err == nil {
		t.Error("ett nil-godkännande ska ge ett fel")
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) > 0 && len(needle) > 0 && (haystack == needle ||
		len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0)
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("write: %v", err)
	}
}

// TestVäntandeFörslagGerIngetNytt kontrollerar att agenten inte föreslår
// samma sak igen medan en människa redan håller på att svara.
func TestVäntandeFörslagGerIngetNytt(t *testing.T) {
	state := &planServer{
		approvals: []planclient.AgentApproval{{ID: 1, AgentRunID: 7, Status: "pending"}},
	}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	coder := New(client, &fakeProposer{}, Config{RepoPath: newTestRepo(t)})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if len(state.created) != 0 {
		t.Errorf("antal nya godkännanden = %d, vill ha 0", len(state.created))
	}
}

// TestAvslagetFörslagGerIngetNytt kontrollerar att ett avslaget förslag
// inte körs om automatiskt.
func TestAvslagetFörslagGerIngetNytt(t *testing.T) {
	state := &planServer{
		approvals: []planclient.AgentApproval{{ID: 1, AgentRunID: 7, Status: "rejected"}},
	}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	coder := New(client, &fakeProposer{}, Config{RepoPath: newTestRepo(t)})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if len(state.created) != 0 {
		t.Errorf("antal nya godkännanden = %d, vill ha 0", len(state.created))
	}
}

// TestExecuteLäggerÄndringenPåBranch kontrollerar att den godkända
// ändringen hamnar på en egen branch och att huvudgrenen är orörd.
func TestExecuteLäggerÄndringenPåBranch(t *testing.T) {
	repo := newTestRepo(t)

	approved := map[string]any{
		"changes": []any{
			map[string]any{"path": "main.go", "content": "package main\n\nfunc main() {}\n\nfunc Extra() {}\n"},
		},
		"test_command": "cat main.go",
	}

	state := &planServer{
		approvals: []planclient.AgentApproval{{ID: 1, AgentRunID: 7, Status: "approved", Payload: approved}},
	}
	server := newPlanServer(t, state)

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	coder := New(client, &fakeProposer{}, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till Extra"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	// Branchen ska finnas i källrepoet.
	out, err := exec.Command("git", "-C", repo, "branch", "--list", "agent/goal-42-run-7").CombinedOutput()
	if err != nil {
		t.Fatalf("git branch: %v", err)
	}
	if !strings.Contains(string(out), "agent/goal-42-run-7") {
		t.Errorf("branchen finns inte: %s", strings.TrimSpace(string(out)))
	}

	// Huvudgrenen ska vara orörd.
	head, err := exec.Command("git", "-C", repo, "show", "HEAD:main.go").CombinedOutput()
	if err != nil {
		t.Fatalf("git show: %v", err)
	}
	if strings.Contains(string(head), "Extra") {
		t.Error("huvudgrenen ändrades - den ska lämnas orörd")
	}
}

// TestProposeRättarTrasigtFörslag kontrollerar att ett förslag vars eget
// testkommando misslyckas rättas innan det skickas till människan.
//
// Annars skulle människan läsa ett förslag som inte håller.
func TestProposeRättarTrasigtFörslag(t *testing.T) {
	repo := newTestRepo(t)

	broken := &proposer.Proposal{
		Title:       "Lägg till en fil",
		Description: "en fil",
		TestCommand: "grep -q 'hej' main.go",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n",
		}},
	}
	fixed := &proposer.Proposal{
		Title:       "Lägg till en fil",
		Description: "en fil",
		TestCommand: "grep -q 'package main' main.go",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n",
		}},
	}

	state := &planServer{}
	server := newPlanServer(t, state)
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	fake := &fakeProposer{proposal: broken, revised: fixed}
	coder := New(client, fake, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if fake.revisions != 1 {
		t.Errorf("revideringar = %d, vill ha 1", fake.revisions)
	}
	if len(state.created) != 1 {
		t.Fatalf("godkännanden = %d, vill ha 1", len(state.created))
	}
	if state.created[0].VerificationPlan != fixed.TestCommand {
		t.Errorf("testkommando = %q, vill ha det rättade", state.created[0].VerificationPlan)
	}
}

// TestProposeGerVarningNärFörslagetInteGårAttRätta kontrollerar att människan
// får veta att förslagets egen verifiering misslyckades.
func TestProposeGerVarningNärFörslagetInteGårAttRätta(t *testing.T) {
	repo := newTestRepo(t)

	broken := &proposer.Proposal{
		Title:       "Lägg till en fil",
		Description: "en fil",
		TestCommand: "grep -q 'hej' main.go",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n",
		}},
	}

	state2 := &planServer{}
	server2 := newPlanServer(t, state2)
	defer server2.Close()

	t.Setenv("PLAN_API_URL", server2.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client2, err2 := planclient.New()
	if err2 != nil {
		t.Fatalf("create Plan API client: %v", err2)
	}

	coder := New(client2, &fakeProposer{proposal: broken}, Config{RepoPath: repo})

	err := coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if len(state2.created) != 1 {
		t.Fatalf("godkännanden = %d, vill ha 1", len(state2.created))
	}
	if !strings.Contains(state2.created[0].Description, "gick inte igenom") {
		t.Errorf("beskrivning = %q, vill ha en varning", state2.created[0].Description)
	}
}

// TestProposeFixarTrasigtTestKommando kontrollerar att ett förslag med
// trasigt test får ett nytt testkommando utan att innehållet ändras.
//
// Annars skulle människan få ett förslag vars egen verifiering inte håller.
func TestProposeFixarTrasigtTestKommando(t *testing.T) {
	repo := newTestRepo(t)

	broken := &proposer.Proposal{
		Title:       "Lägg till en fil",
		Description: "en fil",
		TestCommand: "grep -q 'hej' main.go",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n",
		}},
	}

	state := &planServer{}
	server := newPlanServer(t, state)
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	fake := &fakeProposer{proposal: broken, newTest: "grep -q 'package main' main.go"}
	coder := New(client, fake, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if fake.fixTests != 1 {
		t.Errorf("testfixar = %d, vill ha 1", fake.fixTests)
	}
	if fake.revisions != 0 {
		t.Errorf("revideringar = %d, vill ha 0 - innehållet ska vara orört", fake.revisions)
	}
	if len(state.created) != 1 {
		t.Fatalf("godkännanden = %d, vill ha 1", len(state.created))
	}
	if state.created[0].VerificationPlan != "grep -q 'package main' main.go" {
		t.Errorf("testkommando = %q", state.created[0].VerificationPlan)
	}
}

// TestLivsprofilenFårInteÄndraKod kontrollerar att en körning med
// livsprofil inte får göra kodändringar.
//
// Profilen är en teknisk gräns - modellen ska inte kunna övertala den.
func TestLivsprofilenFårInteÄndraKod(t *testing.T) {
	repo := newTestRepo(t)

	proposal := &proposer.Proposal{
		Title:       "Lägg till en funktion",
		Description: "Målet kräver en ny funktion.",
		TestCommand: "go test ./...",
		Changes: []executor.Change{{
			Path:    "main.go",
			Content: "package main\n\nfunc main() {}\n\nfunc Ny() {}\n",
		}},
	}

	state := &planServer{}
	server := newPlanServer(t, state)
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	fake := &fakeProposer{proposal: proposal}
	coder := New(client, fake, Config{RepoPath: repo})

	err = coder.Process(context.Background(),
		planclient.AgentRun{ID: 7, GoalID: 42, AgentProfile: "life"},
		planclient.Goal{ID: 42, Title: "Lägg till funktion"})
	if err != nil {
		t.Fatalf("Process: %v", err)
	}

	if fake.calls != 0 {
		t.Error("livsprofilen ska inte få föreslå kodändringar")
	}
	if len(state.created) != 0 {
		t.Errorf("godkännanden = %d, vill ha 0", len(state.created))
	}
}
