package agentworker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"plan-agent/internal/definer"
	"plan-agent/internal/planclient"
)

type fakeDefiner struct{}

func (fakeDefiner) Define(context.Context, string, string, string) (*definer.Definition, error) {
	return &definer.Definition{
		Title:       "Ett definierat mål",
		Description: "Målet har tydliga krav och ett verifierbart slutkriterium.",
	}, nil
}

// TestRunOnceDefinesGoalEndToEnd verifierar fas 1:s första hela del-flöde:
// agentmål -> AgentRun -> definition -> Plan-status plan -> klar körning.
// Testet använder samma HTTP-kontrakt som den riktiga Plan-klienten men en
// deterministisk definerare, så det kräver varken en LLM-nyckel eller nätverk.
func TestRunOnceDefinesGoalEndToEnd(t *testing.T) {
	var createdRun bool
	var updatedGoal bool
	var updatedRun bool
	var updatedPhase bool
	var createRequest planclient.CreateAgentRunRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Gör något viktigt", Description: "lös idé", StatusID: 1, Agent: true,
			}})
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{})
		case "POST /api/v1/agent-runs":
			createdRun = true
			if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
				t.Errorf("decode agent run create: %v", err)
			}
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7, GoalID: 42, Phase: PhaseDefine, Status: StatusRunning})
		case "PATCH /api/v1/goals/42":
			updatedGoal = true
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42, StatusID: 1}})
		case "PATCH /api/v1/goals/42/status":
			updatedPhase = true
			_ = json.NewEncoder(w).Encode(map[string]any{"status_id": 2})
		case "PATCH /api/v1/agent-runs/7":
			var request planclient.UpdateAgentRunRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode agent run update: %v", err)
			}
			if request.Status != nil {
				updatedRun = *request.Status == StatusDone
			}
			if request.Phase != nil {
				updatedPhase = updatedPhase && *request.Phase == PhasePlan
			}
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if !createdRun || !updatedGoal || !updatedRun || !updatedPhase {
		t.Fatalf("flow incomplete: createdRun=%v updatedGoal=%v updatedRun=%v updatedPhase=%v", createdRun, updatedGoal, updatedRun, updatedPhase)
	}
}

// TestRunOnceRegistrerarTriggerOchInitiativtagare verifierar att workern
// markerar körningen som agentstartad och vem som startade den.
func TestRunOnceRegistrerarTriggerOchInitiativtagare(t *testing.T) {
	var createRequest planclient.CreateAgentRunRequest

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Gör något viktigt", Description: "lös idé", StatusID: 1, Agent: true,
			}})
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{})
		case "POST /api/v1/agent-runs":
			if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
				t.Errorf("decode agent run create: %v", err)
			}
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7, GoalID: 42})
		case "PATCH /api/v1/goals/42":
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42, StatusID: 1}})
		case "PATCH /api/v1/goals/42/status":
			_ = json.NewEncoder(w).Encode(map[string]any{"status_id": 2})
		case "PATCH /api/v1/agent-runs/7":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{
		OnlyGoalId: 42,
		Initiator:  "worker:agent-worker-goal-42",
	})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if createRequest.Trigger != TriggerAgent {
		t.Errorf("trigger = %q, vill ha %q", createRequest.Trigger, TriggerAgent)
	}
	if createRequest.Initiator != "worker:agent-worker-goal-42" {
		t.Errorf("initiator = %q, vill ha %q", createRequest.Initiator, "worker:agent-worker-goal-42")
	}
}

// TestRunOnceÅterupptarBlockeradKörning verifierar att workern fortsätter
// en körning som väntat på ett mänskligt godkännande.
func TestRunOnceÅterupptarBlockeradKörning(t *testing.T) {
	var processedGoal int64
	var processedRun int64

	coder := &fakeCodeRunner{
		process: func(_ context.Context, run planclient.AgentRun, goal planclient.Goal) error {
			processedGoal = goal.ID
			processedRun = run.ID
			return nil
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Ändra något", Description: "beskrivning", StatusID: 2, Agent: true,
			}})
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 7, GoalID: 42, Phase: PhasePlan, Status: StatusBlocked,
			}})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42, CodeRunner: coder})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if processedGoal != 42 {
		t.Errorf("mål = %d, vill ha 42", processedGoal)
	}
	if processedRun != 7 {
		t.Errorf("körning = %d, vill ha 7", processedRun)
	}
}

// fakeCodeRunner gör ingenting - den bara registrerar att den anropades.
type fakeCodeRunner struct {
	process func(context.Context, planclient.AgentRun, planclient.Goal) error
}

func (f *fakeCodeRunner) Process(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) error {
	if f.process == nil {
		return nil
	}
	return f.process(ctx, run, goal)
}

