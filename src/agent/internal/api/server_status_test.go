package api

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// orchestratorStatus är den del av /status och /health som beskriver
// orchestratorn.
type orchestratorStatus struct {
	Initialized bool   `json:"initialized"`
	Error       string `json:"error"`
}

func hamtaOrchestratorStatus(t *testing.T, body []byte) orchestratorStatus {
	t.Helper()

	var s struct {
		Orchestrator orchestratorStatus `json:"orchestrator"`
	}
	if err := json.Unmarshal(body, &s); err != nil {
		t.Fatalf("kunde inte tolka svaret: %v", err)
	}
	return s.Orchestrator
}

// Ett initieringsfel ska synas - inte bara ligga i loggen. Utan
// orchestratorn saknas gate-endpointern, och det ska gå att se att de
// saknas.
func TestStatusVisarInitieringsfelet(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orchestratorReady = false
	orchestratorInitErr = errors.New("StateStore kunde inte anslutas")
	t.Cleanup(func() {
		orchestratorReady = false
		orchestratorInitErr = nil
	})

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/status", nil)
	statusHandler(c)

	status := hamtaOrchestratorStatus(t, w.Body.Bytes())
	if status.Initialized {
		t.Error("orchestratorn ska inte vara initialiserad")
	}
	if status.Error == "" {
		t.Error("initieringsfelet ska synas i svaret")
	}
}

// När orchestratorn startade ska /status visa att den är klar.
func TestStatusVisarInitialiseradOrchestrator(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orchestratorReady = true
	orchestratorInitErr = nil
	t.Cleanup(func() { orchestratorReady = false })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/status", nil)
	statusHandler(c)

	status := hamtaOrchestratorStatus(t, w.Body.Bytes())
	if !status.Initialized {
		t.Error("orchestratorn ska vara initialiserad")
	}
	if status.Error != "" {
		t.Errorf("det ska inte finnas något fel, fick %q", status.Error)
	}
}

// Health svarar 200 även när orchestratorn inte startade - men tillståndet
// ska gå att se där också.
func TestHealthVisarOrchestratornsTillstand(t *testing.T) {
	gin.SetMode(gin.TestMode)

	orchestratorReady = true
	orchestratorInitErr = nil
	t.Cleanup(func() { orchestratorReady = false })

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/health", nil)
	healthHandler(c)

	if w.Code != 200 {
		t.Fatalf("health ska svara 200, fick %d", w.Code)
	}
	status := hamtaOrchestratorStatus(t, w.Body.Bytes())
	if !status.Initialized {
		t.Error("health ska visa att orchestratorn är initialiserad")
	}
}
