package k8sjob

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// Config holds configuration for the K8s job runner
type Config struct {
	Namespace               string
	Image                   string
	ServiceAccount          string
	ImagePullSecret         string
	CPURequest              string
	MemoryRequest           string
	CPULimit                string
	MemoryLimit             string
	TTLSecondsAfterFinished int32 // Clean up job after N seconds
	BackoffLimit            int32
	ActiveDeadlineSeconds   int64
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Namespace:               "agents",
		Image:                   "ghcr.io/simon/agent:latest",
		ServiceAccount:          "agent",
		CPURequest:              "100m",
		MemoryRequest:           "256Mi",
		CPULimit:                "1000m",
		MemoryLimit:             "2Gi",
		TTLSecondsAfterFinished: 3600, // 1 hour
		BackoffLimit:            3,
		ActiveDeadlineSeconds:   3600, // 1 hour
	}
}

// Runner runs Kubernetes jobs
type Runner struct {
	client *kubernetes.Clientset
	config *Config
}

// NewRunner creates a new K8s job runner
func NewRunner(config *Config) (*Runner, error) {
	if config == nil {
		config = DefaultConfig()
	}

	client, err := getClient()
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes client: %w", err)
	}

	return &Runner{
		client: client,
		config: config,
	}, nil
}

// getClient creates a Kubernetes client from config or in-cluster
func getClient() (*kubernetes.Clientset, error) {
	var config *rest.Config
	var err error

	// Try in-cluster config first
	config, err = rest.InClusterConfig()
	if err != nil {
		// Fall back to kubeconfig
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			home, _ := os.UserHomeDir()
			kubeconfig = home + "/.kube/config"
		}

		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, fmt.Errorf("failed to get Kubernetes config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create clientset: %w", err)
	}

	return clientset, nil
}

// JobSpec defines what the job should run
type JobSpec struct {
	Name         string
	Command      []string
	Args         []string
	Env          []corev1.EnvVar
	Volumes      []corev1.Volume
	VolumeMounts []corev1.VolumeMount
	Task         string // Task description for the agent
}

// Run creates and runs a Kubernetes job
func (r *Runner) Run(ctx context.Context, spec *JobSpec) (*batchv1.Job, error) {
	job := r.createJob(spec)

	result, err := r.client.BatchV1().Jobs(r.config.Namespace).Create(ctx, job, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create job: %w", err)
	}

	log.Printf("Created job: %s", result.Name)
	return result, nil
}

// RunAndWait creates a job and waits for completion
func (r *Runner) RunAndWait(ctx context.Context, spec *JobSpec, timeout time.Duration) (*JobResult, error) {
	job, err := r.Run(ctx, spec)
	if err != nil {
		return nil, err
	}

	return r.WaitForJob(ctx, job.Name, timeout)
}

// WaitForJob waits for a job to complete
func (r *Runner) WaitForJob(ctx context.Context, jobName string, timeout time.Duration) (*JobResult, error) {
	watch, err := r.client.BatchV1().Jobs(r.config.Namespace).Watch(ctx, metav1.SingleObject(metav1.ObjectMeta{
		Name: jobName,
	}))
	if err != nil {
		return nil, fmt.Errorf("failed to watch job: %w", err)
	}
	defer watch.Stop()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	result := &JobResult{JobName: jobName}
	startTime := time.Now()

	for {
		select {
		case <-ctx.Done():
			result.TimedOut = true
			return result, ctx.Err()
		case event, ok := <-watch.ResultChan():
			if !ok {
				return result, fmt.Errorf("watch channel closed")
			}

			job, ok := event.Object.(*batchv1.Job)
			if !ok {
				continue
			}

			result.Job = job
			result.Duration = time.Since(startTime)

			// Check job status
			if job.Status.Succeeded > 0 {
				result.Succeeded = true
				return result, nil
			}
			if job.Status.Failed > 0 {
				result.Failed = true
				result.FailureCount = int(job.Status.Failed)
				return result, fmt.Errorf("job failed")
			}
		}
	}
}

