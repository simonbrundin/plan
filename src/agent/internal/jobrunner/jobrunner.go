// Package jobrunner bygger Kubernetes Jobs för AgentRuns.
//
// En agentkörning ska inte dela processutrymme med API-servern. Ett Job ger
// isolering: egen pod, egna resursgränser, egen timeout och en ServiceAccount
// utan rättigheter. Jobbet kan avslutas utan att resten av systemet märker
// något.
package jobrunner

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"plan/agent/internal/planclient"
	"plan/agent/internal/profile"
)

// Config beskriver hur jobb ska byggas.
type Config struct {
	// Namespace är namespace jobbet skapas i.
	Namespace string
	// Image är agent-imagen som kör jobbet.
	Image string
	// ServiceAccount är den ServiceAccount jobbet använder.
	//
	// Den ska ha så få rättigheter som möjligt - jobbet behöver inte prata
	// med Kubernetes-API:t alls.
	ServiceAccount string
	// RepoURL är repot som klonas innan agenten börjar.
	RepoURL string
	// RepoBranch är branchen som klonas.
	RepoBranch string
	// PlanAPIURL är adressen till Plan.
	PlanAPIURL string
	// APIKeySecretName är hemligheten med Plan-nyckeln.
	APIKeySecretName string
	// APIKeySecretKey är nyckeln i hemligheten.
	APIKeySecretKey string
	// LLMSecretName är hemligheten med modellnyckeln.
	LLMSecretName string
	// LLMSecretKey är nyckeln i hemligheten.
	LLMSecretKey string
	// GitSecretName är hemligheten med git-autentiseringen.
	//
	// Tomt innebär att agenten bara kan läsa repot - den kan inte skicka
	// brancher tillbaka.
	GitSecretName string
	// GitSecretKey är nyckeln i hemligheten.
	GitSecretKey string
	// CPURequest, MemoryRequest, CPULimit, MemoryLimit sätter resursgränserna.
	CPURequest    string
	MemoryRequest string
	CPULimit      string
	MemoryLimit   string
	// ActiveDeadlineSeconds stoppar jobbet om det kör för länge.
	ActiveDeadlineSeconds int64
	// BackoffLimit är hur många gånger jobbet får starta om.
	BackoffLimit int32
	// TTLSecondsAfterFinished städar jobbet efter att det är klart.
	TTLSecondsAfterFinished int32
	// ImagePullPolicy styr hur imagen hämtas.
	ImagePullPolicy corev1.PullPolicy
}

// DefaultConfig ger en config med rimliga värden för en agentkörning.
//
// Image lämnas tom - då läses agentens egen image ur podens spec, så att
// jobben alltid kör den kod som är utbyggd.
func DefaultConfig() Config {
	return Config{
		Namespace:               "plan",
		Image:                   "",
		ServiceAccount:          "plan-agent-worker",
		CPURequest:              "100m",
		MemoryRequest:           "256Mi",
		CPULimit:                "1000m",
		MemoryLimit:             "2Gi",
		ActiveDeadlineSeconds:   3600,
		BackoffLimit:            0,
		TTLSecondsAfterFinished: 3600,
		ImagePullPolicy:         corev1.PullIfNotPresent,
	}
}

// Spec beskriver vad ett jobb ska göra.
type Spec struct {
	// GoalID är målet som ska utföras.
	GoalID int64
	// RunID är AgentRun-körningen som jobbet hör till.
	RunID int64
	// Name är jobbets namn. Det måste vara unikt i namespace.
	Name string
	// Profile är agentprofilen - coding eller life.
	//
	// Profilen bestämmer om jobbet får klona repot och vilka
	// resursgränser som gäller.
	Profile string
}

// Runner skapar och läser Kubernetes Jobs.
type Runner struct {
	client kubernetes.Interface
	config Config
}

// New skapar en runner mot den givna klienten.
func New(client kubernetes.Interface, config Config) *Runner {
	return &Runner{client: client, config: config}
}

