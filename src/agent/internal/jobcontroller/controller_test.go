package jobcontroller

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"plan-agent/internal/jobrunner"
	"plan-agent/internal/planclient"
)

type planState struct {
	goals       []planclient.Goal
	runs        []planclient.AgentRun
	updatedRuns []planclient.UpdateAgentRunRequest
	events      []planclient.CreateAgentRunEventRequest
}

// newPlanServer ger en testserver som svarar på Plan-anropen.
func newPlanServer(t *testing.T, state *planState) *httptest.Server {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		switch r.Method + " " + r.URL.Path {
		case "GET /api/v1/goals":
			_ = json.NewEncoder(w).Encode(state.goals)
		case "GET /api/v1/agent-runs":
			_ = json.NewEncoder(w).Encode(state.runs)
		case "GET /api/v1/statuses":
			_ = json.NewEncoder(w).Encode([]planclient.Status{
				{ID: 1, Name: "define"},
				{ID: 2, Name: "plan"},
				{ID: 4, Name: "verify"},
			})
		case "PATCH /api/v1/agent-runs/7":
			var req planclient.UpdateAgentRunRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			state.updatedRuns = append(state.updatedRuns, req)
			_ = json.NewEncoder(w).Encode(planclient.AgentRun{ID: 7})
		case "POST /api/v1/agent-runs/7/events":
			var req planclient.CreateAgentRunEventRequest
			_ = json.NewDecoder(r.Body).Decode(&req)
			state.events = append(state.events, req)
			w.WriteHeader(http.StatusCreated)
		default:
			http.NotFound(w, r)
		}
	}))

	t.Cleanup(server.Close)
	return server
}

func newController(t *testing.T, state *planState, client *fake.Clientset) *Controller {
	t.Helper()

	server := newPlanServer(t, state)
	t.Setenv("PLAN_API_URL", server.URL)
	t.Setenv("PLAN_API_KEY", "test-key")

	planClient, err := planclient.New()
	if err != nil {
		t.Fatalf("create Plan API client: %v", err)
	}

	config := jobrunner.DefaultConfig()
	config.Image = "plan-agent:test"

	return New(planClient, jobrunner.New(client, config), Config{
		Namespace: "plan",
	})
}

// TestReconcileStartarJobb kontrollerar att en köad körning får ett Job.
func TestReconcileStartarJobb(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: "queued"}},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	job, err := client.BatchV1().Jobs("plan").Get(context.Background(), "agent-worker-run-7", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("jobbet skapades inte: %v", err)
	}
	if job.Spec.Template.Spec.Containers[0].Name != "agentworker" {
		t.Errorf("containern heter %q", job.Spec.Template.Spec.Containers[0].Name)
	}
}

// TestReconcileStartarInteDubblaJobb kontrollerar att samma mål inte får
// två jobb.
func TestReconcileStartarInteDubblaJobb(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: "queued"}},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	for i := 0; i < 3; i++ {
		if err := controller.Reconcile(context.Background()); err != nil {
			t.Fatalf("Reconcile: %v", err)
		}
	}

	jobs, err := client.BatchV1().Jobs("plan").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Items) != 1 {
		t.Errorf("antal jobb = %d, vill ha 1", len(jobs.Items))
	}
}

// TestReconcileMarkerarMisslyckatJobb kontrollerar att ett jobb som föll
// gör att körningen i Plan markeras som misslyckad.
func TestReconcileMarkerarMisslyckatJobb(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: "running"}},
	}

	client := fake.NewSimpleClientset()
	_, err := client.BatchV1().Jobs("plan").Create(context.Background(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-worker-run-7", Namespace: "plan"},
		Status:     batchv1.JobStatus{Failed: 1},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	controller := newController(t, state, client)
	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(state.updatedRuns) != 1 {
		t.Fatalf("antal uppdateringar = %d, vill ha 1", len(state.updatedRuns))
	}
	if state.updatedRuns[0].Status == nil || *state.updatedRuns[0].Status != "failed" {
		t.Errorf("status = %v, vill ha failed", state.updatedRuns[0].Status)
	}
	if len(state.events) != 1 || state.events[0].Type != planclient.EventRunFailed {
		t.Errorf("händelser = %v", state.events)
	}
}

// TestReconcileMarkerarRunningUtanJobbSomMisslyckad kontrollerar att en
// körning inte blir kvar för evigt när Kubernetes-jobbet försvinner.
func TestReconcileMarkerarRunningUtanJobbSomMisslyckad(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: planclient.RunStatusRunning}},
	}
	controller := newController(t, state, fake.NewSimpleClientset())

	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(state.updatedRuns) != 1 || state.updatedRuns[0].Status == nil || *state.updatedRuns[0].Status != planclient.RunStatusFailed {
		t.Fatalf("uppdateringar = %v, vill ha en failed-uppdatering", state.updatedRuns)
	}
	if len(state.events) != 1 || state.events[0].Type != planclient.EventRunFailed {
		t.Fatalf("händelser = %v, vill ha run_failed", state.events)
	}
	if state.events[0].Payload["reason"] != "job_missing" {
		t.Errorf("orsak = %v, vill ha job_missing", state.events[0].Payload["reason"])
	}
}

