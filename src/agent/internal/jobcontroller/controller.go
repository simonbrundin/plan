// Package jobcontroller håller ihop Plan och Kubernetes.
//
// Plan äger målen och körningarna. Klustert äger jobben. Controllern gör
// de två världarna mötas: en körning som står i kö - en uttrycklig begäran
// om att få köras - får ett eget Kubernetes Job, och jobbets utfall förs
// tillbaka till Plan.
package jobcontroller

import (
	"context"
	"fmt"
	"log"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"

	"plan-agent/internal/jobrunner"
	"plan-agent/internal/planclient"
)

// Config styr hur controllern arbetar.
type Config struct {
	// PollInterval är hur ofta Plan och klustret läses av.
	PollInterval time.Duration
	// Namespace är namespace jobben skapas i.
	Namespace string
	// MaxConcurrentJobs begränsar hur många agentjobb som får köra samtidigt.
	// Noll använder standardvärdet fyra.
	MaxConcurrentJobs int
}

// Controller håller Plan och Kubernetes i synk.
type Controller struct {
	planClient *planclient.Client
	runner     *jobrunner.Runner
	config     Config
}

// New skapar en controller.
func New(planClient *planclient.Client, runner *jobrunner.Runner, config Config) *Controller {
	interval := config.PollInterval
	if interval == 0 {
		interval = time.Minute
	}

	return &Controller{
		planClient: planClient,
		runner:     runner,
		config: Config{
			PollInterval:      interval,
			Namespace:         config.Namespace,
			MaxConcurrentJobs: maxConcurrentJobs(config.MaxConcurrentJobs),
		},
	}
}

// Run kör controllern tills kontexten avbryts.
func (c *Controller) Run(ctx context.Context) {
	log.Printf("Job controller starting, polling every %s", c.config.PollInterval)

	ticker := time.NewTicker(c.config.PollInterval)
	defer ticker.Stop()

	if err := c.Reconcile(ctx); err != nil {
		log.Printf("Job controller reconcile failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Job controller stopping")
			return
		case <-ticker.C:
			if err := c.Reconcile(ctx); err != nil {
				log.Printf("Job controller reconcile failed: %v", err)
			}
		}
	}
}

// Reconcile gör en genomgång av körningarna och jobben.
//
// En körning som står i kö får ett Job. En körning som redan har ett Job
// lämnas ifred - bara ett utfall rapporteras tillbaka.
func (c *Controller) Reconcile(ctx context.Context) error {
	runs, err := c.planClient.GetAgentRuns(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch agent runs: %w", err)
	}

	for _, run := range runs {
		if err := c.reconcileRun(ctx, run); err != nil {
			log.Printf("Failed to reconcile run %d: %v", run.ID, err)
		}
	}

	return nil
}

// reconcileRun ser till att en köad körning får ett Job.
func (c *Controller) reconcileRun(ctx context.Context, run planclient.AgentRun) error {
	jobName := jobrunner.JobName(run.ID)

	// En avbruten körning ska inte lämna resurser efter sig.
	if run.Status == planclient.RunStatusCancelled {
		return c.cleanUpCancelled(ctx, run, jobName)
	}

	job, err := c.runner.Get(ctx, jobName)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return fmt.Errorf("failed to get job %s: %w", jobName, err)
		}

		// Jobbet finns inte. Bara en köad körning får skapa ett nytt jobb.
		if run.Status == "queued" {
			activeJobs, err := c.runner.ActiveJobCount(ctx)
			if err != nil {
				return err
			}
			if activeJobs >= c.config.MaxConcurrentJobs {
				log.Printf("Job capacity reached (%d) - run %d stays queued", c.config.MaxConcurrentJobs, run.ID)
				return nil
			}

			// Ett mål får högst ett pågående jobb - två körningar mot
			// samma repo skulle skriva över varandra.
			active, err := c.runner.HasActiveJobForGoal(ctx, run.GoalID, run.ID)
			if err != nil {
				return err
			}
			if active {
				log.Printf("Goal #%d already has a running job - run %d stays queued", run.GoalID, run.ID)
				return nil
			}

			if _, err := c.runner.Create(ctx, jobrunner.Spec{
				GoalID:  run.GoalID,
				RunID:   run.ID,
				Name:    jobName,
				Profile: run.AgentProfile,
			}); err != nil {
				return err
			}
			return nil
		}

		// En körning som redan är running utan motsvarande Job kan inte
		// återupptas säkert: workspace och process är borta. Markera den
		// som failed i stället för att lämna den fast för alltid eller
		// skapa ett nytt jobb med okänd historik. waiting_approval lämnas
		// orörd eftersom jobbet kan vara färdigt medan människan beslutar.
		if run.Status == planclient.RunStatusRunning {
			return c.reportMissingJob(ctx, run, jobName)
		}
		return nil
	}

	// Jobbet finns - för över ett misslyckande till körningen i Plan.
	return c.reportFailure(ctx, run, job)
}

