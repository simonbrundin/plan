package agentworker

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"plan/agent/internal/planclient"
)

func TestSyncPhaseKeepsGoalAndRunPhaseAligned(t *testing.T) {
	phases := []string{PhaseDefine, PhasePlan, PhaseExecute, PhaseVerify, PhaseCompleted}
	statusIDs := map[string]int{"define": 5, "plan": 8, "execute": 13, "verify": 21, "completed": 34}
	goalUpdates := make(map[int]string)
	runUpdates := make(map[string]bool)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/v1/statuses":
			statuses := make([]planclient.Status, 0, len(phases))
			for _, phase := range phases {
				statuses = append(statuses, planclient.Status{ID: int64(statusIDs[phase]), Name: phase})
			}
			_ = json.NewEncoder(w).Encode(statuses)
		case "/api/v1/goals/42/status":
			var request struct {
				StatusID int `json:"status_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode goal status: %v", err)
			}
			for phase, statusID := range statusIDs {
				if statusID == request.StatusID {
					goalUpdates[request.StatusID] = phase
				}
			}
		case "/api/v1/agent-runs/7":
			var request planclient.UpdateAgentRunRequest
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode agent run update: %v", err)
			}
			if request.Phase == nil || request.Status != nil {
				t.Errorf("run update = %+v, want only phase", request)
			} else {
				runUpdates[*request.Phase] = true
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
	worker := &Worker{planClient: client}

	for _, phase := range phases {
		if err := worker.syncPhase(context.Background(), 42, 7, phase); err != nil {
			t.Fatalf("syncPhase(%q): %v", phase, err)
		}
	}

	if len(goalUpdates) != len(phases) || len(runUpdates) != len(phases) {
		t.Fatalf("goal updates = %v, run updates = %v", goalUpdates, runUpdates)
	}
	for _, phase := range phases {
		if got := goalUpdates[statusIDs[phase]]; got != phase {
			t.Errorf("goal phase for status %d = %q, want %q", statusIDs[phase], got, phase)
		}
		if !runUpdates[phase] {
			t.Errorf("agent run phase %q was not updated", phase)
		}
	}
}