// TestRunOnceTarÖverKöadKörning verifierar att en uttrycklig begäran om
// att köra ett mål tas över av workern.
func TestRunOnceTarÖverKöadKörning(t *testing.T) {
	var processedRun int64
	var processedGoal int64

	coder := &fakeCodeRunner{
		process: func(_ context.Context, run planclient.AgentRun, goal planclient.Goal) error {
			processedRun = run.ID
			processedGoal = goal.ID
			return nil
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Ändra något", Description: "beskrivning", StatusID: 2, Agent: true,
			}})
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 11, GoalID: 42, Phase: PhaseDefine, Status: StatusQueued,
			}})
		case "PATCH /api/v1/agent-runs/11":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11})
		case "POST /api/v1/agent-runs/11/events":
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42, CodeRunner: coder})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if processedGoal != 42 {
		t.Errorf("mål = %d, vill ha 42", processedGoal)
	}
	if processedRun != 11 {
		t.Errorf("körning = %d, vill ha 11", processedRun)
	}
}

// TestLivsprofilenAvslutarKörningenUtanÄndringar verifierar att en
// profil som inte får röra kod avslutar körningen i stället för att
// lämna den öppen.
func TestLivsprofilenAvslutarKörningenUtanÄndringar(t *testing.T) {
	var updated planclient.UpdateAgentRunRequest
	var events []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Livsprofil", Description: "beskrivning", StatusID: 2, Agent: true,
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 11, GoalID: 42, Phase: PhasePlan, Status: StatusQueued, AgentProfile: "life",
			}})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/agent-runs/11":
			_ = json.NewDecoder(r.Body).Decode(&updated)
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs/11/events":
			var event planclient.CreateAgentRunEventRequest
			_ = json.NewDecoder(r.Body).Decode(&event)
			events = append(events, event.Type)
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42, Profile: "life"})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if updated.Status == nil || *updated.Status != StatusDone {
		t.Errorf("körningens status = %v, vill ha %s", updated.Status, StatusDone)
	}

	// Historiken ska visa att körningen startade och avslutades.
	if len(events) < 2 || events[0] != planclient.EventRunStarted {
		t.Errorf("händelser = %v, vill ha run_started först", events)
	}
	if events[len(events)-1] != planclient.EventRunCompleted {
		t.Errorf("händelser = %v, vill ha run_completed sist", events)
	}
}

// TestKöadKörningTasÖverÄvenFörMålSomInteÄrAgentmål verifierar att en
// uttrycklig körningsbegäran inte glöms bara för att målet inte är
// markerat som agentmål. Utan det skulle jobbet köra i evighet.
func TestKöadKörningTasÖverÄvenFörMålSomInteÄrAgentmål(t *testing.T) {
	var processedRun int64
	var processedGoal int64

	coder := &fakeCodeRunner{
		process: func(_ context.Context, run planclient.AgentRun, goal planclient.Goal) error {
			processedRun = run.ID
			processedGoal = goal.ID
			return nil
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/goals":
			// Agentmålen - målet 42 finns inte bland dem.
			_ = json.NewEncoder(w).Encode([]planclient.Goal{})
		case "GET /api/v1/goals/42":
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {
				ID: 42, Title: "Kör det här", Description: "beskrivning", StatusID: 2, Agent: false,
			}})
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 11, GoalID: 42, Phase: PhasePlan, Status: StatusQueued,
			}})
		case "PATCH /api/v1/agent-runs/11":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11})
		case "POST /api/v1/agent-runs/11/events":
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42, CodeRunner: coder})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if processedGoal != 42 {
		t.Errorf("mål = %d, vill ha 42", processedGoal)
	}
	if processedRun != 11 {
		t.Errorf("körning = %d, vill ha 11", processedRun)
	}
}

// TestDefinieradKörningSlutarMedRunCompleted verifierar att en körning som
// avslutas efter definitionen också registreras som avslutad. Annars ser
// historiken ut som om körningen bara upphörde.
func TestDefinieradKörningSlutarMedRunCompleted(t *testing.T) {
	var events []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Ett mål", Description: "beskrivning", StatusID: 1, Agent: true,
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11, GoalID: 42})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/goals/42":
			_ = json.NewEncoder(w).Encode(planclient.Goal{ID: 42})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/goals/42/status":
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/agent-runs/11":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs/11/events":
			var event planclient.CreateAgentRunEventRequest
			_ = json.NewDecoder(r.Body).Decode(&event)
			events = append(events, event.Type)
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	var completed bool
	for _, event := range events {
		if event == planclient.EventRunCompleted {
			completed = true
		}
	}
	if !completed {
		t.Errorf("händelser = %v, vill ha run_completed", events)
	}
}
