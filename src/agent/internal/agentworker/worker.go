// Package agentworker tar sig an de mål som är markerade för agentarbete.
//
// Arbetsflödet är avsiktligt litet: agenten specificerar målet, skriver
// den egna definitionen tillbaka i Plan och sätter målets status till "plan".
// Resten - planering, utförande, verifiering - är senare faser.
package agentworker

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	"plan-agent/internal/definer"
	"plan-agent/internal/planclient"
	"plan-agent/internal/profile"
)

// Faser och tillstånd speglar agent_runs-tabellen i Plan.
const (
	PhaseDefine    = "define"
	PhasePlan      = "plan"
	PhaseExecute   = "execute"
	PhaseVerify    = "verify"
	PhaseCompleted = "completed"

	StatusRunning = "running"
	StatusDone    = "done"
	StatusFailed  = "failed"
	// StatusBlocked betyder att körningen väntar på en människa.
	StatusBlocked = "blocked"
	// StatusQueued betyder att körningen väntar på att få ett Job.
	StatusQueued = "queued"
	// StatusCancelled betyder att en människa avbröt körningen.
	StatusCancelled = planclient.RunStatusCancelled
)

// ErrFinished betyder att målet är färdigt och workern kan avslutas.
var ErrFinished = errors.New("goal finished")

// TriggerAgent beskriver att körningen startades av agenten själv.
const TriggerAgent = "agent"

// Agentprofiler - vilken sorts agent som utför körningen.
const (
	ProfileCoding = profile.Coding
	ProfileLife   = profile.Life
)

// Worker pollar Plan efter agentmål och definierar dem.
type GoalDefiner interface {
	Define(ctx context.Context, goalTitle, existingDescription, feedback string) (*definer.Definition, error)
}

// CodeRunner driver kodändringar för ett mål.
type CodeRunner interface {
	Process(ctx context.Context, run planclient.AgentRun, goal planclient.Goal) error
}

type Worker struct {
	planClient   *planclient.Client
	definer      GoalDefiner
	codeRunner   CodeRunner
	pollInterval time.Duration
	// onlyGoalId begränsar workern till ett mål. Noll betyder alla.
	onlyGoalId int64
	// initiator beskriver vem eller vad som startade körningen.
	initiator string
	// profile är vilken sorts agent som kör - coding eller life.
	profile string
	// stopWhenFinished gör att workern slutar när målet är färdigt.
	stopWhenFinished bool
}

// Config konfigurerar workern
type Config struct {
	// PollInterval är hur ofta Plan läses. Noll ger en minut.
	PollInterval time.Duration
	// OnlyGoalId begränsar workern till ett enda mål - används för att
	// köra ett uppdrag manuellt.
	OnlyGoalId int64
	// Initiator beskriver vem eller vad som startade körningen, t.ex.
	// "worker:agent-worker-goal-502". Tom sträng ger en generell agent.
	Initiator string
	// Profile är vilken sorts agent som kör - coding eller life.
	// Tom sträng ger coding.
	Profile string
	// CodeRunner driver kodändringar. Nil betyder att workern bara
	// definierar mål och inte utför ändringar.
	CodeRunner CodeRunner
	// StopWhenFinished gör att workern slutar när målet är färdigt.
	//
	// Det används när workern körs i ett Kubernetes Job - jobbet ska avslutas
	// när uppgiften är klar, inte köra i evighet.
	StopWhenFinished bool
}

// New creates a worker that defines the goals marked for agent work
func New(planClient *planclient.Client, goalDefiner GoalDefiner, cfg Config) *Worker {
	interval := cfg.PollInterval
	if interval == 0 {
		interval = time.Minute
	}

	// Profilen bestämmer vad workern får göra - den är en teknisk
	// gräns, inte en instruktion till modellen.
	agentProfile, err := profile.PolicyFor(cfg.Profile)
	if err != nil {
		log.Printf("Invalid profile %q, using %s", cfg.Profile, profile.Coding)
		agentProfile = profile.CodingPolicy()
	}

	// En profil som inte får röra kod får inte köra codern.
	codeRunner := cfg.CodeRunner
	if !agentProfile.MayChangeCode {
		codeRunner = nil
	}

	return &Worker{
		planClient:       planClient,
		definer:          goalDefiner,
		codeRunner:       codeRunner,
		pollInterval:     interval,
		onlyGoalId:       cfg.OnlyGoalId,
		initiator:        cfg.Initiator,
		profile:          agentProfile.Name,
		stopWhenFinished: cfg.StopWhenFinished,
	}
}