// GetLogs retrieves logs from a completed job
func (r *Runner) GetLogs(ctx context.Context, jobName string) (string, error) {
	// Get the pod selector
	job, err := r.client.BatchV1().Jobs(r.config.Namespace).Get(ctx, jobName, metav1.GetOptions{})
	if err != nil {
		return "", fmt.Errorf("failed to get job: %w", err)
	}

	selector, err := metav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return "", fmt.Errorf("failed to get selector: %w", err)
	}

	// Get pods
	pods, err := r.client.CoreV1().Pods(r.config.Namespace).List(ctx, metav1.ListOptions{
		LabelSelector: selector.String(),
	})
	if err != nil {
		return "", fmt.Errorf("failed to list pods: %w", err)
	}

	var allLogs strings.Builder
	for _, pod := range pods.Items {
		// Get logs for each container
		for _, container := range pod.Spec.Containers {
			logs, err := r.getPodLogs(ctx, pod.Name, container.Name)
			if err != nil {
				logs = fmt.Sprintf("[Error getting logs: %v]\n", err)
			}
			_, _ = fmt.Fprintf(&allLogs, "=== Pod: %s, Container: %s ===\n", pod.Name, container.Name)
			allLogs.WriteString(logs)
			allLogs.WriteString("\n")
		}
	}

	return allLogs.String(), nil
}

// getPodLogs retrieves logs from a specific pod and container
func (r *Runner) getPodLogs(ctx context.Context, podName, containerName string) (string, error) {
	req := r.client.CoreV1().Pods(r.config.Namespace).GetLogs(podName, &corev1.PodLogOptions{
		Container: containerName,
	})

	logs, err := req.Stream(ctx)
	if err != nil {
		return "", err
	}
	defer func() { _ = logs.Close() }()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, logs)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// Delete removes a job and its pods
func (r *Runner) Delete(ctx context.Context, jobName string) error {
	// Delete the job (pods will be deleted due to TTL if configured)
	propagation := metav1.DeletePropagationForeground
	err := r.client.BatchV1().Jobs(r.config.Namespace).Delete(ctx, jobName, metav1.DeleteOptions{
		PropagationPolicy: &propagation,
	})
	if err != nil {
		return fmt.Errorf("failed to delete job: %w", err)
	}

	log.Printf("Deleted job: %s", jobName)
	return nil
}

// DeleteCompleted deletes completed jobs older than the TTL
func (r *Runner) DeleteCompleted(ctx context.Context) error {
	jobs, err := r.client.BatchV1().Jobs(r.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list jobs: %w", err)
	}

	for _, job := range jobs.Items {
		if job.Status.Succeeded > 0 || job.Status.Failed > 0 {
			// Job is complete, check TTL
			if job.Spec.TTLSecondsAfterFinished != nil {
				finishedTime := job.Status.CompletionTime
				if finishedTime != nil {
					age := time.Since(finishedTime.Time)
					if age > time.Duration(*job.Spec.TTLSecondsAfterFinished)*time.Second {
						if err := r.Delete(ctx, job.Name); err != nil {
							log.Printf("failed to delete expired job %s: %v", job.Name, err)
						}
					}
				}
			}
		}
	}

	return nil
}

// List returns all jobs in the namespace
func (r *Runner) List(ctx context.Context) ([]*batchv1.Job, error) {
	jobs, err := r.client.BatchV1().Jobs(r.config.Namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list jobs: %w", err)
	}

	result := make([]*batchv1.Job, len(jobs.Items))
	for i := range jobs.Items {
		result[i] = &jobs.Items[i]
	}

	return result, nil
}

// GetJob returns a specific job
func (r *Runner) GetJob(ctx context.Context, name string) (*batchv1.Job, error) {
	return r.client.BatchV1().Jobs(r.config.Namespace).Get(ctx, name, metav1.GetOptions{})
}

// JobResult holds the result of a completed job
type JobResult struct {
	JobName      string
	Job          *batchv1.Job
	Succeeded    bool
	Failed       bool
	TimedOut     bool
	Duration     time.Duration
	FailureCount int
}

// createJob creates a Kubernetes Job from a spec
func (r *Runner) createJob(spec *JobSpec) *batchv1.Job {
	restartPolicy := corev1.RestartPolicyOnFailure

	cpuRequest, _ := resource.ParseQuantity(r.config.CPURequest)
	memoryRequest, _ := resource.ParseQuantity(r.config.MemoryRequest)
	cpuLimit, _ := resource.ParseQuantity(r.config.CPULimit)
	memoryLimit, _ := resource.ParseQuantity(r.config.MemoryLimit)

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      spec.Name,
			Namespace: r.config.Namespace,
			Labels: map[string]string{
				"app":      "agent",
				"job-name": spec.Name,
			},
		},
		Spec: batchv1.JobSpec{
			BackoffLimit:            &r.config.BackoffLimit,
			ActiveDeadlineSeconds:   &r.config.ActiveDeadlineSeconds,
			TTLSecondsAfterFinished: &r.config.TTLSecondsAfterFinished,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app":      "agent",
						"job-name": spec.Name,
					},
				},
				Spec: corev1.PodSpec{
					RestartPolicy:      restartPolicy,
					ServiceAccountName: r.config.ServiceAccount,
					Containers: []corev1.Container{
						{
							Name:            "agent",
							Image:           r.config.Image,
							ImagePullPolicy: corev1.PullAlways,
							Command:         spec.Command,
							Args:            spec.Args,
							Env:             spec.Env,
							VolumeMounts:    spec.VolumeMounts,
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
						},
					},
					Volumes: spec.Volumes,
				},
			},
		},
	}

	if r.config.ImagePullSecret != "" {
		job.Spec.Template.Spec.ImagePullSecrets = []corev1.LocalObjectReference{
			{Name: r.config.ImagePullSecret},
		}
	}

	return job
}