// Create skapar jobbet för en körning.
func (r *Runner) Create(ctx context.Context, spec Spec) (*batchv1.Job, error) {
	image, err := r.resolveImage(ctx)
	if err != nil {
		return nil, err
	}

	job, err := r.Build(spec, image)
	if err != nil {
		return nil, err
	}

	created, err := r.client.BatchV1().Jobs(r.config.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job %s: %w", spec.Name, err)
	}

	log.Printf("Created agent job %s for goal #%d (run #%d) with image %s", created.Name, spec.GoalID, spec.RunID, image)
	return created, nil
}

// resolveImage bestämmer vilken image jobbet ska köra.
//
// En angiven image har företräde. Annars läses agentens egen image ur
// podens spec - jobbet ska alltid köra exakt den kod som är utbyggd,
// aldrig en äldre tagg.
func (r *Runner) resolveImage(ctx context.Context) (string, error) {
	if r.config.Image != "" {
		return r.config.Image, nil
	}

	podName := os.Getenv("POD_NAME")
	if podName == "" {
		return "", fmt.Errorf("no image configured and POD_NAME is not set")
	}

	pod, err := r.client.CoreV1().Pods(r.config.Namespace).Get(ctx, podName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to read own pod %s: %w", podName, err)
	}
	if len(pod.Spec.Containers) == 0 {
		return "", fmt.Errorf("pod %s has no containers", podName)
	}

	return pod.Spec.Containers[0].Image, nil
}

// Get hämtar ett jobb.
func (r *Runner) Get(ctx context.Context, name string) (*batchv1.Job, error) {
	return r.client.BatchV1().Jobs(r.config.Namespace).Get(ctx, name, metav1.GetOptions{})
}

// HasActiveJobForGoal svarar på om ett mål redan har ett pågående jobb.
//
// Ett mål ska ha högst ett jobb i taget - annars kan två agentkörningar
// skriva i samma repo och förstöra varandras ändringar.
func (r *Runner) HasActiveJobForGoal(ctx context.Context, goalID int64, excludeRunID int64) (bool, error) {
	jobs, err := r.client.BatchV1().Jobs(r.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("plan.dev/goal-id=%d", goalID),
	})
	if err != nil {
		return false, fmt.Errorf("failed to list jobs for goal %d: %w", goalID, err)
	}

	for _, job := range jobs.Items {
		if job.Labels["plan.dev/goal-id"] != fmt.Sprintf("%d", goalID) {
			continue
		}
		if job.Labels["plan.dev/agent-run-id"] == fmt.Sprintf("%d", excludeRunID) {
			continue
		}
		if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
			continue
		}
		return true, nil
	}

	return false, nil
}

// ActiveJobCount returnerar antalet jobb som fortfarande kör.
// Avslutade jobb räknas inte, även om TTL-cleanup ännu inte har hunnit ske.
func (r *Runner) ActiveJobCount(ctx context.Context) (int, error) {
	jobs, err := r.client.BatchV1().Jobs(r.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: "app=agent-worker",
	})
	if err != nil {
		return 0, fmt.Errorf("failed to list active agent jobs: %w", err)
	}

	active := 0
	for _, job := range jobs.Items {
		if job.Status.Succeeded == 0 && job.Status.Failed == 0 {
			active++
		}
	}
	return active, nil
}

// Delete tar bort ett jobb och dess pods.
func (r *Runner) Delete(ctx context.Context, name string) error {
	propagation := metav1.DeletePropagationForeground
	return r.client.BatchV1().Jobs(r.config.Namespace).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
}