// Run starts the worker (blocks until the context is cancelled)
func (w *Worker) Run(ctx context.Context) {
	log.Printf("Agent worker starting, polling every %s for goals with agent=true", w.pollInterval)

	ticker := time.NewTicker(w.pollInterval)
	defer ticker.Stop()

	if err := w.RunOnce(ctx); err != nil {
		if errors.Is(err, ErrFinished) {
			log.Println("Agent worker finished the goal and stops")
			return
		}
		log.Printf("Agent worker run failed: %v", err)
	}

	for {
		select {
		case <-ctx.Done():
			log.Println("Agent worker stopping")
			return
		case <-ticker.C:
			err := w.RunOnce(ctx)
			if err == nil {
				continue
			}
			if errors.Is(err, ErrFinished) {
				log.Println("Agent worker finished the goal and stops")
				return
			}
			log.Printf("Agent worker run failed: %v", err)
		}
	}
}

// RunOnce processes the agent goals once
func (w *Worker) RunOnce(ctx context.Context) error {
	goals, err := w.planClient.GetAgentGoals(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch agent goals: %w", err)
	}

	defineStatusID, err := w.planClient.GetStatusIDByName(ctx, PhaseDefine)
	if err != nil {
		return fmt.Errorf("failed to resolve define phase: %w", err)
	}

	runs, err := w.planClient.GetAgentRuns(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to fetch agent runs: %w", err)
	}

	// Körningar per mål - den senaste används.
	latestRun := make(map[int64]planclient.AgentRun, len(runs))
	runCount := make(map[int64]int, len(runs))
	for _, run := range runs {
		existing, found := latestRun[run.GoalID]
		if !found || run.ID > existing.ID {
			latestRun[run.GoalID] = run
		}
		runCount[run.GoalID]++
	}

	// En köad körning är en uttrycklig begäran om att få köras. Workern
	// tar över den i stället för att skapa en egen.
	queuedRun := findQueuedRun(runs)

	// Mål som redan behandlats - en köad körning får inte hanteras två
	// gånger.
	handled := make(map[int64]bool, len(goals))

	planStatusID, err := w.planClient.GetStatusIDByName(ctx, PhasePlan)
	if err != nil {
		return fmt.Errorf("failed to resolve plan phase: %w", err)
	}

	for _, goal := range goals {
		if !goal.Agent {
			continue
		}
		if w.onlyGoalId != 0 && goal.ID != w.onlyGoalId {
			continue
		}

		run, hasRun := latestRun[goal.ID]

		// En avbruten körning ska inte återupptas - en människa har sagt
		// stopp. Jobbet avslutas i stället för att börja om.
		if hasRun && run.Status == StatusCancelled {
			log.Printf("Run %d for goal #%d was cancelled - the worker stops", run.ID, goal.ID)
			return ErrFinished
		}

		// En köad körning är en uttrycklig begäran om att få köras. Den
		// tas över och arbetet utförs i den körningen.
		if queued, found := queuedRun[goal.ID]; found {
			handled[goal.ID] = true
			if err := w.runQueued(ctx, goal, queued, defineStatusID); err != nil {
				log.Printf("Failed to run queued run %d: %v", queued.ID, err)
			}
			continue
		}

		// Ett mål som väntar på ett mänskligt beslut ska återupptas så
		// snart beslutet är fattat.
		if hasRun && w.codeRunner != nil && run.Status == StatusBlocked {
			if err := w.codeRunner.Process(ctx, run, goal); err != nil {
				log.Printf("Failed to resume goal #%d: %v", goal.ID, err)
			}
			continue
		}

		// En pågående körning lämnas ifred.
		if hasRun && run.Status == StatusRunning {
			continue
		}

		// Ett mål i definitionfasen specificeras först.
		if goal.StatusID == defineStatusID {
			if err := w.processGoal(ctx, goal, hasRun, run); err != nil {
				log.Printf("Failed to process goal #%d: %v", goal.ID, err)
			}
			continue
		}

		// Ett definierat mål lämnas över till codern. Det sker bara en gång
		// per mål - efter definitionen finns exakt en körning, och när
		// kodkörningen finns är det två.
		if w.codeRunner != nil && goal.StatusID == planStatusID && runCount[goal.ID] < 2 {
			if err := w.startCodeRun(ctx, goal, false, planclient.AgentRun{}); err != nil {
				log.Printf("Failed to start code run for goal #%d: %v", goal.ID, err)
			}
			continue
		}

		// Ett färdigt mål avslutar workern - ett Kubernetes Job ska inte
		// köra i evighet.
		if w.stopWhenFinished && w.isFinished(goal, run, hasRun) {
			return ErrFinished
		}
	}

	// En köad körning är en uttrycklig begäran och ska inte glömmas bara
	// för att målet inte är ett agentmål. Utan det här skulle jobbet köra
	// i evighet på ett mål som aldrig behandlas.
	for goalID, queued := range queuedRun {
		if handled[goalID] {
			continue
		}
		if w.onlyGoalId != 0 && goalID != w.onlyGoalId {
			continue
		}

		goal, err := w.planClient.GetGoal(ctx, goalID)
		if err != nil {
			log.Printf("Failed to fetch goal #%d for queued run %d: %v", goalID, queued.ID, err)
			continue
		}

		if err := w.runQueued(ctx, *goal, queued, defineStatusID); err != nil {
			log.Printf("Failed to run queued run %d: %v", queued.ID, err)
		}
	}

	// Ett jobb finns för att utföra en körning. När körningen är klar är
	// jobbet klart - det ska inte stå och vänta på en fas som aldrig
	// kommer, som när en livsprofil lämnar målet i planfasen.
	if w.stopWhenFinished && w.onlyGoalId != 0 {
		done, err := w.goalRunDone(ctx, w.onlyGoalId)
		if err != nil {
			log.Printf("Failed to check whether goal #%d is finished: %v", w.onlyGoalId, err)
		} else if done {
			return ErrFinished
		}
	}

	return nil
}

