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

	"plan-agent/internal/jobrunner"
	"plan-agent/internal/planclient"
)

// Config styr hur controllern arbetar.
type Config struct {
	// PollInterval är hur ofta Plan och klustret läses av.
	PollInterval time.Duration
	// Namespace är namespace jobben skapas i.
	Namespace string
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
			PollInterval: interval,
			Namespace:    config.Namespace,
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

	job, err := c.runner.Get(ctx, jobName)
	if err != nil {
		// Jobbet finns inte. Bara en köad körning startar ett nytt jobb.
		if run.Status != "queued" {
			return nil
		}

		// Ett mål får ha högst ett pågående jobb - två körningar mot
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

	// Jobbet finns - för över ett misslyckande till körningen i Plan.
	return c.reportFailure(ctx, run, job)
}

// reportFailure markerar körningen som misslyckad om jobbet föll.
func (c *Controller) reportFailure(ctx context.Context, run planclient.AgentRun, job *batchv1.Job) error {
	if job == nil || job.Status.Failed == 0 {
		return nil
	}
	if run.Status == "failed" || run.Status == "done" {
		return nil
	}

	status := "failed"
	summary := fmt.Sprintf("Jobbet %s misslyckades", job.Name)
	result := fmt.Sprintf("Jobbet %s misslyckades utan att ge ett resultat", job.Name)
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