// cleanUpCancelled tar bort jobbet och podden för en avbruten körning.
//
// Avbrytning ska lämna klustret som det var - annars fortsätter en agent
// att köra kod mot repot efter att en människa sagt stopp.
func (c *Controller) cleanUpCancelled(ctx context.Context, run planclient.AgentRun, jobName string) error {
	job, err := c.runner.Get(ctx, jobName)
	if err != nil || job == nil {
		// Jobbet finns redan inte - inget att städa.
		return nil
	}

	if err := c.runner.Delete(ctx, jobName); err != nil {
		return fmt.Errorf("failed to delete cancelled job %s: %w", jobName, err)
	}

	log.Printf("Deleted job %s - run %d was cancelled", jobName, run.ID)
	return nil
}

// reportMissingJob markerar en running-körning som misslyckad när dess Job
// är borta. Det är säkrare än att återskapa jobbet, eftersom den gamla
// workspacen och processhistoriken inte längre går att identifiera.
func (c *Controller) reportMissingJob(ctx context.Context, run planclient.AgentRun, jobName string) error {
	status := planclient.RunStatusFailed
	summary := fmt.Sprintf("Jobbet %s försvann medan körningen pågick", jobName)
	result := "AgentRun markerades som failed eftersom dess Kubernetes Job inte längre finns"
	finishedAt := time.Now().UTC()

	if _, err := c.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to mark orphaned run %d as failed: %w", run.ID, err)
	}

	if err := c.planClient.CreateAgentRunEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type: planclient.EventRunFailed,
		Payload: map[string]any{
			"reason":  "job_missing",
			"job":     jobName,
			"goal_id": run.GoalID,
			"message": "running AgentRun saknar sitt Kubernetes Job",
		},
	}); err != nil {
		log.Printf("Failed to record missing-job event for run %d: %v", run.ID, err)
	}

	return nil
}

// reportFailure markerar körningen som misslyckad om jobbet föll eller
// överskred sin aktiva deadline.
func (c *Controller) reportFailure(ctx context.Context, run planclient.AgentRun, job *batchv1.Job) error {
	if job == nil || !jobFailed(job) {
		return nil
	}
	if run.Status == planclient.RunStatusFailed || run.Status == planclient.RunStatusSucceeded {
		return nil
	}

	status := planclient.RunStatusFailed
	summary := fmt.Sprintf("Jobbet %s misslyckades", job.Name)
	result := fmt.Sprintf("Jobbet %s misslyckades utan att ge ett resultat", job.Name)
	if jobDeadlineExceeded(job) {
		summary = fmt.Sprintf("Jobbet %s överskred sin tidsgräns", job.Name)
		result = fmt.Sprintf("Jobbet %s stoppades av Kubernetes efter att deadline passerat", job.Name)
	}
	finishedAt := time.Now().UTC()
	if _, err := c.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to mark run %d as failed: %w", run.ID, err)
	}

	if err := c.planClient.CreateAgentRunEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type: planclient.EventRunFailed,
		Payload: map[string]any{
			"error":   "the kubernetes job failed",
			"job":     job.Name,
			"goal_id": run.GoalID,
		},
	}); err != nil {
		log.Printf("Failed to record run_failed for run %d: %v", run.ID, err)
	}

	log.Printf("Job %s failed - run %d for goal #%d marked as failed", job.Name, run.ID, run.GoalID)
	return nil
}

func maxConcurrentJobs(value int) int {
	if value <= 0 {
		return 4
	}
	return value
}

func jobFailed(job *batchv1.Job) bool {
	return job.Status.Failed > 0 || jobDeadlineExceeded(job)
}

func jobDeadlineExceeded(job *batchv1.Job) bool {
	for _, condition := range job.Status.Conditions {
		if condition.Type == batchv1.JobConditionType("DeadlineExceeded") && condition.Status == "True" {
			return true
		}
	}
	return false
}