// goalRunDone berättar om den senaste körningen för målet är klar.
func (w *Worker) goalRunDone(ctx context.Context, goalID int64) (bool, error) {
	runs, err := w.planClient.GetAgentRuns(ctx, nil)
	if err != nil {
		return false, fmt.Errorf("failed to fetch agent runs: %w", err)
	}

	var latest planclient.AgentRun
	found := false
	for _, run := range runs {
		if run.GoalID != goalID {
			continue
		}
		if !found || run.ID > latest.ID {
			latest = run
			found = true
		}
	}

	if !found {
		return false, nil
	}

	return latest.Status == StatusDone, nil
}

// runQueued tar över en köad körning och utför arbetet i den.
func (w *Worker) runQueued(ctx context.Context, goal planclient.Goal, queued planclient.AgentRun, defineStatusID int) error {
	if err := w.adoptRun(ctx, queued); err != nil {
		return err
	}

	if goal.StatusID == defineStatusID {
		return w.processGoal(ctx, goal, true, queued)
	}

	// En profil som inte får röra kod har ingen kodfas - körningen
	// avslutas utan ändringar.
	if w.codeRunner == nil {
		return w.closeRunWithoutChanges(ctx, goal, queued)
	}

	return w.startCodeRun(ctx, goal, true, queued)
}

// findQueuedRun ger de köade körningarna, per mål.
//
// En köad körning är en uttrycklig begäran om att få köras - jobbcontrollern
// skapar ett Kubernetes Job för den.
func findQueuedRun(runs []planclient.AgentRun) map[int64]planclient.AgentRun {
	queued := make(map[int64]planclient.AgentRun)
	for _, run := range runs {
		if run.Status != StatusQueued {
			continue
		}
		existing, found := queued[run.GoalID]
		if !found || run.ID > existing.ID {
			queued[run.GoalID] = run
		}
	}
	return queued
}

// adoptRun tar över en köad körning och markerar den som pågående.
func (w *Worker) adoptRun(ctx context.Context, run planclient.AgentRun) error {
	status := StatusRunning
	summary := "Körningen har startat"

	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:  &status,
		Summary: &summary,
	}); err != nil {
		return fmt.Errorf("failed to adopt run %d: %w", run.ID, err)
	}

	log.Printf("Adopted queued agent run %d for goal #%d", run.ID, run.GoalID)
	return nil
}