// TestReconcileRörInteKlartKörning kontrollerar att en pågående körning
// inte får ett nytt jobb.
func TestReconcileRörInteKlartKörning(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: "running"}},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	jobs, err := client.BatchV1().Jobs("plan").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Items) != 0 {
		t.Errorf("antal jobb = %d, vill ha 0", len(jobs.Items))
	}
}

// TestReconcileDeadlineMarkerarMisslyckatJobb kontrollerar att Kubernetes
// deadline rapporteras även innan Failed-räknaren hunnit uppdateras.
func TestReconcileDeadlineMarkerarMisslyckatJobb(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: planclient.RunStatusRunning}},
	}
	client := fake.NewSimpleClientset()
	_, err := client.BatchV1().Jobs("plan").Create(context.Background(), &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-worker-run-7", Namespace: "plan"},
		Status: batchv1.JobStatus{Conditions: []batchv1.JobCondition{{
			Type:   batchv1.JobConditionType("DeadlineExceeded"),
			Status: "True",
		}}},
	}, metav1.CreateOptions{})
	if err != nil {
		t.Fatalf("create job: %v", err)
	}

	controller := newController(t, state, client)
	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	if len(state.updatedRuns) != 1 || state.updatedRuns[0].Status == nil || *state.updatedRuns[0].Status != planclient.RunStatusFailed {
		t.Fatalf("uppdateringar = %v, vill ha failed", state.updatedRuns)
	}
	if got := *state.updatedRuns[0].Summary; got != "Jobbet agent-worker-run-7 överskred sin tidsgräns" {
		t.Errorf("summary = %q", got)
	}
}

// TestReconcileHopparÖverKörningarSomInteÄrKöade kontrollerar att en
// körning som redan pågår inte får ett nytt jobb.
func TestReconcileHopparÖverKörningarSomInteÄrKöade(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: planclient.RunStatusSucceeded}},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	jobs, err := client.BatchV1().Jobs("plan").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Items) != 0 {
		t.Errorf("antal jobb = %d, vill ha 0", len(jobs.Items))
	}
}

var _ = time.Second

// TestReconcileStartarFleraOlikaMålParallellt kontrollerar att jobb för
// olika mål inte blockeras av regeln om ett aktivt jobb per mål.
func TestReconcileStartarFleraOlikaMålParallellt(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{
			{ID: 7, GoalID: 42, Status: "queued"},
			{ID: 8, GoalID: 43, Status: "queued"},
		},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	jobs, err := client.BatchV1().Jobs("plan").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Items) != 2 {
		t.Fatalf("antal jobb = %d, vill ha 2", len(jobs.Items))
	}
}

// TestReconcileStartarEttJobbPerMål kontrollerar att två köade körningar
// för samma mål inte ger två jobb.
//
// Två körningar mot samma repo skulle skriva över varandras ändringar.
func TestReconcileStartarEttJobbPerMål(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{
			{ID: 7, GoalID: 42, Status: "queued"},
			{ID: 8, GoalID: 42, Status: "queued"},
		},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	for i := 0; i < 3; i++ {
		if err := controller.Reconcile(context.Background()); err != nil {
			t.Fatalf("Reconcile: %v", err)
		}
	}

	jobs, err := client.BatchV1().Jobs("plan").List(context.Background(), metav1.ListOptions{})
	if err != nil {
		t.Fatalf("list jobs: %v", err)
	}
	if len(jobs.Items) != 1 {
		t.Errorf("antal jobb = %d, vill ha 1", len(jobs.Items))
	}
}

// TestReconcileTarBortJobbFörAvbrutenKörning kontrollerar att en
// avbruten körning inte lämnar jobb eller podd efter sig i klustret.
func TestReconcileTarBortJobbFörAvbrutenKörning(t *testing.T) {
	state := &planState{
		runs: []planclient.AgentRun{{ID: 7, GoalID: 42, Status: planclient.RunStatusCancelled}},
	}
	client := fake.NewSimpleClientset()
	controller := newController(t, state, client)

	// Först skapas jobbet - sedan avbryts körningen.
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-worker-run-7", Namespace: "plan"},
	}
	if _, err := client.BatchV1().Jobs("plan").Create(context.Background(), job, metav1.CreateOptions{}); err != nil {
		t.Fatalf("skapa jobb: %v", err)
	}

	if err := controller.Reconcile(context.Background()); err != nil {
		t.Fatalf("Reconcile: %v", err)
	}

	_, err := client.BatchV1().Jobs("plan").Get(context.Background(), "agent-worker-run-7", metav1.GetOptions{})
	if err == nil {
		t.Error("jobbet finns kvar - det borde ha tagits bort")
	}
}
