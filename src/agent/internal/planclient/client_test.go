package planclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestUpdateGoalPhaseUsesStatusNameAndDedicatedEndpoint(t *testing.T) {
	phases := []string{"define", "plan", "execute", "verify", "completed"}
	statusIDs := map[string]int{"define": 17, "plan": 23, "execute": 31, "verify": 47, "completed": 59}
	updated := make(map[string]int)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-API-Key") != "test-key" {
			t.Errorf("X-API-Key = %q, want test-key", r.Header.Get("X-API-Key"))
		}

		switch r.URL.Path {
		case "/api/v1/statuses":
			statuses := make([]Status, 0, len(phases))
			for _, phase := range phases {
				statuses = append(statuses, Status{ID: int64(statusIDs[phase]), Name: phase})
			}
			_ = json.NewEncoder(w).Encode(statuses)
		case "/api/v1/goals/42/status":
			if r.Method != http.MethodPatch {
				t.Errorf("method = %s, want PATCH", r.Method)
			}
			var request struct {
				StatusID int `json:"status_id"`
			}
			if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
				t.Errorf("decode status request: %v", err)
			}
			for name, id := range statusIDs {
				if request.StatusID == id {
					updated[name] = request.StatusID
				}
			}
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := &Client{baseURL: server.URL, apiKey: "test-key", httpClient: server.Client()}
	for _, phase := range phases {
		if err := client.UpdateGoalPhase(context.Background(), 42, phase); err != nil {
			t.Fatalf("UpdateGoalPhase(%q): %v", phase, err)
		}
	}

	if !reflect.DeepEqual(updated, statusIDs) {
		t.Fatalf("updated status IDs = %v, want %v", updated, statusIDs)
	}
}

func TestUpdateGoalPhaseRejectsUnknownPhase(t *testing.T) {
	client := &Client{}
	if err := client.UpdateGoalPhase(context.Background(), 42, "unknown"); err == nil {
		t.Fatal("UpdateGoalPhase accepted an unknown phase")
	}
}