// Build bygger jobbets manifest.
//
// Jobbet har tre säkerhetsgränser:
//
//   - ingen ServiceAccount-token monteras, jobbet kan inte röra API-servern
//   - resursgränser sätter en tak på CPU och minne
//   - en deadline stoppar jobbet om det fastnar
//
// Repot klonas av en init-container, så agenten arbetar på en kopia som
// bara finns under jobbets livstid.
func (r *Runner) Build(spec Spec, image string) (*batchv1.Job, error) {
	if spec.Name == "" {
		return nil, fmt.Errorf("job name is required")
	}
	if image == "" {
		return nil, fmt.Errorf("job image is required")
	}
	if spec.GoalID == 0 {
		return nil, fmt.Errorf("goal id is required")
	}

	cpuRequest, err := resource.ParseQuantity(valueOr(r.config.CPURequest, "100m"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu request: %w", err)
	}
	memoryRequest, err := resource.ParseQuantity(valueOr(r.config.MemoryRequest, "256Mi"))
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory request: %w", err)
	}
	policy, err := profile.PolicyFor(spec.Profile)
	if err != nil {
		return nil, err
	}

	cpuLimit, err := resource.ParseQuantity(valueOr(policy.CPULimit, valueOr(r.config.CPULimit, "1000m")))
	if err != nil {
		return nil, fmt.Errorf("failed to parse cpu limit: %w", err)
	}
	memoryLimit, err := resource.ParseQuantity(valueOr(policy.MemoryLimit, valueOr(r.config.MemoryLimit, "2Gi")))
	if err != nil {
		return nil, fmt.Errorf("failed to parse memory limit: %w", err)
	}

	labels := map[string]string{
		"app":                   "agent-worker",
		"plan.dev/goal-id":      fmt.Sprintf("%d", spec.GoalID),
		"plan.dev/agent-run-id": fmt.Sprintf("%d", spec.RunID),
	}

	workspace := corev1.Volume{
		Name: "workspace",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	tmp := corev1.Volume{
		Name: "tmp",
		VolumeSource: corev1.VolumeSource{
			EmptyDir: &corev1.EmptyDirVolumeSource{},
		},
	}

	env, err := r.env(spec)
	if err != nil {
		return nil, err
	}

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: r.config.Namespace,
			Labels:    labels,
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &r.config.BackoffLimit,
			ActiveDeadlineSeconds:   &r.config.ActiveDeadlineSeconds,
			TTLSecondsAfterFinished: &r.config.TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      corev1.RestartPolicyNever,
					ServiceAccountName: r.config.ServiceAccount,
					// Jobbet ska inte kunna prata med Kubernetes-API:t.
					AutomountServiceAccountToken: boolPtr(false),
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: boolPtr(true),
						RunAsUser:    int64Ptr(65532),
						RunAsGroup:   int64Ptr(65532),
						FSGroup:      int64Ptr(65532),
						SeccompProfile: &corev1.SeccompProfile{
							Type: corev1.SeccompProfileTypeRuntimeDefault,
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:            "clone",
							Image:           image,
							ImagePullPolicy: r.config.ImagePullPolicy,
							Command:         []string{"sh", "-c", cloneScript(r.repoURL(policy), r.config.RepoBranch)},
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
								{Name: "tmp", MountPath: "/tmp"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuRequest,
									corev1.ResourceMemory: memoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							SecurityContext: containerSecurityContext(),
						},
					},
					Containers: []corev1.Container{
						{
							Name:            "agent-worker",
							Image:           image,
							ImagePullPolicy: r.config.ImagePullPolicy,
							Command:         []string{"/app/agent-worker"},
							Env:             env,
							VolumeMounts: []corev1.VolumeMount{
								{Name: "workspace", MountPath: "/workspace"},
								{Name: "tmp", MountPath: "/tmp"},
							},
							Resources: corev1.ResourceRequirements{
								Requests: corev1.ResourceList{
									corev1.ResourceCPU:    cpuRequest,
									corev1.ResourceMemory: memoryRequest,
								},
								Limits: corev1.ResourceList{
									corev1.ResourceCPU:    cpuLimit,
									corev1.ResourceMemory: memoryLimit,
								},
							},
							SecurityContext: containerSecurityContext(),
						},
					},
					Volumes: []corev1.Volume{workspace, tmp},
				},
			},
		},
	}

	return job, nil
}