// isFinished berättar om målet är färdigt och ingen körning pågår.
func (w *Worker) isFinished(goal planclient.Goal, run planclient.AgentRun, hasRun bool) bool {
	if !hasRun {
		return false
	}
	// En misslyckad körning är inte färdig - jobbet ska inte avslutas
	// innan en människa eller en ny körning tagit ställning till felet.
	if run.Status != StatusDone {
		return false
	}

	// En profil som får röra kod flyttar målet till verify när ändringen
	// är verifierad. Då är jobbet klart.
	verifyStatusID, err := w.planClient.GetStatusIDByName(context.Background(), PhaseVerify)
	if err == nil && goal.StatusID == verifyStatusID {
		return true
	}

	// En profil som inte får röra kod lämnar målet i plan - men körningen
	// är klar, och jobbet ska inte stå och vänta på en fas som aldrig
	// kommer.
	return w.codeRunner == nil
}

// closeRunWithoutChanges avslutar en körning som inte får göra
// kodändringar.
//
// Profilen säger vad som är tillåtet - körningen ska inte lämnas öppen
// bara för att profilen förbjuder det steget.
func (w *Worker) closeRunWithoutChanges(ctx context.Context, goal planclient.Goal, run planclient.AgentRun) error {
	status := StatusDone
	summary := fmt.Sprintf("Profilen %s får inte ändra kod - körningen avslutades utan ändringar", w.profile)
	finishedAt := time.Now().UTC()

	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &summary,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to close agent run: %w", err)
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunStarted,
		Phase: run.Phase,
		Payload: map[string]any{
			"goal_id":       goal.ID,
			"trigger":       run.Trigger,
			"initiator":     w.initiator,
			"agent_profile": w.profile,
		},
	})

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:    planclient.EventRunCompleted,
		Payload: map[string]any{"reason": "profile does not allow code changes", "profile": w.profile},
	})

	return nil
}

// startCodeRun skapar en körning för ett definierat mål och lämnar över
// till codern.
func (w *Worker) startCodeRun(ctx context.Context, goal planclient.Goal, hasRun bool, existing planclient.AgentRun) error {
	run := &existing
	if !hasRun {
		created, err := w.planClient.CreateAgentRun(ctx, planclient.CreateAgentRunRequest{
			GoalID:       goal.ID,
			Phase:        PhasePlan,
			Status:       StatusRunning,
			Trigger:      TriggerAgent,
			Initiator:    w.initiator,
			AgentProfile: w.profile,
			Summary:      "Agenten tar fram ett kodförslag",
		})
		if err != nil {
			return fmt.Errorf("failed to create agent run: %w", err)
		}
		run = created
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunStarted,
		Phase: PhasePlan,
		Payload: map[string]any{
			"goal_id":       goal.ID,
			"trigger":       TriggerAgent,
			"initiator":     w.initiator,
			"agent_profile": w.profile,
		},
	})

	if err := w.codeRunner.Process(ctx, *run, goal); err != nil {
		return err
	}
	return nil
}