// GenerateManifest generates a Kubernetes Job manifest as YAML
func GenerateManifest(config *Config, spec *JobSpec) ([]byte, error) {
	if config == nil {
		config = DefaultConfig()
	}

	runner := &Runner{config: config}
	job := runner.createJob(spec)

	// Use JSON marshal and convert to YAML
	jsonData, err := json.MarshalIndent(job, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal job: %w", err)
	}

	return jsonData, nil
}

// JobInfo represents simplified job information
type JobInfo struct {
	Name      string        `json:"name"`
	Namespace string        `json:"namespace"`
	Status    string        `json:"status"`
	Succeeded int32         `json:"succeeded"`
	Failed    int32         `json:"failed"`
	Active    int32         `json:"active"`
	StartTime time.Time     `json:"start_time,omitempty"`
	Duration  time.Duration `json:"duration"`
}

// ToInfo converts a Job to JobInfo
func JobToInfo(j *batchv1.Job) *JobInfo {
	info := &JobInfo{
		Name:      j.Name,
		Namespace: j.Namespace,
		Succeeded: j.Status.Succeeded,
		Failed:    j.Status.Failed,
		Active:    j.Status.Active,
	}

	if j.Status.Active > 0 {
		info.Status = "Running"
	} else if j.Status.Succeeded > 0 {
		info.Status = "Succeeded"
	} else if j.Status.Failed > 0 {
		info.Status = "Failed"
	} else {
		info.Status = "Pending"
	}

	if j.Status.StartTime != nil {
		info.StartTime = j.Status.StartTime.Time
		info.Duration = time.Since(j.Status.StartTime.Time)
	}

	return info
}

// ToJSON converts JobInfo to JSON
func (j *JobInfo) ToJSON() ([]byte, error) {
	return json.MarshalIndent(j, "", "  ")
}

// Watcher watches for job events
type Watcher struct {
	client    *kubernetes.Clientset
	namespace string
}

// NewWatcher creates a new job watcher
func NewWatcher(namespace string) (*Watcher, error) {
	client, err := getClient()
	if err != nil {
		return nil, err
	}

	return &Watcher{
		client:    client,
		namespace: namespace,
	}, nil
}

// Event represents a job event
type Event struct {
	Type   string
	Job    *batchv1.Job
	Reason string
}

// Watch watches for job changes
func (w *Watcher) Watch(ctx context.Context) (<-chan *Event, error) {
	watch, err := w.client.BatchV1().Jobs(w.namespace).Watch(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	events := make(chan *Event)

	go func() {
		defer close(events)
		defer watch.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-watch.ResultChan():
				if !ok {
					return
				}

				job, ok := e.Object.(*batchv1.Job)
				if !ok {
					continue
				}

				reason := ""
				if cond := getJobCondition(job); cond != nil {
					reason = cond.Reason
				}

				events <- &Event{
					Type:   string(e.Type),
					Job:    job,
					Reason: reason,
				}
			}
		}
	}()

	return events, nil
}

// getJobCondition gets the current job condition
func getJobCondition(job *batchv1.Job) *batchv1.JobCondition {
	for i := range job.Status.Conditions {
		cond := &job.Status.Conditions[i]
		if cond.Status == corev1.ConditionTrue {
			return cond
		}
	}
	return nil
}

// CreateSecretFromEnv creates a Kubernetes Secret from environment variables
func (r *Runner) CreateSecretFromEnv(ctx context.Context, secretName string, envVars []corev1.EnvVar) error {
	data := make(map[string][]byte)
	for _, env := range envVars {
		if env.ValueFrom != nil {
			continue
		}
		data[env.Name] = []byte(env.Value)
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: r.config.Namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: data,
	}

	_, err := r.client.CoreV1().Secrets(r.config.Namespace).Create(ctx, secret, metav1.CreateOptions{})
	return err
}
