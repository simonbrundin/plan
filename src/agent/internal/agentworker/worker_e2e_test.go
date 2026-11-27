package agentworker

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"plan/agent/internal/definer"
	"plan/agent/internal/goaldefinition"
	"plan/agent/internal/planclient"
)

type fakeDefiner struct {
	definition *definer.Definition
}

func (f fakeDefiner) Define(context.Context, definer.Input) (*definer.Definition, error) {
	if f.definition != nil {
		return f.definition, nil
	}
	return &definer.Definition{
		Title:       "Ett definierat mål",
		Description: "Målet har tydliga krav och ett verifierbart slutkriterium.",
	}, nil
}

// TestRunOnceDefinesGoalEndToEnd verifierar fas 1:s första hela del-flöde:
// agentmål -> AgentRun -> definition -> godkännande -> väntan på människa.
// Testet använder samma HTTP-kontrakt som den riktiga Plan-klienten men en
// deterministisk definerare, så det kräver varken en LLM-nyckel eller
// nätverk.
//
// Målet skrivs inte till förrän en människa godkänt förslaget - det är
// human gate.
func TestRunOnceDefinesGoalEndToEnd(t *testing.T) {
	var createdRun bool
	var createdApproval bool
	var updatedGoal bool
	var waitingApproval bool
	var approvalPayload map[string]any
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
		case "GET /api/v1/places":
			_ = json.NewEncoder(w).Encode([]planclient.Place{})
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{})
		case "POST /api/v1/agent-runs":
			createdRun = true
			if err := json.NewDecoder(r.Body).Decode(&createRequest); err != nil {
				t.Errorf("decode agent run create: %v", err)
			}
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7, GoalID: 42, Phase: PhaseDefine, Status: StatusRunning})
		case "POST /api/v1/agent-runs/7/approvals":
			createdApproval = true
			var request planclient.CreateAgentApprovalRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode approval create: %v", err)
			}
			approvalPayload = request.Payload
			_ = json.NewEncoder(w).Encode(planclient.AgentApproval{ID: 3, AgentRunID: 7, GoalID: 42, Status: "pending"})
		case "PATCH /api/v1/goals/42":
			updatedGoal = true
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42, StatusID: 1}})
		case "PATCH /api/v1/goals/42/status":
			_ = json.NewEncoder(w).Encode(map[string]any{"status_id": 2})
		case "POST /api/v1/goals/42/phase":
			_ = json.NewEncoder(w).Encode(planclient.PhaseTransition{GoalID: 42, From: PhaseDefine, To: PhasePlan})
		case "PATCH /api/v1/agent-runs/7":
			var request planclient.UpdateAgentRunRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode agent run update: %v", err)
			}
			if request.Status != nil && *request.Status == StatusWaitingApproval {
				waitingApproval = true
			}
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		case "POST /api/v1/agent-runs/7/events":
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

	if !createdRun || !createdApproval || !waitingApproval {
		t.Fatalf("flow incomplete: createdRun=%v createdApproval=%v waitingApproval=%v", createdRun, createdApproval, waitingApproval)
	}
	if updatedGoal {
		t.Error("målet ska inte skrivas till förrän en människa godkänt förslaget")
	}
	if approvalPayload["kind"] != goaldefinition.Kind {
		t.Errorf("godkännandets typ = %v, vill ha %q", approvalPayload["kind"], goaldefinition.Kind)
	}
}

// TestGodkändDefinitionAppliceras verifierar att exakt de godkända
// värdena skrivs till målet när en människa godkänt förslaget.
func TestGodkändDefinitionAppliceras(t *testing.T) {
	var updatedGoal planclient.UpdateGoalRequest
	var updatedPhase bool
	var completed bool

	proposal := goaldefinition.Build(goaldefinition.Context{
		Goal: planclient.Goal{ID: 42, Title: "Gör något viktigt", Description: "lös idé", StatusID: 1, Agent: true},
		Definition: definer.Definition{
			Title:       "Ett definierat mål",
			Description: "Målet har tydliga krav.",
			Icon:        "heroicons:flag",
		},
		NextPhase: PhasePlan,
	})

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
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 7, GoalID: 42, Phase: PhaseDefine, Status: StatusWaitingApproval,
			}})
		case "GET /api/v1/agent-runs/7/approvals":
			_ = json.NewEncoder(w).Encode([]planclient.AgentApproval{{
				ID: 3, AgentRunID: 7, GoalID: 42, Status: "approved", Payload: proposal.Payload(),
			}})
		case "PATCH /api/v1/goals/42":
			if err := json.NewDecoder(r.Body).Decode(&updatedGoal); err != nil {
				t.Errorf("decode goal update: %v", err)
			}
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42, StatusID: 2}})
		case "PATCH /api/v1/goals/42/status":
			updatedPhase = true
			_ = json.NewEncoder(w).Encode(map[string]any{"status_id": 2})
		case "POST /api/v1/goals/42/phase":
			updatedPhase = true
			_ = json.NewEncoder(w).Encode(planclient.PhaseTransition{GoalID: 42, From: PhaseDefine, To: PhasePlan})
		case "PATCH /api/v1/agent-runs/7":
			var request planclient.UpdateAgentRunRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode agent run update: %v", err)
			}
			if request.Status != nil && *request.Status == StatusSucceeded {
				completed = true
			}
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		case "POST /api/v1/agent-runs/7/events":
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

	if updatedGoal.Title == nil || *updatedGoal.Title != "Ett definierat mål" {
		t.Errorf("titel = %v, vill ha %q", updatedGoal.Title, "Ett definierat mål")
	}
	if updatedGoal.Icon == nil || *updatedGoal.Icon != "heroicons:flag" {
		t.Errorf("ikon = %v, vill ha %q", updatedGoal.Icon, "heroicons:flag")
	}
	if !updatedPhase {
		t.Error("målet ska flyttas till planfasen")
	}
	if !completed {
		t.Error("körningen ska avslutas när definitionen är applicerad")
	}
}