// env bygger miljövariablerna för jobbet.
//
// Hemligheter kommer från Kubernetes Secrets - de skrivs aldrig i
// manifestet och syns inte i loggar.
func (r *Runner) env(spec Spec) ([]corev1.EnvVar, error) {
	policy, err := profile.PolicyFor(spec.Profile)
	if err != nil {
		return nil, err
	}

	// En profil som inte får röra kod får inte ens en sökväg till repot.
	repoPath := ""
	if policy.NeedsRepo {
		repoPath = "/workspace/repo"
	}

	env := []corev1.EnvVar{
		{Name: "AGENT_GOAL_ID", Value: fmt.Sprintf("%d", spec.GoalID)},
		{Name: "AGENT_REPO_PATH", Value: repoPath},
		{Name: "AGENT_PROFILE", Value: policy.Name},
		{Name: "AGENT_INITIATOR", Value: fmt.Sprintf("job:%s", spec.Name)},
		{Name: "AGENT_POLL_INTERVAL", Value: "15s"},
		{Name: "AGENT_STOP_WHEN_FINISHED", Value: "true"},
		{Name: "HOME", Value: "/tmp"},
	}

	if r.config.PlanAPIURL != "" {
		env = append(env, corev1.EnvVar{Name: "PLAN_API_URL", Value: r.config.PlanAPIURL})
	}
	if r.config.APIKeySecretName != "" {
		env = append(env, corev1.EnvVar{
			Name: "PLAN_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: r.config.APIKeySecretName},
					Key:                  valueOr(r.config.APIKeySecretKey, "PLAN_API_KEY"),
				},
			},
		})
	}
	if r.config.GitSecretName != "" && policy.MayChangeCode {
		env = append(env, corev1.EnvVar{
			Name: "GIT_AUTH_HEADER",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: r.config.GitSecretName},
					Key:                  valueOr(r.config.GitSecretKey, "GIT_AUTH_HEADER"),
				},
			},
		})
	}
	if r.config.LLMSecretName != "" {
		env = append(env, corev1.EnvVar{
			Name: "MINIMAX_API_KEY",
			ValueFrom: &corev1.EnvVarSource{
				SecretKeyRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: r.config.LLMSecretName},
					Key:                  valueOr(r.config.LLMSecretKey, "MINIMAX_API_KEY"),
				},
			},
		})
	}

	return env, nil
}

// repoURL ger repot som profilen får klona.
//
// En profil som inte får röra kod får inte ens se repot.
func (r *Runner) repoURL(policy profile.Policy) string {
	if !policy.NeedsRepo {
		return ""
	}
	return r.config.RepoURL
}

// cloneScript bygger skriptet som klonar repot.
func cloneScript(url, branch string) string {
	if url == "" {
		return "mkdir -p /workspace/repo"
	}

	script := "set -eu\n"
	script += "git clone --depth 1"
	if branch != "" {
		script += " --branch " + branch
	}
	script += " " + url + " /workspace/repo\n"
	return script
}

// containerSecurityContext ger containern minimala rättigheter.
func containerSecurityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: boolPtr(false),
		ReadOnlyRootFilesystem:   boolPtr(true),
		RunAsNonRoot:             boolPtr(true),
		RunAsUser:                int64Ptr(65532),
		Capabilities: &corev1.Capabilities{
			Drop: []corev1.Capability{"ALL"},
		},
		SeccompProfile: &corev1.SeccompProfile{
			Type: corev1.SeccompProfileTypeRuntimeDefault,
		},
	}
}

// JobName ger ett jobbnamn som är unikt för en körning.
func JobName(runID int64) string {
	return fmt.Sprintf("agent-worker-run-%d", runID)
}

// JobNameForGoal ger ett jobbnamn för ett mål.
//
// Namnet används för att hitta samma jobb igen - controllern ska inte
// starta två jobb för samma mål.
func JobNameForGoal(goalID int64) string {
	return fmt.Sprintf("agent-worker-goal-%d", goalID)
}

// JobStatus översätter ett jobbs tillstånd till en AgentRun-status.
//
// Översättningen följer AgentRuns egna statusar, så att en körning som
// startats i ett Job ser samma ut i Plan som en körning som startats av
// schemaläggaren.
func JobStatus(job *batchv1.Job) string {
	if job == nil {
		return "queued"
	}

	switch {
	case job.Status.Succeeded > 0:
		return planclient.RunStatusSucceeded
	case job.Status.Failed > 0:
		return "failed"
	case job.Status.Active > 0:
		return "running"
	default:
		return "queued"
	}
}

func valueOr(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func boolPtr(value bool) *bool {
	return &value
}

func int64Ptr(value int64) *int64 {
	return &value
}