// processGoal defines one goal and writes the result back to Plan
//
// Om det redan finns en körning för målet - en köad körning som workern
// tagit över - används den i stället för att skapa en ny.
func (w *Worker) processGoal(ctx context.Context, goal planclient.Goal, hasRun bool, existing planclient.AgentRun) error {
	log.Printf("Agent worker processing goal #%d: %s", goal.ID, goal.Title)

	run := &existing
	if !hasRun {
		created, err := w.planClient.CreateAgentRun(ctx, planclient.CreateAgentRunRequest{
			GoalID:       goal.ID,
			Phase:        PhaseDefine,
			Status:       StatusRunning,
			Trigger:      TriggerAgent,
			Initiator:    w.initiator,
			AgentProfile: w.profile,
			Summary:      "Agenten specificerar målet",
		})
		if err != nil {
			return fmt.Errorf("failed to create agent run: %w", err)
		}
		run = created
	}
	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunStarted,
		Phase: PhaseDefine,
		Payload: map[string]any{
			"goal_id":       goal.ID,
			"trigger":       TriggerAgent,
			"initiator":     w.initiator,
			"agent_profile": w.profile,
		},
	})

	definition, err := w.definer.Define(ctx, goal.Title, goal.Description, "")
	if err != nil {
		w.failRun(ctx, run.ID, err)
		return fmt.Errorf("failed to define goal: %w", err)
	}
	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventDefinitionCreated,
		Phase: PhaseDefine,
	})

	title := definition.Title
	description := definition.Description
	if _, err := w.planClient.UpdateGoal(ctx, goal.ID, planclient.UpdateGoalRequest{
		Title:       &title,
		Description: &description,
	}); err != nil {
		w.failRun(ctx, run.ID, err)
		return fmt.Errorf("failed to update goal: %w", err)
	}
	if err := w.syncPhase(ctx, goal.ID, run.ID, PhasePlan); err != nil {
		w.failRun(ctx, run.ID, err)
		return fmt.Errorf("failed to synchronize plan phase: %w", err)
	}

	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventPhaseChanged,
		Phase: PhasePlan,
		Payload: map[string]any{
			"status": StatusDone,
		},
	})

	// Med en coder igång fortsätter samma körning - den ändrar bara kod
	// efter att en människa godkänt förslaget.
	if w.codeRunner != nil {
		if err := w.codeRunner.Process(ctx, *run, goal); err != nil {
			return fmt.Errorf("failed to hand over to the coder: %w", err)
		}
		log.Printf("Goal #%d defined and handed over to the coder", goal.ID)
		return nil
	}

	status := StatusDone
	summary := definition.Description
	result := "Målet definierat och flyttat till plan"
	finishedAt := time.Now().UTC()
	if _, err := w.planClient.UpdateAgentRun(ctx, run.ID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		return fmt.Errorf("failed to update agent run: %w", err)
	}

	// En avslutad körning ska alltid sluta med run_completed - annars ser
	// historiken ut som om körningen bara upphörde.
	w.recordEvent(ctx, run.ID, planclient.CreateAgentRunEventRequest{
		Type:    planclient.EventRunCompleted,
		Phase:   PhasePlan,
		Payload: map[string]any{"reason": "goal defined and moved to planning"},
	})

	log.Printf("Goal #%d defined and moved to planning", goal.ID)
	return nil
}

func (w *Worker) recordEvent(ctx context.Context, runID int64, event planclient.CreateAgentRunEventRequest) {
	if err := w.planClient.CreateAgentRunEvent(ctx, runID, event); err != nil {
		log.Printf("Failed to record event for agent run %d: %v", runID, err)
	}
}

func (w *Worker) syncPhase(ctx context.Context, goalID, runID int64, phase string) error {
	if err := w.planClient.UpdateGoalPhase(ctx, goalID, phase); err != nil {
		return fmt.Errorf("failed to update goal phase: %w", err)
	}
	if _, err := w.planClient.UpdateAgentRun(ctx, runID, planclient.UpdateAgentRunRequest{
		Phase: &phase,
	}); err != nil {
		return fmt.Errorf("failed to update agent run phase: %w", err)
	}
	return nil
}

// failRun marks a run as failed so the goal does not get stuck
func (w *Worker) failRun(ctx context.Context, runID int64, cause error) {
	status := StatusFailed
	summary := "Agenten kunde inte specificera målet: " + cause.Error()
	result := "Misslyckades: " + cause.Error()
	finishedAt := time.Now().UTC()
	if _, err := w.planClient.UpdateAgentRun(ctx, runID, planclient.UpdateAgentRunRequest{
		Status:     &status,
		Summary:    &summary,
		Result:     &result,
		FinishedAt: &finishedAt,
	}); err != nil {
		log.Printf("Failed to mark agent run %d as failed: %v", runID, err)
	}
	w.recordEvent(ctx, runID, planclient.CreateAgentRunEventRequest{
		Type:  planclient.EventRunFailed,
		Phase: PhaseDefine,
		Payload: map[string]any{
			"error": cause.Error(),
		},
	})
}

// PollIntervalFromEnv reads the poll interval from the environment
func PollIntervalFromEnv() time.Duration {
	raw := os.Getenv("AGENT_POLL_INTERVAL")
	if raw == "" {
		return time.Minute
	}

	interval, err := time.ParseDuration(raw)
	if err != nil {
		log.Printf("Invalid AGENT_POLL_INTERVAL %q, using one minute", raw)
		return time.Minute
	}
	return interval
}