// TestAvslagenDefinitionKörsInte verifierar att ett avslaget förslag
// inte skrivs till målet - och att agenten inte föreslår om det.
func TestAvslagenDefinitionKörsInte(t *testing.T) {
	var updatedGoal bool
	var createdApproval bool

	proposal := goaldefinition.Build(goaldefinition.Context{
		Goal:       planclient.Goal{ID: 42, Title: "Gör något viktigt", StatusID: 1, Agent: true},
		Definition: definer.Definition{Title: "Ett definierat mål"},
		NextPhase:  PhasePlan,
	})

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
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 7, GoalID: 42, Phase: PhaseDefine, Status: StatusWaitingApproval,
			}})
		case "GET /api/v1/agent-runs/7/approvals":
			_ = json.NewEncoder(w).Encode([]planclient.AgentApproval{{
				ID: 3, AgentRunID: 7, GoalID: 42, Status: "rejected", Payload: proposal.Payload(),
			}})
		case "POST /api/v1/agent-runs/7/approvals":
			createdApproval = true
			_ = json.NewEncoder(w).Encode(planclient.AgentApproval{ID: 4, AgentRunID: 7, GoalID: 42, Status: "pending"})
		case "PATCH /api/v1/goals/42":
			updatedGoal = true
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42, StatusID: 1}})
		case "PATCH /api/v1/agent-runs/7":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		case "POST /api/v1/agent-runs/7/events":
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

	if updatedGoal {
		t.Error("ett avslaget förslag ska inte skrivas till målet")
	}
	if createdApproval {
		t.Error("agenten ska inte föreslå om ett avslaget förslag")
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
		case "GET /api/v1/places":
			_ = json.NewEncoder(w).Encode([]planclient.Place{})
		case "POST /api/v1/agent-runs/7/approvals":
			_ = json.NewEncoder(w).Encode(planclient.AgentApproval{ID: 3, AgentRunID: 7, GoalID: 42, Status: "pending"})
		case "PATCH /api/v1/goals/42":
			_ = json.NewEncoder(w).Encode(map[string]planclient.Goal{"goal": {ID: 42, StatusID: 1}})
		case "PATCH /api/v1/goals/42/status":
			_ = json.NewEncoder(w).Encode(map[string]any{"status_id": 2})
		case "POST /api/v1/goals/42/phase":
			_ = json.NewEncoder(w).Encode(planclient.PhaseTransition{GoalID: 42, From: PhaseDefine, To: PhasePlan})
		case "PATCH /api/v1/agent-runs/7":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		case "POST /api/v1/agent-runs/7/events":
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
				ID: 7, GoalID: 42, Phase: PhasePlan, Status: StatusWaitingApproval,
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

	if updated.Status == nil || *updated.Status != StatusSucceeded {
		t.Errorf("körningens status = %v, vill ha %s", updated.Status, StatusSucceeded)
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

// TestUtanNyaFältAvslutarKörningen verifierar att en körning avslutas
// när definitionen inte har något nytt att föreslå - då finns inget att
// godkänna, så målet går bara vidare till nästa fas.
func TestUtanNyaFältAvslutarKörningen(t *testing.T) {
	var events []string
	var updatedPhase bool

	definer := fakeDefiner{definition: &definer.Definition{
		Title:       "Ett mål",
		Description: "beskrivning",
	}}

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
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/places":
			_ = json.NewEncoder(w).Encode([]planclient.Place{})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11, GoalID: 42})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/goals/42/status":
			updatedPhase = true
			w.WriteHeader(http.StatusNoContent)
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/goals/42/phase":
			updatedPhase = true
			_ = json.NewEncoder(w).Encode(planclient.PhaseTransition{GoalID: 42, From: PhaseDefine, To: PhasePlan})
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

	worker := New(client, definer, Config{OnlyGoalId: 42})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if !updatedPhase {
		t.Error("målet ska flyttas till planfasen")
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

// TestJobbetAvslutasNärKörningenÄrKlar verifierar att en jobbworker
// lämnar jobbet så snart körningen är klar - även om målet inte har nått
// verifieringsfasen, som när en livsprofil lämnar målet i planfasen.
func TestJobbetAvslutasNärKörningenÄrKlar(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/agent-runs":
			// Den senaste körningen är klar.
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{
				{ID: 10, GoalID: 42, Phase: PhasePlan, Status: StatusSucceeded},
			})
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

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42, Profile: "life", StopWhenFinished: true})
	err = worker.RunOnce(context.Background())
	if !errors.Is(err, ErrFinished) {
		t.Errorf("fel = %v, vill ha %v", err, ErrFinished)
	}
}

// TestAvbrutenKörningStopparWorkern verifierar att workern inte börjar
// om en körning som en människa avbrutit.
func TestAvbrutenKörningStopparWorkern(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/goals":
			_ = json.NewEncoder(w).Encode([]planclient.Goal{{
				ID: 42, Title: "Avbrutet mål", Description: "beskrivning", StatusID: 1, Agent: true,
			}})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: PhaseDefine},
				{ID: 2, Name: PhasePlan},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode([]planclient.AgentRun{{
				ID: 11, GoalID: 42, Phase: PhasePlan, Status: StatusCancelled,
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

	worker := New(client, fakeDefiner{}, Config{OnlyGoalId: 42})
	err = worker.RunOnce(context.Background())
	if !errors.Is(err, ErrFinished) {
		t.Errorf("fel = %v, vill ha %v", err, ErrFinished)
	}
}
