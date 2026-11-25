package jobrunner

import (
	"context"
	"strings"
	"testing"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"plan-agent/internal/planclient"
)

// testConfig ger en config med en angiven image.
//
// Utan image läser runnern agentens egen image ur podens spec, vilket
// inte finns i testerna.
func testConfig() Config {
	config := DefaultConfig()
	config.Image = "plan-agent:test"
	config.RepoURL = "https://github.com/simonbrundin/plan.git"
	return config
}

func testRunner() *Runner {
	client := fake.NewSimpleClientset()
	return New(client, testConfig())
}

func TestBuildSätterGränser(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if job.Name != "agent-worker-run-7" {
		t.Errorf("namn = %q", job.Name)
	}
	if job.Spec.ActiveDeadlineSeconds == nil || *job.Spec.ActiveDeadlineSeconds == 0 {
		t.Error("jobbet saknar deadline")
	}
	if job.Spec.TTLSecondsAfterFinished == nil {
		t.Error("jobbet saknar TTL")
	}

	container := job.Spec.Template.Spec.Containers[0]
	if container.Resources.Limits.Cpu().IsZero() {
		t.Error("containern saknar cpu-gräns")
	}
	if container.Resources.Limits.Memory().IsZero() {
		t.Error("containern saknar minnesgräns")
	}
}

// TestBuildMonterarIngenToken kontrollerar att jobbet inte kan prata med
// Kubernetes-API:t.
func TestBuildMonterarIngenToken(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	token := job.Spec.Template.Spec.AutomountServiceAccountToken
	if token == nil || *token {
		t.Error("jobbet monterar en ServiceAccount-token - den ska inte göra det")
	}
}

// TestBuildBegränsarRättigheter kontrollerar att containern är låst.
func TestBuildBegränsarRättigheter(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for _, container := range append(job.Spec.Template.Spec.Containers, job.Spec.Template.Spec.InitContainers...) {
		security := container.SecurityContext
		if security == nil {
			t.Fatalf("containern %s saknar security context", container.Name)
		}
		if security.AllowPrivilegeEscalation == nil || *security.AllowPrivilegeEscalation {
			t.Errorf("containern %s tillåter privilege escalation", container.Name)
		}
		if security.ReadOnlyRootFilesystem == nil || !*security.ReadOnlyRootFilesystem {
			t.Errorf("containern %s har skrivbart rotsystem", container.Name)
		}
		if security.RunAsNonRoot == nil || !*security.RunAsNonRoot {
			t.Errorf("containern %s kör som root", container.Name)
		}
		if len(security.Capabilities.Drop) == 0 {
			t.Errorf("containern %s behåller capabilities", container.Name)
		}
	}
}

// TestBuildHämtarHemligheterFrånSecrets kontrollerar att hemligheter inte
// skrivs i klartext i manifestet.
func TestBuildHämtarHemligheterFrånSecrets(t *testing.T) {
	client := fake.NewSimpleClientset()
	config := DefaultConfig()
	config.APIKeySecretName = "plan-agent-api-key"
	config.LLMSecretName = "plan-agent-llm"
	runner := New(client, config)

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	env := job.Spec.Template.Spec.Containers[0].Env
	found := map[string]bool{}
	for _, variable := range env {
		if variable.Value != "" && (variable.Name == "PLAN_API_KEY" || variable.Name == "MINIMAX_API_KEY") {
			t.Errorf("%s skrivs i klartext", variable.Name)
		}
		if variable.ValueFrom != nil && variable.ValueFrom.SecretKeyRef != nil {
			found[variable.Name] = true
		}
	}

	if !found["PLAN_API_KEY"] {
		t.Error("PLAN_API_KEY kommer inte från en Secret")
	}
	if !found["MINIMAX_API_KEY"] {
		t.Error("MINIMAX_API_KEY kommer inte från en Secret")
	}
}

// TestBuildKlonarRepot kontrollerar att repot klonas av en init-container.
func TestBuildKlonarRepot(t *testing.T) {
	client := fake.NewSimpleClientset()
	config := DefaultConfig()
	config.RepoURL = "https://github.com/simonbrundin/plan.git"
	config.RepoBranch = "main"
	runner := New(client, config)

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	if len(job.Spec.Template.Spec.InitContainers) != 1 {
		t.Fatalf("antal init-containers = %d, vill ha 1", len(job.Spec.Template.Spec.InitContainers))
	}

	script := job.Spec.Template.Spec.InitContainers[0].Command[2]
	if !contains(script, config.RepoURL) {
		t.Errorf("klonskriptet saknar repot: %s", script)
	}
	if !contains(script, "--branch main") {
		t.Errorf("klonskriptet saknar branchen: %s", script)
	}
}

