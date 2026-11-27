package agentworker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"plan/agent/internal/definer"
	"plan/agent/internal/planclient"
)

// TestStoppatFasbyteVantarPaManska kontrollerar att ett fasbyte som
// fasmaskinen stoppar inte gör körningen till ett fel.
//
// Maskinen svarar 409 med en skälkod när ett villkor saknas - till
// exempel när definitionen inte är godkänd. Då ska agenten stanna och
// vänta på en människa, inte falla och inte gå vidare på egen hand.
func TestStoppatFasbyteVantarPaManska(t *testing.T) {
	var runUpdate planclient.UpdateAgentRunRequest
	var events []planclient.CreateAgentRunEventRequest

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
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/goals/42/phase":
			// Fasmaskinen stoppar bytet - definitionen är inte godkänd.
			w.WriteHeader(http.StatusConflict)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"error":   "Conflict",
				"message": "målet saknar en godkänd definition",
				"code":    "definition_not_approved",
			})
		case r.Method == http.MethodPatch && r.URL.Path == "/api/v1/agent-runs/11":
			_ = json.NewDecoder(r.Body).Decode(&runUpdate)
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 11})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/agent-runs/11/events":
			var event planclient.CreateAgentRunEventRequest
			_ = json.NewDecoder(r.Body).Decode(&event)
			events = append(events, event)
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

	worker := New(client, fakeDefiner{definition: &definer.Definition{
		Title:       "Ett mål",
		Description: "beskrivning",
	}}, Config{OnlyGoalId: 42})
	if err := worker.RunOnce(context.Background()); err != nil {
		t.Fatalf("RunOnce: %v", err)
	}

	if runUpdate.Status == nil || *runUpdate.Status != StatusWaitingApproval {
		t.Errorf("körningens status = %v, vill ha %s", runUpdate.Status, StatusWaitingApproval)
	}
	if runUpdate.FinishedAt != nil {
		t.Error("en körning som väntar på en människa ska inte avslutas")
	}

	// Skälet ska synas i historiken - annars vet människan inte vad
	// som saknas.
	var hittade bool
	for _, event := range events {
		if event.Type != planclient.EventApprovalRequested {
			continue
		}
		if event.Payload["reason"] == "definition_not_approved" {
			hittade = true
		}
	}
	if !hittade {
		t.Errorf("händelserna ska innehålla skälet, fick %v", events)
	}
}

// TestSkalkodenLasesUrApisvaret kontrollerar att agenten kan läsa skälkoden
// ur ett stoppat fasbyte.
//
// Utan koden kan agenten inte veta om den ska be om hjälp eller göra om
// något - och felet blir en text som ingen parsar.
func TestSkalkodenLasesUrApisvaret(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_ = json.NewEncoder(w).Encode(map[string]any{
			"error":   "Conflict",
			"message": "målet saknar delmål",
			"code":    "missing_planned_children",
		})
	}))
	defer server.Close()

	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")
	client, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	_, err = client.RequestPhase(context.Background(), 42, "execute")
	if err == nil {
		t.Fatal("fasbytet ska ha stoppats")
	}
	if reason := planclient.ReasonCode(err); reason != "missing_planned_children" {
		t.Errorf("skälkod = %q, vill ha missing_planned_children", reason)
	}
}