// TestBuildSätterMåletsMiljö kontrollerar att jobbet vet vilket mål och
// vilket repo det gäller.
func TestBuildSätterMåletsMiljö(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	env := map[string]string{}
	for _, variable := range job.Spec.Template.Spec.Containers[0].Env {
		env[variable.Name] = variable.Value
	}

	if env["AGENT_GOAL_ID"] != "42" {
		t.Errorf("AGENT_GOAL_ID = %q", env["AGENT_GOAL_ID"])
	}
	if env["AGENT_REPO_PATH"] != "/workspace/repo" {
		t.Errorf("AGENT_REPO_PATH = %q", env["AGENT_REPO_PATH"])
	}
	if env["AGENT_INITIATOR"] != "job:agent-worker-run-7" {
		t.Errorf("AGENT_INITIATOR = %q", env["AGENT_INITIATOR"])
	}
}

// TestCreateSkaparJobbet kontrollerar att jobbet hamnar i klustret.
func TestCreateSkaparJobbet(t *testing.T) {
	client := fake.NewSimpleClientset()
	runner := New(client, testConfig())

	created, err := runner.Create(context.Background(), Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.Name != "agent-worker-run-7" {
		t.Errorf("namn = %q", created.Name)
	}

	fetched, err := runner.Get(context.Background(), "agent-worker-run-7")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if fetched.Spec.Template.Spec.Containers[0].Name != "agentworker" {
		t.Errorf("containern heter %q", fetched.Spec.Template.Spec.Containers[0].Name)
	}
}

// TestDeleteTarBortJobbet kontrollerar att jobbet kan städas bort.
func TestDeleteTarBortJobbet(t *testing.T) {
	client := fake.NewSimpleClientset()
	runner := New(client, testConfig())

	if _, err := runner.Create(context.Background(), Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := runner.Delete(context.Background(), "agent-worker-run-7"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	_, err := runner.Get(context.Background(), "agent-worker-run-7")
	if err == nil {
		t.Error("jobbet finns kvar efter borttagning")
	}
}

// TestJobNameÄrUnikt kontrollerar att jobbnamnet härleds från körningen.
func TestJobNameÄrUnikt(t *testing.T) {
	if JobName(7) != "agent-worker-run-7" {
		t.Errorf("namn = %q", JobName(7))
	}
}

// TestStatusÖversätterJobbetsUtfall kontrollerar att jobbets status blir
// begriplig.
func TestStatusÖversätterJobbetsUtfall(t *testing.T) {
	cases := []struct {
		name   string
		job    *batchv1.Job
		status string
	}{
		{
			name:   "klar",
			job:    &batchv1.Job{Status: batchv1.JobStatus{Succeeded: 1}},
			status: planclient.RunStatusSucceeded,
		},
		{
			name:   "misslyckad",
			job:    &batchv1.Job{Status: batchv1.JobStatus{Failed: 1}},
			status: "failed",
		},
		{
			name:   "pågår",
			job:    &batchv1.Job{Status: batchv1.JobStatus{Active: 1}},
			status: "running",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			if got := JobStatus(testCase.job); got != testCase.status {
				t.Errorf("status = %q, vill ha %q", got, testCase.status)
			}
		})
	}
}

// TestBuildKräverNamnOchMål kontrollerar att ett ofullständigt jobb inte
// byggs.
func TestBuildKräverNamnOchMål(t *testing.T) {
	runner := testRunner()

	if _, err := runner.Build(Spec{RunID: 7}, "plan-agent:test"); err == nil {
		t.Error("ett jobb utan namn ska avvisas")
	}
	if _, err := runner.Build(Spec{Name: "job"}, "plan-agent:test"); err == nil {
		t.Error("ett jobb utan mål ska avvisas")
	}
}

// TestPodSaknarPodSpecKontrollerar att podspecen är komplett.
func TestPodSaknarPodSpecKontrollerar(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	pod := job.Spec.Template.Spec
	if pod.RestartPolicy != corev1.RestartPolicyNever {
		t.Errorf("restart policy = %q", pod.RestartPolicy)
	}
	if len(pod.Volumes) == 0 {
		t.Error("jobbet saknar volymer")
	}
	if pod.SecurityContext == nil || pod.SecurityContext.RunAsNonRoot == nil || !*pod.SecurityContext.RunAsNonRoot {
		t.Error("poden kör som root")
	}
}

var _ = metav1.ObjectMeta{}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && indexOf(haystack, needle) >= 0
}

func indexOf(haystack, needle string) int {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return i
		}
	}
	return -1
}

// TestResolveImageLäserPodensImage kontrollerar att jobben kör samma
// image som agenten själv.
//
// Utan detta kan ett jobb köra en gammal tagg än den som är utbyggd.
func TestResolveImageLäserPodensImage(t *testing.T) {
	client := fake.NewSimpleClientset(&corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "agent-api-abc", Namespace: "plan"},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{Name: "agent-api", Image: "localhost:5000/plan-agent:tilt-123"}},
		},
	})

	runner := New(client, DefaultConfig())
	t.Setenv("POD_NAME", "agent-api-abc")

	image, err := runner.resolveImage(context.Background())
	if err != nil {
		t.Fatalf("resolveImage: %v", err)
	}
	if image != "localhost:5000/plan-agent:tilt-123" {
		t.Errorf("image = %q, vill ha agentens egen image", image)
	}
}

// TestResolveImageAngerFöreträde kontrollerar att en angiven image vinner
// framför podens egen.
func TestResolveImageAngerFöreträde(t *testing.T) {
	config := DefaultConfig()
	config.Image = "plan-agent:explicit"

	runner := New(fake.NewSimpleClientset(), config)

	image, err := runner.resolveImage(context.Background())
	if err != nil {
		t.Fatalf("resolveImage: %v", err)
	}
	if image != "plan-agent:explicit" {
		t.Errorf("image = %q, vill ha den angivna", image)
	}
}

// TestHasActiveJobForGoal kontrollerar att ett mål inte får två
// pågående jobb samtidigt.
//
// Två körningar mot samma repo skulle skriva över varandras ändringar.
func TestHasActiveJobForGoal(t *testing.T) {
	active := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-worker-run-7",
			Namespace: "plan",
			Labels: map[string]string{
				"plan.dev/goal-id":      "42",
				"plan.dev/agent-run-id": "7",
			},
		},
	}
	finished := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "agent-worker-run-8",
			Namespace: "plan",
			Labels: map[string]string{
				"plan.dev/goal-id":      "42",
				"plan.dev/agent-run-id": "8",
			},
		},
		Status: batchv1.JobStatus{Succeeded: 1},
	}

	client := fake.NewSimpleClientset(active, finished)
	runner := New(client, testConfig())

	has, err := runner.HasActiveJobForGoal(context.Background(), 42, 9)
	if err != nil {
		t.Fatalf("HasActiveJobForGoal: %v", err)
	}
	if !has {
		t.Error("målet har ett pågående jobb")
	}

	// Ett avslutat jobb räknas inte.
	has, err = runner.HasActiveJobForGoal(context.Background(), 43, 9)
	if err != nil {
		t.Fatalf("HasActiveJobForGoal: %v", err)
	}
	if has {
		t.Error("målet 43 har inget pågående jobb")
	}
}

// TestLivsprofilenKlonarInteRepot kontrollerar att en profil som inte får
// röra kod inte får tillgång till repot.
func TestLivsprofilenKlonarInteRepot(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7", Profile: "life"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	command := job.Spec.Template.Spec.InitContainers[0].Command[2]
	if strings.Contains(command, "git clone") {
		t.Errorf("livsprofilen får inte klona repot: %s", command)
	}
}

// TestLivsprofilenFårInteGitHemligheten kontrollerar att git-hemligheten
// bara finns för profiler som får skicka brancher.
func TestLivsprofilenFårInteGitHemligheten(t *testing.T) {
	config := testConfig()
	config.GitSecretName = "plan-agent-git"
	runner := New(fake.NewSimpleClientset(), config)

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7", Profile: "life"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "GIT_AUTH_HEADER" {
			t.Error("livsprofilen ska inte få git-hemligheten")
		}
	}
}

// TestKodprofilenKlonarRepot kontrollerar att kodprofilen får repot.
func TestKodprofilenKlonarRepot(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7", Profile: "coding"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	command := job.Spec.Template.Spec.InitContainers[0].Command[2]
	if !strings.Contains(command, "git clone") {
		t.Errorf("kodprofilen ska klona repot: %s", command)
	}
}

// TestLivsprofilenFårIngenReposökväg kontrollerar att en profil som inte
// får röra kod inte får en sökväg till repot.
func TestLivsprofilenFårIngenReposökväg(t *testing.T) {
	runner := testRunner()

	job, err := runner.Build(Spec{GoalID: 42, RunID: 7, Name: "agent-worker-run-7", Profile: "life"}, "plan-agent:test")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}

	for _, env := range job.Spec.Template.Spec.Containers[0].Env {
		if env.Name == "AGENT_REPO_PATH" && env.Value != "" {
			t.Errorf("livsprofilen ska inte få en reposökväg: %q", env.Value)
		}
	}
}
